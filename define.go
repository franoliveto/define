// Copyright 2025 Francisco Oliveto. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Define uses the Wordnik API from the command line to define its argument.
// https://developer.wordnik.com/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Definitions struct {
	ID               string
	PartOfSpeech     string
	AttributionText  string
	SourceDictionary string
	Text             string
	Sequence         string
	Score            int
	Word             string
	ExampleUses      []struct {
		Text     string
		Position int
	}
	AttributionURL string
	WordnikURL     string
	Labels         []struct {
		Text string
		Type string
	}
	Citations []struct {
		Source string
		Cite   string
	}
	Notes        []struct{}
	RelatedWords []struct{}
	TextProns    []struct{}
}

type Pronunciations struct {
	ID              string
	Raw             string
	RawType         string
	Seq             int
	AttributionURL  string
	AttributionText string
}

var key = flag.String("key", "", "Wordnik API key (defaults to $WORDNIKAPIKEY)")

func getWord(word string, object string, values url.Values, v any) error {
	url := "https://api.wordnik.com/v4/word.json/" + url.PathEscape(word) + "/" + object + "?" + values.Encode()
	err := get(url, v)
	if err != nil {
		return err
	}
	return nil
}

func get(url string, v any) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var e struct {
			Message string
		}
		if err := dec.Decode(&e); err != nil {
			return err
		}
		return fmt.Errorf("%s", e.Message)
	}
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}

func main() {
	log.SetFlags(0)
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: define [options] <word>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *key == "" {
		*key = os.Getenv("WORDNIKAPIKEY")
	}
	if *key == "" {
		const name = ".define-api-key"
		filename := filepath.Clean(os.Getenv("HOME") + "/" + name)
		shortFilename := "$HOME/" + name
		data, err := os.ReadFile(filename)
		if err != nil {
			log.Fatal("reading API key: ", err, "\n\n"+
				"Please request your WORDNIK API key at https://wordnik.com and write it\n"+
				"to $WORDNIKAPIKEY or ", shortFilename, " to use this program.\n")
		}
		*key = strings.TrimSpace(string(data))
	}
	v := make(url.Values)
	v.Set("api_key", *key)
	v.Set("sourceDictionaries", "ahd-5")
	word := flag.Arg(0)
	var defs []Definitions
	if err := getWord(word, "definitions", v, &defs); err != nil {
		log.Fatal(err)
	}
	var prons []Pronunciations
	if err := getWord(word, "pronunciations", v, &prons); err != nil {
		log.Fatal(err)
	}

	m := make(map[string][]string)
	for _, d := range defs {
		s := d.PartOfSpeech
		if d.Text != "" {
			m[s] = append(m[s], d.Text)
		}
	}
	fmt.Printf("%s /%s/\n", word, prons[0].Raw)
	for speech, texts := range m {
		fmt.Printf("\n%s\n", speech)
		for i, t := range texts {
			fmt.Printf("%d. %s\n", i+1, t)
		}
	}
}
