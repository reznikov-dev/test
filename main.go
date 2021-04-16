package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	requestTimeout     = 3 * time.Second
	successStatusCount = 2
	messageTemplate    = "%s status ok \n"
)

var urls = []string{
	"http://ozon.ru",
	"http://google.com",
	"ht://er",
	"http://278462476ggfgfff.fcom",
	"http://somesite.com",
	"https://ozon.ru",
	"http://www.ozon.ru",
	"http://ya.ru",
	"http://avito.ru",
	"https://ya.ru",
}

var client *http.Client

func init() {
	client = &http.Client{
		Timeout: requestTimeout,
	}
}

func main() {
	for url := range fetchStatus(getUrls()) {
		fmt.Printf(messageTemplate, url)
	}
}

func getUrls() <-chan string {
	urlsChan := make(chan string)

	go func() {
		for _, url := range urls {
			urlsChan <- url
		}

		close(urlsChan)
	}()

	return urlsChan
}

func fetchStatus(fetchUrl <-chan string) <- chan string  {
	successUrlChan := make(chan string)
	var counter int

	go func() {
		for {
			select {
			case url, ok := <-fetchUrl:
				if !ok {
					close(successUrlChan)
					fmt.Println("all urls handled")
					return
				}

				if resp, err := client.Get(url); err == nil && resp.StatusCode == http.StatusOK {
					counter++
					if counter > successStatusCount {
						close(successUrlChan)
						fmt.Printf("get ok result expected count %d", successStatusCount)
						return
					}

					successUrlChan <- url
				}
			}
		}
	}()

	return successUrlChan
}

