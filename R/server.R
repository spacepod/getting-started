# Copyright 2014 Google Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the 'License');
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an 'AS IS' BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
library(shiny)
library(httr)
library(jsonlite)
library(httpuv)

shinyServer(function(input, output) {

  endpoint = 'https://www.googleapis.com/genomics/v1beta/'

  #
  # This example gets the read bases for a sample at specific a position
  #
  google_token <- reactive({
    validate(
      need(input$clientId != '', label = 'Client ID'),
      need(input$clientSecret != '', label = 'Client secret')
    )
    app <- oauth_app('google', input$clientId, input$clientSecret)
    oauth2.0_token(oauth_endpoints('google'), app,
        scope = 'https://www.googleapis.com/auth/genomics')
  })

  # 1. First find the readset ID for the sample
  readsetId <- reactive({
    validate(
      need(input$datasetId != '', label = 'Dataset ID'),
      need(input$sample != '', label = 'Sample Name')
    )

    body <- list(datasetIds=list(input$datasetId), name=input$sample)

    res <- POST(paste(endpoint, 'readsets/search', sep=''),
        query=list(fields='readsets(id)'),
        body=toJSON(body, auto_unbox=TRUE), config(token=google_token()),
        add_headers('Content-Type'='application/json'))
    stop_for_status(res)

    readsets <- content(res)$readsets
    validate(need(length(readsets) > 0, 'No readsets found for that name'))

    readsets[[1]]$id
  })

  # 2. Once we have the readset ID,
  # lookup the reads at the position we are interested in
  baseCounts <- reactive({
    validate(
      need(input$chr != '', label = 'Sequence name'),
      need(input$position > 0, 'Position must be greater than 0')
    )

    body <- list(readsetIds=list(readsetId()), sequenceName=input$chr,
        sequenceStart=input$position, sequenceEnd=input$position,
        maxResults=1024)

    res <- POST(paste(endpoint, 'reads/search', sep=''),
        query=list(fields='reads(position,originalBases,cigar)'),
        body=toJSON(body, auto_unbox=TRUE), config(token=google_token()),
        add_headers('Content-Type'='application/json'))
    stop_for_status(res)

    reads <- content(res)$reads
    validate(need(length(reads) > 0, 'No reads found for that position'))

    positions = input$position - as.integer(sapply(reads, '[[', 'position')) + 1
    bases = sapply(reads, '[[', 'originalBases')
    bases = substr(bases, positions, positions)

    table(bases)
  })
    
  output$baseCounts <- renderUI({
    counts <- baseCounts()
    text <- list(paste(input$sample, 'bases on', input$chr, 'at',
        input$position, 'are'))
    for(base in names(counts)) {
      text <- append(text, paste(base, ':', counts[[base]]))
    }

    div(lapply(text, div))
  })


  #
  # This example gets the variants for a sample at a specific position
  #

  # 1. First find the call set ID for the sample
  callSetId <- reactive({
    validate(
      need(input$datasetId != '', label = 'Dataset ID'),
      need(input$sample != '', label = 'Sample Name')
    )

    body <- list(variantSetIds=list(input$datasetId), name=input$sample)

    res <- POST(paste(endpoint, 'callsets/search', sep=''),
        query=list(fields='callSets(id)'),
        body=toJSON(body, auto_unbox=TRUE), config(token=google_token()),
        add_headers('Content-Type'='application/json'))
    stop_for_status(res)

    callSets <- content(res)$callSets
    validate(need(length(callSets) > 0, 'No call sets found for that name'))

    callSets[[1]]$id
  })

  # 2. Once we have the call set ID,
  # lookup the variants that overlap the position we are interested in
  output$genotype <- renderUI({
    validate(
      need(input$chr != '', label = 'Sequence name'),
      need(input$position > 0, 'Position must be greater than 0')
    )

    body <- list(callSetIds=list(callSetId()), referenceName=input$chr,
        # Note: currently, variants are 0-based and reads are 1-based,
        # reads will move to 0-based coordinates in the next version of the API
        start=input$position - 1, end=input$position)

    res <- POST(paste(endpoint, 'variants/search', sep=''),
        query=list(fields=
          'variants(names,referenceBases,alternateBases,calls(genotype))'),
        body=toJSON(body, auto_unbox=TRUE), config(token=google_token()),
        add_headers('Content-Type'='application/json'))
    stop_for_status(res)

    variants <- content(res)$variants
    validate(need(length(variants) > 0, 'No variants found for that position'))
    variant <- variants[[1]]
    variantName <- variant$names[[1]]

    genotype <- lapply(variant$calls[[1]]$genotype, function (g) {
      if (g == 0) {
        variant$referenceBases
      } else {
        variant$alternateBases[[g]]
      }
    })

    div(paste('the called genotype is', paste(genotype, collapse = ','),
        'for', variantName)[[1]])
  })
})