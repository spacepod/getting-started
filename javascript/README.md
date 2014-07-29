# Getting started in javascript

1. Javascript requires a slightly different set of client secrets than the other 
   projects. To set up a client ID:

* First create a [Genomics enabled project](https://console.developers.google.com/flows/enableapi?apiid=genomics)
  in the Google Developers Console.

* Once you are redirected to the **Credentials** tab, click **Create new Client ID** under
  the OAuth section.

* Set **Application type** to **Web application**, and change
  the **Authorized javascript origins** to `http://localhost:8000`

* Click the **Create Client ID** button

* From the newly created **Client ID for web application**, save the `Client ID`
  value.

Follow the first step on https://developers.google.com/genomics/ to setup
 some client secrets. Copy the client_secrets.json file into this directory.

2. Run a http server (this requires [python](https://www.python.org/download/)):
```
python -m SimpleHTTPServer 8000
```

3. Using the Client ID value that you made in step 1, view the code at: 
   http://localhost:8000#your-client-id-goes-here
