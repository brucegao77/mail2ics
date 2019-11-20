package main

import (
	"mail2ics/clean"
	"mail2ics/recive"
)

func main() {
	for {
		ContentChannle := make(chan recive.Mail, 10)
		go recive.CheckMail(&ContentChannle)

		for m := range ContentChannle {
			var msg clean.Message

			clean.Pipline(&m, &msg)
			//send.SendEmail(&msg)
		}
	}
}
