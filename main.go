package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"math"
	"strconv"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("shakesearch available at http://localhost:%s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryString := r.URL.Query()
		query, queryOk := queryString["q"]
		page, pageOk := queryString["p"]

		if !queryOk || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		// default page of 0
		pageInt := 0
		if pageOk && len(page[0]) > 0 {
			convertedInt, err := strconv.Atoi(page[0])
			if err == nil {
				pageInt = convertedInt
			}
		}
		
		results := searcher.Search(query[0], pageInt)
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) PopulateSuffixArray(dat []byte) {
	// convert byte array to string
	// make string case insensitive
	caseInsensitiveFileContent := strings.ToLower(string(dat))
	// convert string back to byte array
	// now we have our SuffixArray
	s.SuffixArray = suffixarray.New([]byte(caseInsensitiveFileContent))
}

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.PopulateSuffixArray(dat)
	return nil
}

func (s *Searcher) Search(query string, startPage int) []string {
	startIdx := startPage*20;
	// Lookup returns an unsorted list of at most n indices where the byte
	// string query occurs in the indexed data.
	idxs := s.SuffixArray.Lookup([]byte(strings.ToLower(query)), -1)

	results := []string{}
	if startIdx >= len(idxs) {
		return results;
	}
	// get the end idx, handling out of bounds case
	// only return page size of 20 items
	endIdx := math.Min(float64(startIdx + 20), float64(len(idxs)))
	resultRange := idxs[startIdx:int(endIdx)]

	for _, idx := range resultRange {
		// for each index that s.SuffixArray.Lookup returns for the query, append
		// the 500-character substring centered around the index to the results.
		results = append(results, s.CompleteWorks[idx-250:idx+250])
	}
	// Results contains 500-characer substrings that contain the query at the
	// center.
	return results
}
