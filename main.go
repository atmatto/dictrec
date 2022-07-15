package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type comment struct {
	Id              int64  `json:"id"`
	SentenceId      string `json:"sentence_id"`
	Title           string `json:"title"`
	Translation     string `json:"translation"`
	TranslationLang string `json:"translation_language"`
}

type topicPage struct {
	Comments []comment `json:"comments"`
	More     bool      `json:"more"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:            " + os.Args[0] +
			" TOPIC_ID LANGUAGE_ABBREVIATION\nExample (Korean): " + os.Args[0] +
			" 950      en\nLanguage abbreviation is used to output each sentence and its translation in the right order.")
		return
	}
	topic := os.Args[1]
	lang := "en"
	if len(os.Args) > 2 {
		lang = os.Args[2]
	}

	var wg sync.WaitGroup

	more := true // true if next page available
	var page int64 = 1
	var unavailableTranslations, remaining, all int32
	p := topicPage{}
	maxRequests := 300 // maximum number of goroutines making requests
	guard := make(chan int, maxRequests)
	fmt.Fprintln(os.Stderr, "Downloading...")
	for more {
		resp, err := http.Get("https://duolingo.hobune.stream/data/pages/topic_" + topic + "/top/" + strconv.FormatInt(page, 10) + ".json")
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Fprintln(os.Stderr, "  Error:", err, resp.Status)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "  Error:", err)
		}

		p = topicPage{}

		err = json.Unmarshal(b, &p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "  Error:", err)
		}

		resp.Body.Close()

		more = p.More
		page++

		go func(comments []comment) {
			wg.Add(1)
			atomic.AddInt32(&remaining, int32(len(comments)))
			atomic.AddInt32(&all, int32(len(comments)))
			for _, c := range comments {
				guard <- 1
				go func(c comment) {
					wg.Add(1)
					atomic.AddInt32(&remaining, -1)
					if c.SentenceId != "" {
						found := true
						c.Translation = "?"
						resp, err := http.Get("https://forum-api.duolingo.com/comments/" + strconv.FormatInt(c.Id, 10))
						if err != nil || resp.StatusCode != http.StatusOK {
							resp, err = http.Get("https://duolingo.hobune.stream/data/comments/" + strconv.FormatInt(c.Id, 10) + ".json")
							if err != nil || resp.StatusCode != http.StatusOK {
								fmt.Fprintln(os.Stderr, "  Error ("+c.Title+", "+strconv.FormatInt(c.Id, 10)+"):", err, resp.Status)
								atomic.AddInt32(&unavailableTranslations, 1)
								found = false
							}
						}
						b, err := ioutil.ReadAll(resp.Body)
						if err != nil && found {
							fmt.Fprintln(os.Stderr, "  Error ("+c.Title+", "+strconv.FormatInt(c.Id, 10)+"):", err)
						}

						err = json.Unmarshal(b, &c)
						if err != nil && found {
							fmt.Fprintln(os.Stderr, "  Error ("+c.Title+", "+strconv.FormatInt(c.Id, 10)+"):", err)
						}

						resp.Body.Close()
						if c.TranslationLang == lang {
							fmt.Println(strings.Trim(c.Translation, "\""), "=", strings.Trim(c.Title, "\""))
						} else {
							fmt.Println(strings.Trim(c.Title, "\""), "=", strings.Trim(c.Translation, "\""))
						}
					}
					wg.Done()
					<-guard
				}(c)
			}
			wg.Done()
		}(p.Comments)
	}

	go func() {
		wg.Wait() // wait until the end
		fmt.Fprintln(os.Stderr, "Finished:", unavailableTranslations, "translations missing")
		os.Exit(0)
	}()

	for {
		<-time.After(1 * time.Second)
		if remaining != 0 {
			fmt.Fprintln(os.Stderr, "Remaining:", remaining, "/", all)
		}
	}
}
