package main

import (
	"log"
	"mail2ics/clean"
	"mail2ics/recive"
	"time"
)

func main() {
	ContentChannle := make(chan recive.Mail, 10)

	// Check mail every minute
	go func() {
		for {
			err := recive.CheckMail(&ContentChannle)
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Minute)
		}
	}()

	// Handle the mail body
	for m := range ContentChannle {
		log.Println("Handling mail: ", m.Subject)
		var msg clean.Message

		clean.Pipline(&m, &msg)
		//send.SendEmail(&msg)

		log.Println("Waiting for email")
	}
}
