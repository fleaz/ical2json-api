package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/imroc/req/v3"
	"github.com/teambition/rrule-go"
)

type Event struct {
	Name        string    `json:"name"`
	DateStart   time.Time `json:"dateStart"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
}

func getEventsInPeriod(start time.Time, end time.Time) []Event {
	calURL := "https://davical.darmstadt.ccc.de/public.php/cda/public/"
	resp, _ := req.Get(calURL)

	var events []Event

	cal, _ := ics.ParseCalendar(resp.Body)

	for _, e := range cal.Events() {
		timeStart, _ := e.GetStartAt()
		name := e.GetProperty(ics.ComponentPropertySummary).Value
		descriptionProp := e.GetProperty(ics.ComponentPropertyDescription)
		locationProp := e.GetProperty(ics.ComponentPropertyLocation)

		description := ""
		location := ""

		if descriptionProp != nil {
			description = descriptionProp.BaseProperty.Value
		}
		if locationProp != nil {
			location = locationProp.BaseProperty.Value
		}

		// Check if it's a single event in our timeperiod
		if start.Before(timeStart) && end.After(timeStart) {
			e := Event{
				Name:        name,
				DateStart:   timeStart,
				Description: description,
				Location:    location,
			}
			events = append(events, e)
		}

		// Get possible RRule from iCal entry
		icalRule := e.GetProperty(ics.ComponentPropertyRrule)
		if icalRule != nil {
			set, err := rrule.StrToRRuleSet(fmt.Sprintf("RRULE:%s", icalRule.BaseProperty.Value))
			if err != nil {
				fmt.Println(err)
				break
			}

			// Use the original start date of the event as the start for the repeat rule
			set.DTStart(timeStart)

			// Get timestamps of all events repeats in this timeframe
			// Array will be empty if there are none
			repeatsInPeriod := set.Between(start, end, false)
			for _, ts := range repeatsInPeriod {
				e := Event{
					Name:        name,
					DateStart:   ts,
					Description: description,
					Location:    location,
				}
				events = append(events, e)
			}
		}

	}
	// Sort by startDate
	sort.Slice(events, func(i, j int) bool {
		return events[i].DateStart.Before(events[j].DateStart)
	})

	return events
}

func Handler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	end := now.AddDate(0, 0, 14)
	events := getEventsInPeriod(now, end)
	b, _ := json.Marshal(events)
	io.WriteString(w, string(b))
}

func main() {
	http.HandleFunc("/", Handler)
	http.ListenAndServe(":5078", nil)
}
