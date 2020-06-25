package main

import (
	"io"
	"regexp"

	"log"

	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"

	"github.com/emersion/go-imap"
	"io/ioutil"
)

func fetchCodeFromMailbox(user, password string) string {
	code := ""
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(user, password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	// Get the last message
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 10 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 10
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(from, to)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 10)
	go func() {
		if err := c.Fetch(seqSet, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	for msg := range messages {
		if msg == nil {
			log.Fatal("Server didn't returned message")
		}

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
		if date, err := header.Date(); err == nil {
			log.Println("Date:", date)
		}
		if from, err := header.AddressList("From"); err == nil {
			log.Println("From:", from)
		}
		if to, err := header.AddressList("To"); err == nil {
			log.Println("To:", to)
		}
		if subject, err := header.Subject(); err == nil {
			log.Println("Subject:", subject)
			if subject != "Solicitação de envio código de acesso PJ" {
				continue
			}
		}

		// Process each message's part
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Println(err)
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				// This is the message's text (can be plain-text or HTML)
				b, _ := ioutil.ReadAll(p.Body)
				log.Printf("Got text: %v\n", string(b))
				reg, err := regexp.Compile("<strong>([0-9]+)</strong>")
				if err != nil {
					log.Fatal(err)
				}
				matches := reg.FindSubmatch(b)
				if len(matches) > 0 {
					code = string(matches[1])
				}
			case *mail.AttachmentHeader:
				// This is an attachment
				filename, _ := h.Filename()
				log.Printf("Got attachment: %v\n", filename)
			}
		}
	}
	return code
}
