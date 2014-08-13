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

  endpoint='https://www.googleapis.com/genomics/v1beta/'

  #
  # This example gets the read bases for NA12878 at specific a position
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

  # 1. First find the readset ID for NA12878
  readsetId <- reactive({
    validate(
      need(input$datasetId != '', label = 'Dataset ID'),
      need(input$readsetName != '', label = 'Readset Name')
    )

    body <- list(datasetIds=list(input$datasetId), name=input$readsetName)

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
    text <- list(paste(input$readsetName, 'bases on', input$chr, 'at',
        input$position))
    for(base in names(counts)) {
      text <- append(text, paste(base, ':', counts[[base]]))
    }

    div(lapply(text, div))
  })
})