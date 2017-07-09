package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string, depth int, ch_out chan<- Result, done chan<- bool)
}

type Result struct {
	url, body string
	urls      []string
	err       error
	depth     int
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	cache := make(map[string]bool)

	results := make(chan Result)
	done := make(chan bool)

	defer close(results)
	defer close(done)

	var fetchers int

	fetchers++
	go fetcher.Fetch(url, depth, results, done)

	for fetchers != 0 {
		select {
		case res := <-results:
			cache[res.url] = true

			if res.err != nil {
				fmt.Println(res.err)
				continue
			}
			fmt.Printf("found: %s %q\n", res.url, res.body)

			if res.depth <= 0 {
				break
			}

			for _, u := range res.urls {
				if cache[u] {
					fmt.Printf("cache: %s\n", u)
					continue
				}
				fetchers++
				go fetcher.Fetch(u, res.depth-1, results, done)
			}

		case <-done:
			fetchers--
		}
	}
	return
}

func main() {
	Crawl("http://golang.org/", 4, fetcher)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string, depth int, ch_out chan<- Result, done chan<- bool) {
	defer func() { done <- true }()
	if res, ok := f[url]; ok {
		ch_out <- Result{url, res.body, res.urls, nil, depth}
		return
	}
	ch_out <- Result{url, "", nil, fmt.Errorf("not found: %s", url), depth}
	return
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
