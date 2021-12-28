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
tabulon -p --tail=10 coinbase/20211227/20211227.ticker.csv
type   |recv_ts                             |time                        |product_id |sequence    |qty        |price    |side |trade_id  |best_bid |best_ask |open_24h |low_24h |high_24h |volume_24h      |volume_30d       |
-------+------------------------------------+----------------------------+-----------+------------+-----------+---------+-----+----------+---------+---------+---------+--------+---------+----------------+-----------------+
ticker |2021-12-27T21:16:36.794742785-05:00 |2021-12-28T02:16:36.783324Z |BTC-USD    |32522639262 |0.0001     |49711.11 |sell |255807331 |49711.11 |49713.43 |50848.38 |49650   |52100    |13583.69858522  |450664.26012436  |
ticker |2021-12-27T21:16:36.925744196-05:00 |2021-12-28T02:16:36.917442Z |ETH-USD    |23894314999 |0.36       |3976.48  |buy  |198380437 |3976.47  |3976.76  |4073.06  |3968.2  |4128.9   |117478.61385221 |5878037.80416139 |
ticker |2021-12-27T21:16:37.033402378-05:00 |2021-12-28T02:16:37.027262Z |ETH-USD    |23894315098 |0.02525817 |3976.55  |sell |198380438 |3976.55  |3976.75  |4073.06  |3968.2  |4128.9   |117478.63911038 |5878037.82941956 |
ticker |2021-12-27T21:16:37.136770805-05:00 |2021-12-28T02:16:37.128556Z |BTC-USD    |32522639436 |0.04804844 |49714.39 |buy  |255807332 |49714.38 |49714.39 |50848.38 |49650   |52100    |13583.74663366  |450664.30817280  |
ticker |2021-12-27T21:16:37.161934162-05:00 |2021-12-28T02:16:37.148106Z |ETH-USD    |23894315125 |0.01840392 |3976.75  |buy  |198380439 |3976.57  |3976.75  |4073.06  |3968.2  |4128.9   |117478.65751430 |5878037.84782348 |
ticker |2021-12-27T21:16:37.320433039-05:00 |2021-12-28T02:16:37.311613Z |ETH-USD    |23894315202 |0.00225278 |3976.75  |buy  |198380440 |3976.60  |3976.75  |4073.06  |3968.2  |4128.9   |117478.65976708 |5878037.85007626 |
ticker |2021-12-27T21:16:37.546501578-05:00 |2021-12-28T02:16:37.539220Z |BTC-USD    |32522639623 |0.00018028 |49714.39 |buy  |255807333 |49714.38 |49714.39 |50848.38 |49650   |52100    |13583.74681394  |450664.30835308  |
ticker |2021-12-27T21:16:37.601654569-05:00 |2021-12-28T02:16:37.594545Z |BTC-USD    |32522639635 |0.00000478 |49714.39 |buy  |255807334 |49714.38 |49716.35 |50848.38 |49650   |52100    |13583.74681872  |450664.30835786  |
ticker |2021-12-27T21:16:37.60186841-05:00  |2021-12-28T02:16:37.594545Z |BTC-USD    |32522639637 |0.00007544 |49716.35 |buy  |255807335 |49714.38 |49716.35 |50848.38 |49650   |52100    |13583.74689416  |450664.30843330  |
ticker |2021-12-27T21:16:38.089480651-05:00 |2021-12-28T02:16:38.083777Z |ETH-USD    |23894315554 |0.29572443 |3976.43  |sell |198380441 |3976.43  |3976.44  |4073.06  |3968.2  |4128.9   |117478.95549151 |5878038.14580069 |
```
