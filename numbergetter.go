// This file contains a default implementation of URLGetter interface. This is a
// basic implementation of the interface that is used by default by the package.
package numbers

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

// defaultGet implements URLGetter. This type has a simple implementation and only
// requires an explicit timeout (response timeout) for instantiation.
type defaultGet struct {
	client *http.Client
}

func NewDefaultGet(t time.Duration) *defaultGet {
	return &defaultGet{
		client: &http.Client{
			Timeout: t,
		},
	}
}

// Get performs the network request to GET the URL. The requests are created with
// the input context so that they may respect global timeouts and cancellations.
func (g *defaultGet) Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := g.Client().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("service unavailable")
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Client returns the http.Client associated with the type.
func (g *defaultGet) Client() *http.Client {
	if g.client == nil {
		g.client = &http.Client{}
	}
	return g.client
}
