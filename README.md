appstatsd-client
================

Golang client for [appstatsd].

Heavily based on  https://github.com/etsy/statsd/blob/master/examples/go/statsd.go.

````go
import "github.com/RangelReale/appstatsd-client"

client := appstatsdclient.NewLocal("apdc-test")
defer client.Close()

for {
	client.Log(appstatsdclient.WARNING, fmt.Sprintf("First warning"))
	client.Increment("conn.proj#1.ct", 1)
	time.Sleep(3 * time.Second)
	client.Timing("conn.dr", 1250)
	client.Increment("proc#chkstatus#t1.proj#1.ct", 1)
	time.Sleep(1 * time.Second)
	client.Log(appstatsdclient.ERROR, fmt.Sprintf("An errror"))
}
````

Author
------

Rangel Reale


[appstatsd]: https://github.com/RangelReale/appstatsd

