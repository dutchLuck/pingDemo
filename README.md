# pingDemo
Ping another network computer or network device, by using a IP4 icmp socket to send an ICMP Echo request.
This demo code originated from the ideas in section 3.11 of the "Socket-level Programming" chapter of the
Jan Newmarch ebook; -

https://jan.newmarch.name/go/socket/chapter-socket.html

I couldn't get the ping code published in the ebook to work for me, but after some editing it began working.

Compile Code
```
At least go versions "go1.17.5 windows/amd64" & "go1.17.6 linux/amd64" can compile the code; -
go build pingDemo.go
```

Run on Windows
Start cmd window with "Run as Administrator" then; -
```
C:\pingDemo>pingDemo github.com
github.com ( 13.236.229.21 ) responded in 34.3197ms
```

Run on Linux
```
dl@dl:~/pingDemo$ sudo ./pingDemo github
github.com ( 13.236.229.21 ) responded in 288.000125ms
```

Useage
```
pingDemo [-c Int] [-D] [-p Float] [-v] [-w Float] name_Or_IP4NumbersOfDeviceToPing
Usage of pingDemo:
  -D    Turn on Debug output
  -c int
        Count number of Echo Requests for each host (default 1)
  -p string
        Pause between Echo requests in seconds (default "0.1")
  -v    Turn on verbose output
  -w string
        Individual Echo Reply Timeout wait time in seconds (default "2")
```

Versions
```
0v5 Specify duration between multiple pings with pause option -p X.X
0v4 Output ping response delay & Specify multiple pings of the same host with count option -c X 
0v3 Increase size of sent ICMP echo reqest packet & more Debug output
0v2 Better command line option handling: -D(ebug), -v(erbose) -w(ait) <duration>
0v1 Better handling of timeout error messages.
0v0 Code now times-out when device doesn't respond. Flattened error checks.
```
