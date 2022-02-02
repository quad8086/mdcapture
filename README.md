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
<git clone somewhere>
go mod download github.com/gorilla/websocket
go mod download github.com/jessevdk/go-flags
go get github.com/gorilla/websocket
go get github.com/jessevdk/go-flags
go build
```

Or run make.

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

## Data

Robot #1: "Administer the test."<br>
Robot #2: "Which of the following would you most prefer? <br>
A: a puppy, <br>
B: a pretty flower from your sweety, or <br>
C: a large properly formatted data file?"<br>

```
tabulon -p --tail=10 20220201.ticker.csv
type   |recv_ts                |time                        |product_id |sequence  |qty      |price    |side |trade_id |best_bid |best_ask |open_24h |low_24h  |high_24h |volume_24h   |volume_30d     |
-------+-----------------------+----------------------------+-----------+----------+---------+---------+-----+---------+---------+---------+---------+---------+---------+-------------+---------------+
ticker |20220201-213515.125948 |2022-02-02T02:35:15.115526Z |BTC-USD    |492266212 |0.000016 |38727.21 |buy  |37950153 |38074.00 |38727.21 |38497.17 |30989.81 |45632.17 |114.29059189 |24824.97309690 |
ticker |20220201-213520.132215 |2022-02-02T02:35:20.119573Z |BTC-USD    |492266263 |0.000016 |38074    |sell |37950154 |38074.00 |38727.21 |38497.17 |30989.81 |45632.17 |114.29060789 |24824.97311290 |
ticker |20220201-213525.124184 |2022-02-02T02:35:25.113293Z |BTC-USD    |492266269 |0.000016 |38727.21 |buy  |37950155 |38074.00 |38727.21 |38497.17 |30989.81 |45632.17 |114.29062389 |24824.97312890 |
ticker |20220201-213530.135352 |2022-02-02T02:35:30.124619Z |BTC-USD    |492266272 |0.000016 |38074    |sell |37950156 |38074.00 |38727.21 |38497.17 |30989.81 |45632.17 |114.29063989 |24824.97314490 |
ticker |20220201-213535.116365 |2022-02-02T02:35:35.106240Z |BTC-USD    |492266275 |0.000016 |38727.21 |buy  |37950157 |38074.00 |38727.21 |38497.17 |30989.81 |45632.17 |114.29065589 |24824.97316090 |
ticker |20220201-213540.128401 |2022-02-02T02:35:40.117215Z |BTC-USD    |492266281 |0.000016 |38074    |sell |37950158 |38074.00 |38727.21 |38497.17 |30989.81 |45632.17 |114.29075189 |24824.97325690 |
ticker |20220201-213545.114619 |2022-02-02T02:35:45.102714Z |BTC-USD    |492266284 |0.000016 |38727.21 |buy  |37950159 |38074.00 |38727.21 |38497.17 |30989.81 |45632.17 |114.29076789 |24824.97327290 |
ticker |20220201-213550.115534 |2022-02-02T02:35:50.103987Z |BTC-USD    |492266335 |0.000016 |38074    |sell |37950160 |38074.00 |38727.21 |38497.17 |30989.81 |45632.17 |114.29078389 |24824.97328890 |
ticker |20220201-213555.138683 |2022-02-02T02:35:55.121339Z |BTC-USD    |492266341 |0.000016 |38727.21 |buy  |37950161 |38074.00 |38727.21 |38497.17 |30989.81 |45632.17 |114.29079989 |24824.97330490 |
ticker |20220201-213600.134162 |2022-02-02T02:36:00.118787Z |BTC-USD    |492266344 |0.000016 |38074    |sell |37950162 |38074.00 |38727.21 |38497.17 |30989.81 |45632.17 |114.29081589 |24824.97332090 |
```
