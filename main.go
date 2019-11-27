package main

import (
	"fmt"
	"log"
	"mail2ics/clean"
	"mail2ics/recive"
	"mail2ics/send"
	"mail2ics/task"
	"time"
)

func main() {
	messageChannel := make(chan clean.Message, 10)

	go mail(&messageChannel)
	go movie(&messageChannel)

	// Send email
	for m := range messageChannel {
		if err := send.SendEmail(&m); err != nil {
			log.Fatal(err)
		}

		log.Println(fmt.Sprintf("An activity has send to %s!", m.From))
	}
}

func mail(messageChannel *chan clean.Message) {
	contentChannel := make(chan recive.Mail, 10)

	// Check mail every hour
	go func() {
		for {
			if err := recive.CheckMail(&contentChannel); err != nil {
				log.Fatal(err)
			}

			time.Sleep(time.Hour)
		}
	}()

	// Handle the mail body
	for c := range contentChannel {
		log.Println("Handling: ", c.Subject)
		var msg clean.Message

		if err := clean.Pipeline(&c, &msg, messageChannel); err != nil {
			log.Fatal(err)
		}
	}
}

func movie(mc *chan clean.Message) {
	for {
		if err := task.MovieSchedule(mc); err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Hour * 24 * 7)
	}
}
