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
	sample := "NA12872"
	referenceName := "22"
	referencePosition := uint64(51003836)

	// 1. First find the readset ID for the sample
	// TODO: The go client library doesn't currently have support for partial responses
	// see https://code.google.com/p/google-api-go-client/issues/detail?id=38
	rsRes, err := svc.Readsets.Search(&genomics.SearchReadsetsRequest{
		DatasetIds: []string{datasetId},
		Name:       sample,
	}).Do()
	if err != nil {
		log.Fatal(err)
	}
	if len(rsRes.Readsets) != 1 {
		fmt.Fprintln(os.Stderr, "Searching for "+sample+" didn't return the right number of readsets")
		return
	}
	readsetId := rsRes.Readsets[0].Id

	// 2. Once we have the readset ID,
	// lookup the reads at the position we are interested in
	rRes, err := svc.Reads.Search(&genomics.SearchReadsRequest{
		ReadsetIds:    []string{readsetId},
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

	fmt.Printf("%s bases on %s at %d are\n", sample, referenceName, referencePosition)
	for base, count := range bases {
		fmt.Printf("%c: %d\n", base, count)
	}

	//
	// This example gets the variants for a sample at a specific position
	// TODO: The Go client library hasn't updated in a long while, so it
	// doesn't have real variant support and none of this works!

	//	// 1. First find the call set ID for the sample
	//	csRes, err := svc.Callsets.Search(&genomics.SearchCallSetsRequest{
	//		VariantSetIds: []string{datasetId},
	//		Name:       sample,
	//	}).Do()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	if len(csRes.CallSets) != 1 {
	//		fmt.Fprintln(os.Stderr, "Searching for " + sample + " didn't return the right number of call sets")
	//		return
	//	}
	//	callSetId := csRes.CallSets[0].Id
	//
	//	// 2. Once we have the call set ID,
	//	// lookup the variants that overlap the position we are interested in
	//	vRes, err := svc.Variants.Search(&genomics.SearchVariantsRequest{
	//		CallSetIds:    []string{callSetId},
	//		ReferenceName:  referenceName,
	//      // Note: currently, variants are 0-based and reads are 1-based,
	//      // reads will move to 0-based coordinates in the next version of the API
	//		Start: referencePosition - 1,
	//		End:   referencePosition,
	//	}).Do()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//  variant := vRes.Variants[0]
	//  variantName := variant.Names[0]
	//
	//	genotype := make([]string, 2)
	//	for i, g := range variant.Calls[0].Genotype {
	//		if (g == 0) {
	//			genotype[i] = variant.ReferenceBases
	//		} else {
	//			genotype[i] = variant.AlternateBases[g - 1]
	//		}
	//	}
	//
	//	fmt.Printf("the called genotype is %s for %s", sample, variantName)
}
