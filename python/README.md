# Getting started in python

1. Follow the first step on https://developers.google.com/genomics/ to setup
 some client secrets. Save the "client ID" and "client secret" values from the
"Client ID for native application" you just made.

2. [Install pip](http://www.pip-installer.org/en/latest/installing.html)
3. Install the python client library and run the code:

    ```
    pip install --upgrade google-api-python-client
    python main.py client_id client_secret
    ```


# More information

This example is using [Google's python client library](https://developers.google.com/api-client-library/python/), which has [pydoc for the genomics methods](https://developers.google.com/resources/api-libraries/documentation/genomics/v1beta/python/latest/).
