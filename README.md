
This is a stripped down version of the web server that runs on [www.aicodix.de](https://www.aicodix.de/)

It serves markdown files processed to html.

There are no configuration files.
If you need to change something, edit the source!

To run this web server, you need a signed SSL certificate.
For testing, you can create a self-signed SSL certificate yourself:
```
# openssl req -new -x509 -sha256 -newkey rsa:2048 -nodes -keyout key -out cer -subj "/CN=localhost"
```

To build the web server, you will need Blackfriday from Russ Ross:
```
# go get -v -u gopkg.in/russross/blackfriday.v2
```

To run the web server inside a chroot, it is convenient to have a static binary:
```
# CGO_ENABLED=0 go build www.go
```

To be able to run the web server without root privileges needed to open ports 80 and 443:
```
# sudo setcap cap_net_bind_service=+ep www
```

And finally, starting the web server inside a chroot:
```
# sudo chroot --userspec=nobody:nobody . /www
```

