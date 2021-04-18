package main

import (
	"context"
	"fmt"
	"net/http"
	netUrl "net/url"
	"sync"
	"time"
)

const (
	requestTimeout           = 3 * time.Second
	successStatusCount       = 2
	messageOkTemplate        = "%s status ok \n"
	messageFetchTryTemplate  = "fetch try for url : %s \n"
	messageFetchStopTemplate = "fetch stopped from outside for url : %s"
	messageFetchTimeout      = "fetch timeout for url : %s \n"
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

var (
	successUrlsChan chan string
	attemptsPool chan struct{}
	finalizeChan chan string
)

func init() {
	client = &http.Client{
		Timeout: requestTimeout,
	}

	successUrlsChan = make(chan string)
	attemptsPool = make(chan struct{}, successStatusCount)
	finalizeChan = make(chan string) //If no need for finalize message - could be removed
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
    defer func() {
    	cancel()

    	//Just to see routines finalize messages
		for {
			select {
			case finalizeMessage := <- finalizeChan:
				fmt.Println(finalizeMessage)
			case <- time.After(time.Second):
				//Leave 1 sec timeout to wait for all finalize messages
				close(finalizeChan)
				return
			}
		}
	}()

	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go fetchStatus(url, ctx, &wg)
	}

	//Donovan, Kernighan - used this in their book (2019 release - seems not so far)
	go func() {
		wg.Wait()
		close(successUrlsChan)
	}()

	counter := 0

	for url := range successUrlsChan {
		counter++
		fmt.Printf(messageOkTemplate, url)
		if counter == successStatusCount {
			return
		}
	}
}

func fetchStatus(url string, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			finalizeChan <- fmt.Sprintf(messageFetchStopTemplate, url)
			return
		case attemptsPool <- struct{}{} :
			fmt.Printf(messageFetchTryTemplate, url)

			resp, err := client.Do(request)
			if err == nil && resp.StatusCode == http.StatusOK {
				successUrlsChan <- url
			}

			if err != nil && err.(*netUrl.Error).Timeout() {
				fmt.Printf(messageFetchTimeout, url)
			}

			<- attemptsPool
			return
		}
	}
}

