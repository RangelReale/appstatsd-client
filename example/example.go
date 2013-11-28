package main

import (
	"fmt"
	"github.com/RangelReale/appstatsd-client"
	"time"
)

func main() {
	client := appstatsdclient.NewLocal("apdc-test")
	defer client.Close()

	for {
		client.Log(appstatsdclient.WARNING, fmt.Sprintf("First warning"))
		client.Increment("conn.proj#1.ct")
		time.Sleep(3 * time.Second)
		client.Timing("conn.dr", 1250)
		client.Increment("proc#chkstatus#t1.proj#1.ct")
		time.Sleep(1 * time.Second)
		client.Log(appstatsdclient.ERROR, fmt.Sprintf("An errror"))
	}
}
