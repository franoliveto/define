// Copyright 2025 Francisco Oliveto. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Define uses the dictionaryapi.dev API from the command line to define its argument.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Response struct {
	Word      string
	Phonetic  string
	Phonetics []struct {
		Text      string
		Audio     string
		SourceURL string
		License   License
	}
	Meanings   []Meaning
	License    License
	SourceURLs []string
}

type Meaning struct {
	PartOfSpeech string
	Definitions  []struct {
		Definition string
		Synonyms   []string
		Antonyms   []string
		Example    string
	}
	Synonyms []string
	Antonyms []string
}

type License struct {
	Name string
	URL  string
}

type Error struct {
	Title      string
	Message    string
	Resolution string
}

var language = flag.String("language", "en", "dictionary language (two-letter code)")

func main() {
	log.SetFlags(0)
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: define [options] word")
		flag.PrintDefaults()
		os.Exit(1)
	}
	word := flag.Arg(0)
	url := "https://api.dictionaryapi.dev/api/v2/entries/" + url.PathEscape(*language) + "/" + url.PathEscape(word)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var e Error
		if err := dec.Decode(&e); err != nil {
			log.Fatal(err)
		}
		log.Fatalln(e.Title)
	}
	var responses []Response
	if err := dec.Decode(&responses); err != nil {
		log.Fatal(err)
	}
	var phonetic string
	m := make(map[string][]string)
	for _, r := range responses {
		word = r.Word
		phonetic = r.Phonetic
		for _, meaning := range r.Meanings {
			s := meaning.PartOfSpeech
			for _, d := range meaning.Definitions {
				m[s] = append(m[s], d.Definition)
			}
		}
	}
	fmt.Printf("%s %s\n", word, phonetic)
	for speech, definitions := range m {
		fmt.Printf("\n%s\n", speech)
		for i, d := range definitions {
			fmt.Printf("%d. %s\n", i+1, d)
		}
	}
}
