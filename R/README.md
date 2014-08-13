# Getting started in R

1. Follow the first step on https://developers.google.com/genomics/ to setup
  some client secrets.  Save the "client ID" and "client secret" values from the
  "Client ID for native application" you just made.

2. Run this code from R with:
  ```
  devtools::install_github("shiny", "rstudio")
  library(shiny)
  runGitHub("getting-started", "googlegenomics", subdir = "R")
  ```

  and fill in the client ID and secret values in the web page that opens.

3. Or view the hosted version here:
  http://googlegenomics.shinyapps.io/getting-started

  Note: the hosted version is using hardcoded auth, because OAuth won't work in
  that environment. (The validate lines are unfortunately also disabled until
  the next version of Shiny is available)