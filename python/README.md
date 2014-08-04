# Getting started in python

1. Follow the first step on https://developers.google.com/genomics/ to setup
 some client secrets. Save the "client ID" and "client secret" values from the
"Client ID for native application" you just made.

2. [Install pip](http://www.pip-installer.org/en/latest/installing.html)
3. Install the python client library and run the code:

    ```
    pip install --upgrade google-api-python-client
    python main.py client_id_string client_secret_string
    ```

# More information

This example is using [Google's python client library](https://developers.google.com/api-client-library/python/), which has [pydoc for the genomics methods](https://developers.google.com/resources/api-libraries/documentation/genomics/v1beta/python/latest/).

# Troubleshooting

If your browser opens a window which says "Error: invalid_client" (as in [this issue](https://github.com/googlegenomics/getting-started/issues/1)) then the client_id and secret are most likely invalid. 

Make sure you followed Step #1 above to setup the "Client ID for native application". The resulting client ID should look something like `xxx.apps.googleusercontent.com`, and the client secret will be a random string like `abc123`. The python command would then be similar to:
```
python main.py xxx.apps.googleusercontent.com abc123
```
