// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"bufio"
	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/genomics/v1beta"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

var config = &oauth.Config{
	ClientId:     "", // Set by the command line args
	ClientSecret: "", // Set by the command line args
	Scope:        genomics.GenomicsScope,
	RedirectURL:  "oob",
	AuthURL:      "https://accounts.google.com/o/oauth2/auth",
	TokenURL:     "https://accounts.google.com/o/oauth2/token",
	TokenCache:   oauth.CacheFile(".oauth2_cache.json"),
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: main.go client_id client_secret")
	os.Exit(2)
}

func obtainOauthCode(url string) string {
	fmt.Println("Please visit the below URL to obtain OAuth2 code.")
	fmt.Println()
	fmt.Println(url)
	fmt.Println()
	fmt.Println("Please enter the code here:")

	line, _, _ := bufio.NewReader(os.Stdin).ReadLine()

	return string(line)
}

func getOAuthClient(config *oauth.Config) (*http.Client, error) {
	transport := &oauth.Transport{Config: config}
	token, err := config.TokenCache.Token()
	if err != nil {
		url := config.AuthCodeURL("")
		code := obtainOauthCode(url)
		token, err = transport.Exchange(code)
		if err != nil {
			return nil, err
		}
	}

	transport.Token = token
	client := transport.Client()

	return client, nil
}

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		usage()
	}

	// Authorization
	config.ClientId = flag.Args()[0]
	config.ClientSecret = flag.Args()[1]
	client, err := getOAuthClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// Create a genomics API service
	svc, err := genomics.New(client)
	if err != nil {
		log.Fatal(err)
	}

	//
	// This example gets the read bases for NA12878 at specific a position
	//
	datasetId := "376902546192" // This is the 1000 Genomes dataset ID
	referenceName := "22"
	referencePosition := uint64(51005354)

	// 1. First find the readset ID for NA12878
	// TODO: The go client library doesn't currently have support for partial responses
	// see https://code.google.com/p/google-api-go-client/issues/detail?id=38
	rsRes, err := svc.Readsets.Search(&genomics.SearchReadsetsRequest{
		DatasetIds: []string{datasetId},
		Name:       "NA12878",
	}).Do()
	if err != nil {
		log.Fatal(err)
	}
	if len(rsRes.Readsets) != 1 {
		fmt.Fprintln(os.Stderr, "Searching for NA12878 didn't return the right number of results")
		return
	}
	na12878 := rsRes.Readsets[0].Id

	// 2. Once we have the readset ID,
	// lookup the reads at the position we are interested in
	rRes, err := svc.Reads.Search(&genomics.SearchReadsRequest{
		ReadsetIds:    []string{na12878},
		SequenceName:  referenceName,
		SequenceStart: referencePosition,
		SequenceEnd:   referencePosition,
		MaxResults:    1024,
	}).Do()
	if err != nil {
		log.Fatal(err)
	}

	bases := make(map[uint8]int)
	for _, read := range rRes.Reads {
		// Note: This is simplistic - the cigar should be considered for real code
		base := read.OriginalBases[referencePosition-uint64(read.Position)]
		bases[base]++
	}

	fmt.Printf("NA12878 bases on %s at %d\n", referenceName, referencePosition)
	for base, count := range bases {
		fmt.Printf("%c: %d\n", base, count)
	}
}
