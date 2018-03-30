package imageproxy

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"

	"code.cloudfoundry.org/bytefmt"
	slack "github.com/ashwanthkumar/slack-go-webhook"
	"github.com/pbnjay/memory"
)

var maxMemoryRecorderMutex = &sync.Mutex{}
var maxMemoryUsed uint64 = 0
var memoryAlartAt float64 = 0.75
var memoryAlartAtExact uint64 = 268435456
var shouldLogMemoryUsage bool = true

func init() {
	var s string

	s = os.Getenv("ALARM_AT_MEMORY")
	if s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err == nil {
			memoryAlartAt = v
		}
	}

	s = os.Getenv("ALARM_AT_MEMORY_EXACT")
	if s != "" {
		v, err := bytefmt.ToBytes(s)
		if err == nil {
			memoryAlartAtExact = v
		}
	}

	s = os.Getenv("DEBUG_LOG_MEMORY_USAGE")
	if s != "" {
		v, err := strconv.ParseBool(s)
		if err == nil {
			shouldLogMemoryUsage = v
		}
	}
}

func postErrorToSlack(color string, title string, text string) {
	webhook := os.Getenv("SLACK_ERROR_WEBHOOK")
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

func logMemoryUsage() {
	var memstats runtime.MemStats
	runtime.ReadMemStats(&memstats)

	using := memstats.HeapAlloc
	total := memory.TotalMemory()

	if false {
		if float64(using/total) > memoryAlartAt || using > memoryAlartAtExact {
			postErrorToSlack(
				"danger",
				"記憶體快爆啦！",
				fmt.Sprint(
					"已用記憶體量: ", bytefmt.ByteSize(using), "\n",
					"警告線: ", bytefmt.ByteSize(memoryAlartAtExact), "\n",
					"系統總記憶體量: ", bytefmt.ByteSize(total), "\n",
				),
			)
		}

		log.Println(fmt.Sprint(
			bytefmt.ByteSize(using),
			"/",
			bytefmt.ByteSize(total),
			", alarming at ",
			bytefmt.ByteSize(memoryAlartAtExact),
		))
	}

	maxMemoryRecorderMutex.Lock()
	if maxMemoryUsed < using {
		maxMemoryUsed = using
		log.Println(fmt.Sprint("maxMemoryUsed: ", using))
	}
	maxMemoryRecorderMutex.Unlock()
}
