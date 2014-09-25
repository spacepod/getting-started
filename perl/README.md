# Getting started in perl

1. Follow the first step on https://developers.google.com/genomics/ to setup
 some client secrets. Copy the client_secrets.json file into this directory.

2. Install the perl dependencies and run the code:

    ```
    cpanm Path::Class
    cpanm Net::OAuth2::Client
    cpanm Mozilla::CA
    cpanm LWP::Protocol::https
    perl main.pl
    ```

# Troubleshooting

If you see the error:
```
Can't locate object method "host_port" via package "URI::_generic" at /Library/Perl/5.16/Net/OAuth2/Profile.pm line 197.
```

it's because the access token in `credentials.dat` has expired, and this error
shows up when `Net::OAuth2` tries to refresh it. I don't know how to fix this issue -
so if you do, a pull request (or explanatory issue) would be very welcome!

In the meantime, you can just manually remove the `credentials.dat` file
and everything will work again:
```
cd getting-started/perl
rm credentials.dat
```

For everything else,
[file an issue](https://github.com/googlegenomics/getting-started/issues/new)
if you run into problems and we'll do our best to help!
