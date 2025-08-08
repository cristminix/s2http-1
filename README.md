# s2http

Convert socks proxy to a http(s) proxy

# Prerequisite (if use socks)

A socks proxy (probably a client) is listening

It may not directly connect to a socks server

## Usage

```
Usage of ./s2http:
  -dummy-packet
    	send dummy packets to prevent connection reset errors (default true)
  -dummy-packet-interval int
    	interval in seconds to send dummy packets (default 30)
  -dummy-packet-size int
    	size of dummy packet in bytes (default 1)
  -error-handling string
    	error handling strategy: retry, ignore, or fatal (default "retry")
  -keepalive
    	enable keep-alive connections (default true)
  -keepalive-timeout int
    	keep-alive timeout in seconds (default 60)
  -max-idle-conns int
    	maximum idle connections (default 100)
  -min-packet-interval int
    	minimum interval in seconds to send dummy packets (default 5)
  -port string
    	port to listen (default "8080")
  -pure
    	pure http/https proxy without socks
  -read-timeout int
    	read timeout in seconds (default 5)
  -socks string
    	socks url (default) (default "127.0.0.1:1081")
  -v	verbose (default true)
  -version
    	show version.
  -write-timeout int
    	write timeout in seconds (default 5)
```

## Features

### Dummy Packet Sending

The application can send dummy packets (null bytes) to keep connections alive and prevent connection reset errors such as "broken pipe", "connection reset by peer", and "network is unreachable".

- `-dummy-packet`: Enable/disable dummy packet sending (default: true)
- `-dummy-packet-size`: Size of dummy packet in bytes (default: 1)
- `-dummy-packet-interval`: Interval in seconds to send dummy packets (default: 30)
- `-min-packet-interval`: Minimum interval in seconds to send dummy packets (default: 5)

### Error Handling

Flexible error handling strategies for network errors:

- `-error-handling`: Strategy for handling errors (retry, ignore, or fatal) (default: "retry")

### Connection Management

Advanced connection management features:

- `-keepalive`: Enable/disable keep-alive connections (default: true)
- `-keepalive-timeout`: Keep-alive timeout in seconds (default: 60)
- `-max-idle-conns`: Maximum idle connections (default: 100)

### Timeout Configuration

Separate timeout configuration for read and write operations:

- `-read-timeout`: Read timeout in seconds (default: 5)
- `-write-timeout`: Write timeout in seconds (default: 5)

## Pure proxy

Start as a pure proxy (without socks)

    ./s2http -pure

## Examples

### Basic usage

```
./s2http
```

### Custom port and socks proxy

```
./s2http -port 8081 -socks 127.0.0.1:1080
```

### Enable verbose mode

```
./s2http -v
```

### Configure dummy packet sending

```
./s2http -dummy-packet -dummy-packet-size 10 -dummy-packet-interval 10
```

### Set error handling strategy

```
./s2http -error-handling ignore
```

### Configure timeouts

```
./s2http -read-timeout 10 -write-timeout 10
```

### Pure HTTP/HTTPS proxy (without SOCKS)

```
./s2http -pure
```

### Advanced configuration example

```
./s2http -pure -port 8088 -socks 127.0.0.1:1080 -read-timeout=5000 -write-timeout=1000  -keepalive=true -keepalive-timeout=3000 -max-idle-conns=500 -dummy-packet=true -v -dummy-packet-size=128 -dummy-packet-interval=3
```

## Credits

Inspired by [GopherFromHell](https://www.reddit.com/user/GopherFromHell),
As he provided [the example](https://play.golang.org/p/l0iLtkD1DV)

end.
