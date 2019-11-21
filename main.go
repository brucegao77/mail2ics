package main

import (
	"fmt"
	"log"
	"mail2ics/clean"
	"mail2ics/recive"
	"mail2ics/send"
)

func main() {
	ContentChannel := make(chan recive.Mail, 10)
	MessageChannel := make(chan clean.Message, 10)

	// Check mail every minute
	go func() {
		for {
			if err := recive.CheckMail(&ContentChannel); err != nil {
				log.Fatal(err)
			}
			//time.Sleep(time.Minute)
			break
		}
	}()

	// Handle the mail body
	go func() {
		for c := range ContentChannel {
			log.Println("Handling: ", c.Subject)
			var msg clean.Message

			if err := clean.Pipline(&c, &msg, &MessageChannel); err != nil {
				log.Fatal(err)
			}

			log.Println("Message handled!")
		}
	}()

	// Send email
	for m := range MessageChannel {
		if err := send.SendEmail(&m); err != nil {
			log.Fatal(err)
		}
		log.Println(fmt.Sprintf("An activity has send to %s!", m.From))
	}
}
