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
		time.Sleep(3 * time.Second)
	}
}
