package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/emersion/go-ical"
	"github.com/imroc/req/v3"
)

type Event struct {
	Name string
	Date time.Time
}

func main() {
	calURL := "https://davical.darmstadt.ccc.de/public.php/cda/public/"
	resp, _ := req.Get(calURL)

	var events []Event

	dec := ical.NewDecoder(resp.Body)
	for {
		cal, err := dec.Decode()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		for _, event := range cal.Events() {
			loc, _ := time.LoadLocation("Europe/Berlin")
			t, _ := event.DateTimeStart(loc)
			summary, _ := event.Props.Text(ical.PropSummary)
			events = append(events, Event{Name: summary, Date: t})
		}
	}

	b, _ := json.Marshal(events)
	fmt.Println(string(b))

}
