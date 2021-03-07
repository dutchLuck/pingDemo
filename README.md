# pingDemo
Ping another network computer or network device, by using a IP4 icmp socket to send an ICMP Echo request.
This demo code originated from the ideas in section 3.11 of the "Socket-level Programming" chapter of the
Jan Newmarch ebook; -

https://jan.newmarch.name/go/socket/chapter-socket.html

I couldn't get the ping code published in the ebook to work for me, but after some editing it began working.

```
At least "go version go1.15.6 windows/amd64" can compile the code; -
go build pingDemo.go
```
```
0v1 Better handling of timeout error messages.
0v0 Code now times-out when device doesn't respond. Flattened error checks.
```
