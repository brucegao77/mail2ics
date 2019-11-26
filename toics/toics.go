package toics

import (
	"fmt"
	"mail2ics/clean"
	"os"
)

func ToIcs(msg *clean.Message) error {
	var eventsStr string

	combineEvent(&msg.Events, eventsStr)
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

func combineEvent(events *[]clean.Event, eventsStr string) {
	for _, e := range *events {
		event := fmt.Sprintf(
			"BEGIN:VEVENT\n"+
				"DTSTART:%s\n"+
				"UID:%s"+
				"SUMMARY:%s\n"+
				"DESCRIPTION:%s\n"+
				"LOCATION:%s\n"+
				"BEGIN:VALARM"+
				"ACTION:DISPLAY"+
				"DESCRIPTION:This is an event reminder"+
				"TRIGGER:-P0DT1H0M0S"+
				"END:VALARM"+
				"END:VEVENT\n", e.StartDT, e.Uid, e.Summary, e.Detail, e.Location)

		eventsStr += event
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
