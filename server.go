package numbers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
)

// NumbersGetter is the exported type that handles incoming requests.
// It satisfies the http.ServeHTTP interface.
type NumbersGetter struct {
	Config
}

// ServeHTTP handles incoming requests. It does the following:
//     1. P
func (ng *NumbersGetter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal("incorrect form")
	}

	urls := r.Form["u"]
	log.Printf("URLS: %v", urls)

	ctx, cancel := context.WithTimeout(r.Context(), ng.ResponseTimeout)
	defer cancel()

	numbersCh := ProcessURLs(ctx, &ng.Config, urls)

	numbersMap := make(map[int]bool)
	for ns := range numbersCh {
		for _, n := range ns {
			numbersMap[n] = true
		}
	}

	response := []int{}
	for k, _ := range numbersMap {
		response = append(response, k)
	}

	sort.Ints(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]interface{}{"Numbers": response})
}
