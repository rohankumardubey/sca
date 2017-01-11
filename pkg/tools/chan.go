package tools

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

//MergeChan merge chan (code from Go documentatio)
func MergeChan(cs ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup
	out := make(chan interface{})

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan interface{}) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

//Debounce (inspired from https://nathanleclaire.com/blog/2014/08/03/write-a-function-similar-to-underscore-dot-jss-debounce-in-golang/)
func Debounce(interval time.Duration, maxDebounce int, input <-chan interface{}, f func(arg interface{})) {
	var (
		data  interface{}
		event = 0
	)
	for {
		if data == nil {
			data = <-input
		}
		select {
		case data = <-input:
			log.WithFields(log.Fields{
				"func":     f,
				"debounce": interval,
				"data":     data,
			}).Debug("Debounce call to func")
			event++

			if event >= maxDebounce { //Force call
				log.WithFields(log.Fields{
					"func":     f,
					"debounce": interval,
					"event":    event,
					"data":     data,
				}).Debug("Force call to func (debounced)")
				f(data)
				data = nil
				event = 0
			}
		case <-time.After(interval):
			log.Debug("Call func debounce timeout ended")
			if data != nil {
				f(data)
				data = nil
				event = 0
			}
		}
	}
}
