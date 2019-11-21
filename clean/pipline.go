package clean

import (
	"errors"
	"fmt"
	"mail2ics/recive"
	"regexp"
	"strings"
	"time"
)

type Message struct {
	From     string
	Subject  string
	StartDT  string
	Location string
	Detail   string
}

func (msg *Message) preClean(m *recive.Mail) {
	msg.From = m.From
	msg.Subject = m.Subject
}

func (msg *Message) railWay(m *recive.Mail, mc *chan Message) error {
	// Maybe multiple ticket informations in one mail
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
			return errors.New("can't find informations, maybe the formate of body has changed!")
		}
		// No more ticket informations
		if count > 1 && len(info) < 2 {
			break
		}

		if info[1] != m.User {
			continue
		}

		msg.Subject = fmt.Sprintf("列车行程：%s", info[3])
		if t, err := parseTime(info[2], "2006年01月02日15:04"); err != nil {
			msg.StartDT = t
		}
		msg.Location = strings.Split(info[3], "-")[0]
		msg.Detail = fmt.Sprintf(
			"订单号：%s\n"+
				"乘客：%s\n"+
				"车次：%s\n"+
				"检票口：%s\n"+
				"座位号：%s\n"+
				"席别：%s\n"+
				"票价：%s", strings.ReplaceAll(num, " ", ""),
			info[1], info[4], check, info[5], info[6], info[7])
		*mc <- *msg
		count++
	}

	return nil
}

func Pipline(m *recive.Mail, msg *Message, mc *chan Message) error {
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

func parseTime(t string, form string) (string, error) {
	t1, err := time.Parse(form, t)
	if err != nil {
		return "", err
	}

	return t1.Format("20060102T150405"), err
}
