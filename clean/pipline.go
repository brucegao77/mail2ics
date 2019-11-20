package clean

import (
	"log"
	"mail2ics/recive"
	"time"
)

type Message struct {
	From     string
	Subject  string
	StartDT  string
	EndDT    string
	Location string
	Detail   string
}

func (msg *Message) preClean(m *recive.Mail) {
	msg.From = m.From
	msg.Subject = m.Subject
}

func (msg *Message) railWay(m *recive.Mail) string {
	msg.StartDT =
}

func Pipline(m *recive.Mail, msg *Message) {
	msg.preClean(m)

	switch msg.Subject {
	case "网上购票系统--用户支付通知":
		msg.railWay(m)
	default:

	}
}

func parseTime(t string) string {
	form := "2006年01月02日15:04"
	t1, err := time.Parse(form, t)
	if err != nil {
		log.Fatal(err)
	}
	return t1.Format("20060102T150405")
}