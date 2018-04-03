package imageproxy

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	slack "github.com/ashwanthkumar/slack-go-webhook"
)

var (
	alertedOnce            map[string]bool
	alertedOnceMux         sync.Mutex
	alertedClearupInterval int = 3600
	webhook                string
)

func init() {
	var i int
	var err error

	alertedOnce = make(map[string]bool)
	webhook = os.Getenv("SLACK_ERROR_WEBHOOK")

	i, err = strconv.Atoi(os.Getenv("SLACK_ERROR_ALERTED_CLEARUP_INTERVAL"))
	if err == nil {
		alertedClearupInterval = i
	}

	if alertedClearupInterval != 0 {
		clearAlertedOnceTaskTimer := time.NewTicker(time.Duration(alertedClearupInterval) * time.Second)
		go func() {
			for {
				select {
				case <-clearAlertedOnceTaskTimer.C:
					alertedOnceMux.Lock()
					alertedOnce = make(map[string]bool)
					alertedOnceMux.Unlock()
				}
			}
		}()
	}
}

func postErrorToSlack(color string, title string, text string) {
	alertedOnceMux.Lock()
	if alertedOnce[text] {
		alertedOnceMux.Unlock()
		return
	}
	alertedOnce[text] = true
	alertedOnceMux.Unlock()

	if len(webhook) == 0 {
		log.Println("SLACK_ERROR_WEBHOOK no set, skipping postErrorToSlack")
		return
	}

	errs := slack.Send(
		webhook,
		"",
		slack.Payload{
			Text: "",
			Attachments: []slack.Attachment{slack.Attachment{
				Title: &title,
				Color: &color,
				Text:  &text,
			}},
		},
	)

	if len(errs) > 0 {
		fmt.Printf("error: %s\n", errs)
	}
}
