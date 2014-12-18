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
	"flag"
	"fmt"
	"google.golang.org/api/genomics/v1beta2"
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
	datasetId := "10473108253681171589" // This is the 1000 Genomes dataset ID
	sample := "NA12872"
	referenceName := "22"
	referencePosition := int64(51003835)

	// 1. First find the read group set ID for the sample
	rsRes, err := svc.Readgroupsets.Search(&genomics.SearchReadGroupSetsRequest{
		DatasetIds: []string{datasetId},
		Name:       sample,
	}).Fields("readGroupSets(id)").Do()
	if err != nil {
		log.Fatal(err)
	}
	if len(rsRes.ReadGroupSets) != 1 {
		fmt.Fprintln(os.Stderr, "Searching for "+sample+" didn't return the right number of read group sets")
		return
	}
	readGroupSetId := rsRes.ReadGroupSets[0].Id

	// 2. Once we have the read group set ID,
	// lookup the reads at the position we are interested in
	rRes, err := svc.Reads.Search(&genomics.SearchReadsRequest{
		ReadGroupSetIds: []string{readGroupSetId},
		ReferenceName:   referenceName,
		Start:           referencePosition,
		End:             referencePosition + 1,
		PageSize:        1024,
	}).Fields("alignments(alignment,alignedSequence)").Do()
	if err != nil {
		log.Fatal(err)
	}

	bases := make(map[uint8]int)
	for _, read := range rRes.Alignments {
		// Note: This is simplistic - the cigar should be considered for real code
		base := read.AlignedSequence[referencePosition-int64(read.Alignment.Position.Position)]
		bases[base]++
	}

	fmt.Printf("%s bases on %s at %d are\n", sample, referenceName, referencePosition)
	for base, count := range bases {
		fmt.Printf("%c: %d\n", base, count)
	}

	//
	// This example gets the variants for a sample at a specific position
	//

	// 1. First find the call set ID for the sample
	csRes, err := svc.Callsets.Search(&genomics.SearchCallSetsRequest{
		VariantSetIds: []string{datasetId},
		Name:          sample,
	}).Fields("callSets(id)").Do()
	if err != nil {
		log.Fatal(err)
	}
	if len(csRes.CallSets) != 1 {
		fmt.Fprintln(os.Stderr, "Searching for "+sample+" didn't return the right number of call sets")
		return
	}
	callSetId := csRes.CallSets[0].Id

	// 2. Once we have the call set ID,
	// lookup the variants that overlap the position we are interested in
	vRes, err := svc.Variants.Search(&genomics.SearchVariantsRequest{
		CallSetIds:    []string{callSetId},
		ReferenceName: referenceName,
		Start:         referencePosition,
		End:           referencePosition + 1,
	}).Fields("variants(names,referenceBases,alternateBases,calls(genotype))").Do()
	if err != nil {
		log.Fatal(err)
	}

	variant := vRes.Variants[0]
	variantName := variant.Names[0]

	genotype := make([]string, 2)
	for i, g := range variant.Calls[0].Genotype {
		if g == 0 {
			genotype[i] = variant.ReferenceBases
		} else {
			genotype[i] = variant.AlternateBases[g-1]
		}
	}

	fmt.Printf("the called genotype is %s for %s\n", genotype, variantName)
}
