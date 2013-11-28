// Mostly copied from https://github.com/etsy/statsd/blob/master/examples/go/statsd.go

package appstatsdclient

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

type LogLevel int

const (
	CRITICAL LogLevel = 1
	ERROR             = 2
	WARNING           = 3
	NOTICE            = 4
	INFO              = 5
	DEBUG             = 6
)

// The StatsdClient type defines the relevant properties of a StatsD connection.
type AppStatsdClient struct {
	App          string
	StatsDParams AppStatsdClientParams
	LogParams    AppStatsdClientParams
}

type AppStatsdClientParams struct {
	Host string
	Port int
	conn net.Conn
}

// Factory method to initialize udp connection
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
func New(app string, statsdhost string, statsdport int, loghost string, logport int) *AppStatsdClient {
	if strings.Contains(app, ".") {
		panic("appstatsd app name cannot containt dots.")
	}

	client := AppStatsdClient{
		App:          app,
		StatsDParams: AppStatsdClientParams{Host: statsdhost, Port: statsdport},
		LogParams:    AppStatsdClientParams{Host: loghost, Port: logport},
	}
	client.Open()
	return &client
}

func NewLocal(app string) *AppStatsdClient {
	return New(app, "localhost", 8125, "localhost", 8126)
}

// Method to open udp connection, called by default client factory
func (client *AppStatsdClient) Open() {
	statsdConnectionString := fmt.Sprintf("%s:%d", client.StatsDParams.Host, client.StatsDParams.Port)
	statsdconn, err := net.Dial("udp", statsdConnectionString)
	if err != nil {
		log.Println(err)
	}
	client.StatsDParams.conn = statsdconn

	logConnectionString := fmt.Sprintf("%s:%d", client.LogParams.Host, client.LogParams.Port)
	logconn, err := net.Dial("udp", logConnectionString)
	if err != nil {
		log.Println(err)
	}
	client.LogParams.conn = logconn
}

// Method to close udp connection
func (client *AppStatsdClient) Close() {
	client.StatsDParams.conn.Close()
	client.LogParams.conn.Close()
}

// Log timing information (in milliseconds) without sampling
//
// Usage:
//
//     import (
//         "statsd"
//         "time"
//     )
//
//     client := statsd.New('localhost', 8125)
//     t1 := time.Now()
//     expensiveCall()
//     t2 := time.Now()
//     duration := int64(t2.Sub(t1)/time.Millisecond)
//     client.Timing("foo.time", duration)
func (client *AppStatsdClient) Timing(stat string, time int64) {
	updateString := fmt.Sprintf("%d|ms", time)
	stats := map[string]string{stat: updateString}
	client.SendStats(stats, 1)
}

// Log timing information (in milliseconds) with sampling
//
// Usage:
//
//     import (
//         "statsd"
//         "time"
//     )
//
//     client := statsd.New('localhost', 8125)
//     t1 := time.Now()
//     expensiveCall()
//     t2 := time.Now()
//     duration := int64(t2.Sub(t1)/time.Millisecond)
//     client.TimingWithSampleRate("foo.time", duration, 0.2)
func (client *AppStatsdClient) TimingWithSampleRate(stat string, time int64, sampleRate float32) {
	updateString := fmt.Sprintf("%d|ms", time)
	stats := map[string]string{stat: updateString}
	client.SendStats(stats, sampleRate)
}

// Increments one stat counter without sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.Increment('foo.bar')
func (client *AppStatsdClient) Increment(stat string) {
	stats := []string{stat}
	client.UpdateStats(stats, 1, 1)
}

// Increments one stat counter with sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.Increment('foo.bar', 0.2)
func (client *AppStatsdClient) IncrementWithSampling(stat string, sampleRate float32) {
	stats := []string{stat}
	client.UpdateStats(stats[:], 1, sampleRate)
}

// Decrements one stat counter without sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.Decrement('foo.bar')
func (client *AppStatsdClient) Decrement(stat string) {
	stats := []string{stat}
	client.UpdateStats(stats[:], -1, 1)
}

// Decrements one stat counter with sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.Decrement('foo.bar', 0.2)
func (client *AppStatsdClient) DecrementWithSampling(stat string, sampleRate float32) {
	stats := []string{stat}
	client.UpdateStats(stats[:], -1, sampleRate)
}

// Arbitrarily updates a list of stats by a delta
func (client *AppStatsdClient) UpdateStats(stats []string, delta int, sampleRate float32) {
	statsToSend := make(map[string]string)
	for _, stat := range stats {
		updateString := fmt.Sprintf("%d|c", delta)
		statsToSend[stat] = updateString
	}
	client.SendStats(statsToSend, sampleRate)
}

// Sends data to udp statsd daemon
func (client *AppStatsdClient) SendStats(data map[string]string, sampleRate float32) {
	sampledData := make(map[string]string)
	if sampleRate < 1 {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		rNum := r.Float32()
		if rNum <= sampleRate {
			for stat, value := range data {
				sampledUpdateString := fmt.Sprintf("%s|@%f", value, sampleRate)
				sampledData[stat] = sampledUpdateString
			}
		}
	} else {
		sampledData = data
	}

	for k, v := range sampledData {
		update_string := fmt.Sprintf("%s.%s:%s", client.App, k, v)
		_, err := fmt.Fprintf(client.StatsDParams.conn, update_string)
		if err != nil {
			log.Println(err)
		}
	}
}

// Sends data to udp statsd daemon
func (client *AppStatsdClient) Log(level LogLevel, message string) {
	client.LogId(level, "", message)
}

// Sends data to udp statsd daemon
func (client *AppStatsdClient) LogId(level LogLevel, messageid string, message string) {
	update_string := fmt.Sprintf("%s:%d:%s:%s", client.App, level, messageid, message)
	_, err := fmt.Fprintf(client.LogParams.conn, update_string)
	if err != nil {
		log.Println(err)
	}
}
