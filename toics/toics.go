package toics

import (
	"fmt"
	"mail2ics/clean"
	"os"
)

func ToIcs(msg *clean.Message) error {
	var eventsStr string

	combineEvent(&msg.Events, &eventsStr)
	calendar := fmt.Sprintf(
		"BEGIN:VCALENDAR\n"+
			"PRODID:%s\n"+
			"VERSION:2.0\n"+
			"CALSCALE:GREGORIAN\n"+
			"METHOD:PUBLISH\n"+
			"X-WR-CALNAME:%s\n"+
			"X-WR-TIMEZONE:Asia/Shanghai\n"+
			"%s"+
			"END:VCALENDAR", msg.From, msg.Cal, eventsStr)

	if err := toFile(msg.Filename, calendar); err != nil {
		return err
	}

	return nil
}

func combineEvent(events *[]clean.Event, eventsStr *string) {
	for _, e := range *events {
		if e.Uid == "" {
			continue
		}

		event := fmt.Sprintf(
			"BEGIN:VEVENT\n"+
				"DTSTART%s\n"+
				"DTEND%s\n"+
				"UID:%s\n"+
				"SUMMARY:%s\n"+
				"DESCRIPTION:%s\n"+
				"LOCATION:%s\n"+
				"END:VEVENT\n", e.StartDT, e.EndDT, e.Uid, e.Summary, e.Detail, e.Location)

		*eventsStr += event
	}
}

func toFile(filename string, data string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}
