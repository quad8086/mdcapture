<h1 align="center">mdcapture</h2>
<p align="center">
Websocket market data capture utility for coinbase (written in Go)
</p>
<p align="center">

## Features

* Live subscription and capture of Coinbase market data
* Intended for daily recording workflow
* Supports trade, level2 and match streams
* Output is normalized CSV files per message type
* Captures receive timestamp for latency estimation
* Optionally can capture raw JSON messages instead
* Auto shutdown at the end of the day

## Build

```sh
go mod download github.com/gorilla/websocket
go mod github.com/jessevdk/go-flags
go build
```

## Run

By default mdcapture uses the sandbox websockets endpoint. This allows for easy experimentation without worrying about your IP
getting blacklisted.

```sh
./mdcapture -p BTC-USD -s quotes_trades -D coinbase.sandbox/20211227 --status
```

Once comfortable, you can connect to the live endpoint:

```sh
./mdcapture -p BTC-USD -p ETH-USD -s quotes_trades -E wss://ws-feed.exchange.coinbase.com -D coinbase/20211227 --status
```
