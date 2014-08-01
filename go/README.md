# Getting started in Go

1. Follow the first step on https://developers.google.com/genomics/ to setup
 some client secrets. Save the "client ID" and "client secret" values from the
"Client ID for native application" you just made.

2. [Install go](http://golang.org/doc/install)
(make sure you set up a `$GOPATH`)

3. Get the client library and oauth dependencies.
(Note: this will [require mercurial](http://golang.org/s/gogetcmd))

```
go get code.google.com/p/google-api-go-client/genomics/v1beta
go get code.google.com/p/goauth2/oauth
```

4. Run the code:

```
go run main.go client_id client_secret
```