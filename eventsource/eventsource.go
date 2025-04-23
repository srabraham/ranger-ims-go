package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/launchdarkly/eventsource"
)

type TimeEvent time.Time

func (t TimeEvent) Id() string    { return fmt.Sprint(time.Time(t).UnixNano()) }
func (t TimeEvent) Event() string { return "Tick" }
func (t TimeEvent) Data() string  { return time.Time(t).String() }

const (
	TICK_COUNT = 5
)

func TimePublisher(srv *eventsource.Server) {
	start := time.Date(2013, time.January, 1, 0, 0, 0, 0, time.UTC)
	ticker := time.NewTicker(time.Second)
	for i := 0; i < TICK_COUNT; i++ {
		<-ticker.C
		srv.Publish([]string{"time"}, TimeEvent(start))
		start = start.Add(time.Second)
	}
}

func main() {
	srv := eventsource.NewServer()
	srv.Gzip = true
	defer srv.Close()
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		return
	}
	defer l.Close()
	http.HandleFunc("/time", srv.Handler("time"))
	go http.Serve(l, nil)
	go TimePublisher(srv)
	stream, err := eventsource.Subscribe("http://localhost:8080/time", "")
	if err != nil {
		return
	}
	for i := 0; i < TICK_COUNT; i++ {
		ev := <-stream.Events
		fmt.Println(ev.Id(), ev.Event(), ev.Data())
	}

}
