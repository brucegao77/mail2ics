package recive

import (
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"io"
	"io/ioutil"
	"log"
	"mail2ics/config"
)

type Mail struct {
	From    string
	Subject string
	Content string
}

func CheckMail(cc *chan Mail) {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS(config.Bruce.Addr, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(config.Bruce.Username, config.Bruce.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	// Get the lastest messages
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(mbox.Messages-10, mbox.Messages)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	go func() {
		if err := c.Fetch(seqSet, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	for msg := range messages {
		r := msg.GetBody(&section)
		if r == nil {
			log.Fatal("Server didn't returned message body")
		}
		// Create a new mail reader
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Fatal(err)
		}

		// Print some info about the message
		header := mr.Header
		if from, err := header.AddressList("From"); err == nil && from[0].Address == "brucegxs@gmail.com" {
			subject, err := header.Subject()
			if err != nil {
				log.Fatal(err)
			}
			// Process each message's part
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					log.Fatal(err)
				}

				// This is the message's text (can be plain-text or HTML)
				b, _ := ioutil.ReadAll(p.Body)
				m := Mail{From: from[0].Address, Subject: subject, Content: string(b)}
				*cc <- m
			}
		}
	}

	log.Println("Done!")
}