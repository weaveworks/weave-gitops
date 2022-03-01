
## Prerequisites

You will need openssl cli. You can find some notes on how to install it [here](https://github.com/openssl/openssl#build-and-install)

## Create self-signed certificate
```
openssl req -x509 -nodes -newkey rsa:2048 -keyout server.rsa.key -out server.rsa.crt -days 3650 -subj '/CN=localhost'
```

## Enable your browser to accept the certificate
### Chrome
    
1.- Go to chrome://flags/#allow-insecure-localhost
2.- Select Enable on 'Allow invalid certificates for resources loaded from localhost.'
3.- Click on Relaunch

### Firefox
1.- Go to Preferences
2.- Click on Privacy & Security
3.- Scroll down to Certificates
4.- Click on View Certificates
5.- Select Servers tab
6.- Click on Add Exception 
7.- Enter https://localhost:9447
8.- Click on Get Certificate
9.- Finally click on Confirm Security Exception