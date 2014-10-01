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
type Client struct {
	App          string
	StatsDParams ClientParams
	LogParams    ClientParams
}

type ClientParams struct {
	Host string
	Port int
	conn *net.UDPConn
}

// Factory method to initialize udp connection
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
func New(app string, statsdhost string, statsdport int, loghost string, logport int) *Client {
	if strings.Contains(app, ".") {
		panic("appstatsd app name cannot containt dots.")
	}

	client := Client{
		App:          app,
		StatsDParams: ClientParams{Host: statsdhost, Port: statsdport},
		LogParams:    ClientParams{Host: loghost, Port: logport},
	}
	client.Open()
	return &client
}

func NewLocal(app string) *Client {
	return New(app, "localhost", 8125, "localhost", 8126)
}

// Method to open udp connection, called by default client factory
func (client *Client) Open() {
	statsdConnectionString := fmt.Sprintf("%s:%d", client.StatsDParams.Host, client.StatsDParams.Port)
	statsdaddr, err := net.ResolveUDPAddr("udp", statsdConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	statsdconn, err := net.DialUDP("udp", nil, statsdaddr)
	if err != nil {
		log.Println(err)
	}
	client.StatsDParams.conn = statsdconn

	logConnectionString := fmt.Sprintf("%s:%d", client.LogParams.Host, client.LogParams.Port)
	logaddr, err := net.ResolveUDPAddr("udp", logConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	logconn, err := net.DialUDP("udp", nil, logaddr)
	if err != nil {
		log.Println(err)
	}
	client.LogParams.conn = logconn
}

// Method to close udp connection
func (client *Client) Close() {
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
func (client *Client) Timing(stat string, time int64) {
	updateString := fmt.Sprintf("%d|ms", time)
	stats := map[string]string{stat: updateString}
	client.SendStats(stats, 1)
}

func (client *Client) TimingDuration(stat string, t time.Duration) {
	client.Timing(stat, int64(t/time.Millisecond))
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
func (client *Client) TimingWithSampleRate(stat string, time int64, sampleRate float32) {
	updateString := fmt.Sprintf("%d|ms", time)
	stats := map[string]string{stat: updateString}
	client.SendStats(stats, sampleRate)
}

func (client *Client) TimingWithSampleRateDuration(stat string, t time.Duration, sampleRate float32) {
	client.TimingWithSampleRate(stat, int64(t/time.Millisecond), sampleRate)
}

// Increments one stat counter without sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.Increment('foo.bar', 1)
func (client *Client) Increment(stat string, amount int) {
	stats := []string{stat}
	client.UpdateStats(stats, amount, 1)
}

// Increments one stat counter with sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.Increment('foo.bar', 0.2)
func (client *Client) IncrementWithSampling(stat string, amount int, sampleRate float32) {
	stats := []string{stat}
	client.UpdateStats(stats[:], amount, sampleRate)
}

// Decrements one stat counter without sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.Decrement('foo.bar', 1)
func (client *Client) Decrement(stat string, amount int) {
	stats := []string{stat}
	client.UpdateStats(stats[:], -1*amount, 1)
}

// Decrements one stat counter with sampling
//
// Usage:
//
//     import "statsd"
//     client := statsd.New('localhost', 8125)
//     client.Decrement('foo.bar', 0.2)
func (client *Client) DecrementWithSampling(stat string, amount int, sampleRate float32) {
	stats := []string{stat}
	client.UpdateStats(stats[:], -1*amount, sampleRate)
}

// Arbitrarily updates a list of stats by a delta
func (client *Client) UpdateStats(stats []string, delta int, sampleRate float32) {
	statsToSend := make(map[string]string)
	for _, stat := range stats {
		updateString := fmt.Sprintf("%d|c", delta)
		statsToSend[stat] = updateString
	}
	client.SendStats(statsToSend, sampleRate)
}

// Sends data to udp statsd daemon
func (client *Client) SendStats(data map[string]string, sampleRate float32) {
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
func (client *Client) Log(level LogLevel, message string) {
	client.LogId(level, "", message)
}

// Sends data to udp statsd daemon
func (client *Client) LogId(level LogLevel, messageid string, message string) {
	update_string := fmt.Sprintf("%s:%d:%s:%s", client.App, level, messageid, message)
	_, err := fmt.Fprintf(client.LogParams.conn, update_string)
	if err != nil {
		log.Println(err)
	}
}
