package main

import (
	"fmt"
	"net/http"
	"sync"
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
	successStatusUrl := make(chan string)
	attemptsLimit := make(chan struct{}, successStatusCount)

	done := make(chan struct{})
	allHandled := make(chan struct{})

	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go getStatus(url, successStatusUrl, attemptsLimit, done, &wg)
	}

	go func() {
		wg.Wait()
		close(allHandled)
	}()

	counter := 0

	for {
		select {
		case <- allHandled:
			//кейс когда не будет 2 удачных ответа
			fmt.Println("All urls handled")
			return
		case url := <-successStatusUrl:
			counter++
			fmt.Printf(messageTemplate, url)

			if counter == successStatusCount {
				close(done)

				time.Sleep(1 * time.Second) // таймаут чтобы дождаться сообщений о выходе из горутин
				fmt.Println("Get two ok results")
				return
			}
		}
	}
}

func getStatus(url string, successUrlChan chan<-string, attemptsLimit chan struct{}, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-done:
			fmt.Println("close from outside routine with url " + url)
			return
		case attemptsLimit <- struct{}{}:
			if resp, err := client.Get(url); err == nil && resp.StatusCode == http.StatusOK {
				successUrlChan <- url
				return
			}

			_ = <-attemptsLimit //освобождаем в случае неудачи
			return
		}
	}
}

