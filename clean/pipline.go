package clean

import (
	"errors"
	"fmt"
	"mail2ics/recive"
	"math/rand"
	"regexp"
	"strings"
	"time"
)

type Message struct {
	From     string
	Subject  string
	Events   []Event
	Filename string
	Cal      string
}

type Event struct {
	StartDT  string
	EndDT    string
	Summary  string
	Location string
	Detail   string
	Uid      string
}

func (msg *Message) preClean(m *recive.Mail) {
	events := make([]Event, 1)
	msg.Events = events

	msg.From = m.From
	msg.Subject = m.Subject
	msg.Filename = "activity.ics"
	msg.Cal = "行程"
	msg.Events[0].Summary = m.Subject
}

const ICS_DT = "20060102T150405"

func (msg *Message) railWay(m *recive.Mail, mc *chan Message) error {
	// Maybe multiple ticket information in one mail
	count := 1
	for {
		// main info
		ir, _ := regexp.Compile(
			fmt.Sprintf(`%d.(.*?)，(.*?)开，(.*?)，(.*?),(.*?)，(.*?)，票价(.*?)元`, count))
		info := ir.FindStringSubmatch(m.Content)
		// order number
		nr, _ := regexp.Compile(`订单号码(.*?)。`)
		num := nr.FindStringSubmatch(m.Content)[1]
		// check point
		cr, _ := regexp.Compile(`检票口：(.*?)。`)
		result := cr.FindStringSubmatch(m.Content)
		var check string
		if len(result) < 2 {
			check = "请现场查看"
		} else {
			check = result[1]
		}

		if count == 1 && len(info) < 2 {
			return errors.New("can't find information, maybe the format of body has changed")
		}
		// No more ticket information
		if count > 1 && len(info) < 2 {
			break
		}

		if info[1] != m.User {
			continue
		}

		msg.Subject = fmt.Sprintf("列车行程：%s", info[3])
		// start time and end time
		if st, err := ParseTime(info[2], "2006年01月02日15:04", 8, "-"); err != nil {
			return err
		} else {
			if et, err := ParseTime(st, ICS_DT, 1, "+"); err != nil {
				return err
			} else {
				msg.Events[0].EndDT = fmt.Sprintf(":%s", et)
			}

			msg.Events[0].StartDT = fmt.Sprintf(":%s", st)
		}

		// TODO: Find specific address automatically
		msg.Events[0].Location = strings.Split(info[3], "-")[0]
		msg.Events[0].Detail = fmt.Sprintf(
			"订单号：%s\\n"+
				"乘客：%s\\n"+
				"车次：%s\\n"+
				"检票口：%s\\n"+
				"座位号：%s\\n"+
				"席别：%s\\n"+
				"票价：%s", strings.ReplaceAll(num, " ", ""),
			info[1], info[4], check, info[5], info[6], info[7])
		msg.Events[0].Uid = fmt.Sprintf("XC"+"%d", time.Now().Unix()+int64(rand.Intn(999999)))
		*mc <- *msg
		count++
	}

	return nil
}

func Pipeline(m *recive.Mail, msg *Message, mc *chan Message) error {
	msg.preClean(m)

	switch msg.Subject {
	case "Fwd: 网上购票系统--用户支付通知":
		if err := msg.railWay(m, mc); err != nil {
			return err
		}
	default:

	}

	return nil
}

func ParseTime(t string, form string, hourAdd int, method string) (string, error) {
	t1, err := time.Parse(form, t)
	if err != nil {
		return "", err
	}
	// Don't know why, when add to google calendar, time will +8 hours
	// so -8 hours here
	h, _ := time.ParseDuration(fmt.Sprintf("%s1h", method))
	t2 := t1.Add(time.Duration(hourAdd) * h)

	return t2.Format(ICS_DT), err
}
