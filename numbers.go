package numbers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

// result type is for storing the decoded URL responses.
type result struct {
	Numbers []int `json:"numbers"`
}

// URLGetter defines an interface which specifies how to GET an input URL.
// Can be extended/embedded to include caching and other features.
type URLGetter interface {
	// Get performs the HTTP GET request for the given URL. It must return
	// the response in []byte form.
	Get(ctx context.Context, url string) ([]byte, error)

	// Client returns the http.Client that will be used to make the request.
	Client() *http.Client
}

// Config contains various parameters required to setup various package
// functionalities and options.
type Config struct {
	// ResponseTimeout is the Cumulative global timeout by which every
	// requested URL must be queried. Once this timeout is reached, URL
	// queries in flight are cancelled and remaining URLs are ignored.
	ResponseTimeout time.Duration

	// GetTimeout is the individual timeout for each URL required to be queried.
	GetTimeout time.Duration

	// Te maximum number of goroutines the server process should start up.
	NumGoRoutines int

	// URLGetter is the type that performs the GET request for input URLs.
	// If nil, this is set to DefaultGet.
	URLGetter
}

// numGoRoutines is the maximum number of goroutines allowed to run at a time.
// This value can be configured using Config.
var numGoRoutines = 20

// This function returns a channel of []int instead of int's. This helps in case
// a URL returns a very large list of numbers. Sending out the slice header prevent
// allows the functions querying the URL to return in time.
func ProcessURLs(ctx context.Context, cfg *Config, urls []string) <-chan []int {
	if cfg.NumGoRoutines <= 0 {
		cfg.NumGoRoutines = numGoRoutines
	}
	if cfg.URLGetter == nil {
		cfg.URLGetter = NewDefaultGet(cfg.GetTimeout)
	}

	// numbersCh is the channel returned to the caller. Caller can range over this
	// channel to read the number list responses recieved by GETing the input URLS.
	numbersCh := make(chan []int)

	// processURL takes the responsibility of performing all the requests and
	// relaying their response over to caller. This function is also responsible
	// for closing the outbound channel.
	go processURLs(ctx, cfg, urls, numbersCh)
	return numbersCh
}

// processURLs GETs the input URL and sends their response (list of numbers)
// over the out channel.
// This implementation of processURLs spins a fixed number of goroutines, each
// responsible of handling exactly one input URL at a time.
// The function also watches for input context's cancellation and can perform
// an early return accordingly.
func processURLs(ctx context.Context, cfg *Config, urls []string, out chan<- []int) {
	var wg sync.WaitGroup

	wg.Add(cfg.NumGoRoutines)

	// urlCh is used to fan out the input URL over to several goroutines for processing.
	urlCh := make(chan string)

	// Spin numGoRoutines number fo goroutines. Each goroutine waits on urlCh
	// for new work.
	for i := 0; i < cfg.NumGoRoutines; i++ {
		go func() {
			defer wg.Done()
			for url := range urlCh {
				// out is closed only once ever goroutine returns due to the WaitGroup
				// defined above hence send on a close channel is not possible.
				out <- fetchResponse(ctx, cfg, url)
			}
		}()
	}

	for _, url := range urls {
		select {
		case urlCh <- url:
		case <-ctx.Done():
			break
		}
	}
	close(urlCh)

	wg.Wait()
	close(out)
}

// fetchResponse calls the functions to query the input URL. This function also
// decodes the response into appropriate type and returns only the slice of numbers.
// In case of an error, a nil slice is returned.
func fetchResponse(ctx context.Context, ug URLGetter, url string) []int {
	data, err := ug.Get(ctx, url)
	if err != nil {
		log.Printf("error GETing url %s: %v", url, err)
		return nil
	}

	result := result{}

	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Printf("error reading response for %s: %v -- %v", url, err, data)
		return nil
	}
	return result.Numbers
}

// processURLs2 is an alternative implementation of processURLs that can be
// used as a drop in replacement.
// This implementation creates goroutines to query the URLs as they are required
// upto a maximum allowed count. If the number of input URLs is less than
// numGoRoutines, additional goroutines will not be created. This implementation
// is useful if numGoRoutines can be set very high and the maximum number of
// input URLs can also go very high for some requests.
// However since goroutines are relativeky cheap, this implementation is more useful
// for illustratiove purposes.
// The function also watches for input context's cancellation and can perform
// an early return accordingly.
func processURLs2(ctx context.Context, cfg *Config, urls []string, out chan []int) {
	var wg sync.WaitGroup

	limiter := make(chan struct{}, cfg.NumGoRoutines)

	for _, u := range urls {
		// Below select unblocks only when limiter is not full or ctx is cancelled.
		select {
		case limiter <- struct{}{}:
			wg.Add(1)
		case <-ctx.Done():
			return
		}

		go func(url string) {
			defer func() {
				<-limiter
				wg.Done()
			}()
			// Similar sync based measures to processURLs avoids send on closed channels.
			out <- fetchResponse(ctx, cfg, url)
		}(u)
	}

	wg.Wait()
	close(out)
	close(limiter)
}
