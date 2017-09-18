package numbers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

type testGetter struct {
	getTimeout time.Duration
}

func (t *testGetter) Get(ctx context.Context, url string) ([]byte, error) {
	params := strings.Split(url, ".")
	sr, rt := params[0], params[1]

	timeoutCh := time.After(t.getTimeout)

	if rt, err := strconv.Atoi(rt); err == nil {
		log.Printf("sleeping for %dms: %s", rt, sr)
		time.Sleep(time.Duration(rt) * time.Millisecond)
	}
	select {
	case <-ctx.Done():
		return nil, errors.New("context timeout")
	case <-timeoutCh:
		return nil, errors.New("request timeout")
	default:
	}

	switch sr {
	case "http://fail":
		return nil, errors.New("service unavailable")
	case "http://rand10":
		return nRandomNumbers(10), nil
	case "http://rand100":
		return nRandomNumbers(100), nil
	case "http://rand1000":
		return nRandomNumbers(1000), nil
	}
	return []byte("a response that will not be parsed"), nil
}

func (t *testGetter) Client() *http.Client {
	return nil
}

func nRandomNumbers(n int) []byte {
	nums := rand.Perm(n)
	res := result{Numbers: nums}

	data, _ := json.Marshal(res)
	return data
}

func TestProcessURLsAllOK(t *testing.T) {
	cfg := newConfig(500*time.Millisecond, 110*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ResponseTimeout)
	defer cancel()

	ch := ProcessURLs(ctx, cfg, []string{"http://rand10.50", "http://rand100.100"})
	var slcCount, numCount int
	for ns := range ch {
		slcCount++
		for _ = range ns {
			numCount++
		}
	}
	if slcCount != 2 {
		t.Fatalf("total slice count mismatch: %s", comp(110, slcCount))
	}
	if numCount != 110 {
		t.Fatalf("total numbers count mismatch: %s", comp(110, numCount))
	}
}

func TestProcessURLsRequestTimeout(t *testing.T) {
	cfg := newConfig(500*time.Millisecond, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ResponseTimeout)
	defer cancel()

	ch := ProcessURLs(ctx, cfg, []string{"http://rand10.10", "http://rand100.100"})
	var slcCount, numCount int
	for ns := range ch {
		slcCount++
		for _ = range ns {
			numCount++
		}
	}
	if slcCount != 2 {
		t.Fatalf("total slice count mismatch: %s", comp(110, slcCount))
	}
	if numCount != 10 {
		t.Fatalf("total numbers count mismatch: %s", comp(110, numCount))
	}
}

func TestProcessURLsResponseTimeout(t *testing.T) {
	cfg := newConfig(50*time.Millisecond, 500*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ResponseTimeout)
	defer cancel()

	ch := ProcessURLs(ctx, cfg, []string{"http://rand10.10", "http://rand100.100"})
	var slcCount, numCount int
	for ns := range ch {
		slcCount++
		for _ = range ns {
			numCount++
		}
	}
	if slcCount != 2 {
		t.Fatalf("total slice count mismatch: %s", comp(110, slcCount))
	}
	if numCount != 10 {
		t.Fatalf("total numbers count mismatch: %s", comp(110, numCount))
	}
}

func TestProcessURLsTooManyURLs(t *testing.T) {
	urls := []string{}
	for i := 0; i < 20; i++ { // total sequentiall fetch time == 200ms.
		urls = append(urls, "http://rand10.10")
	}

	cfg := newConfig(70*time.Millisecond, 500*time.Millisecond)
	cfg.NumGoRoutines = 2

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ResponseTimeout)
	defer cancel()

	ch := ProcessURLs(ctx, cfg, urls)
	var nilSlcCount, numCount int
	for ns := range ch {
		if ns == nil {
			nilSlcCount++
		}
		for _ = range ns {
			numCount++
		}
	}
	if nilSlcCount == 0 {
		t.Fatalf("every slice non-nil, no timeouts: %s", comp("at least 1 nil slice", nilSlcCount))
	}
	if numCount >= 200 {
		t.Fatalf("received numbers from all the URLs, no timeout: %s", comp("< 200 numbers", numCount))
	}
}

func newConfig(res, req time.Duration) *Config {
	return &Config{
		ResponseTimeout: res,
		URLGetter:       &testGetter{req},
	}
}

func comp(exp, got interface{}) string {
	return fmt.Sprintf("expected: %v -- got: %v", exp, got)
}
