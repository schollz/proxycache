package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var cache = struct {
	sync.RWMutex
	header map[string]map[string]string
	body   map[string][]byte
}{header: make(map[string]map[string]string), body: make(map[string][]byte)}

var CacheURL string

func init() {
	flag.StringVar(&CacheURL, "url", "", "set URL to be cached")
	flag.Parse()
}

func main() {
	if CacheURL == "" {
		fmt.Println("Must specify url")
		os.Exit(-1)
	}
	if string(CacheURL[len(CacheURL)-1]) != "/" {
		CacheURL = CacheURL + "/"
	}
	fmt.Printf("Caching data on %s", CacheURL)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8001", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer timeTrack(time.Now(), CacheURL+r.URL.Path[1:])
	// check cache
	cache.RLock()
	if _, ok := cache.body[r.URL.Path[1:]]; ok {
		for key := range cache.header[r.URL.Path[1:]] {
			w.Header().Set(key, cache.header[r.URL.Path[1:]][key])
		}
		w.Write(cache.body[r.URL.Path[1:]])
		cache.RUnlock()
	} else {
		cache.RUnlock()
		response, err := http.Get(CacheURL + r.URL.Path[1:])
		if err != nil {
			fmt.Fprintf(w, "Error processing %s", r.URL.Path[1:])
		} else {
			defer response.Body.Close()
			cache.Lock()

			cache.header[r.URL.Path[1:]] = make(map[string]string)
			for key := range response.Header {
				cache.header[r.URL.Path[1:]][key] = response.Header.Get(key)
				w.Header().Set(key, response.Header.Get(key))
			}
			hah, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Fprintf(w, "%s", err)
			}
			cache.body[r.URL.Path[1:]] = hah
			cache.Unlock()
			w.Write(hah)
		}
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
