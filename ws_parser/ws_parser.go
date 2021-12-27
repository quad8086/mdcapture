package ws_parser

import (
	"os"
	"io"
	"fmt"
	"log"
	"time"
	"encoding/csv"
	"encoding/json"
	"crypto/tls"
	"github.com/gorilla/websocket"
)

const (
	channelTicker = "ticker"
	channelLevel2 = "level2"
	channelFull = "full"
	channelStatus = "status"
	channelMatches = "matches"
	channelReceived = "received"
	channelHeartbeat = "heartbeat"
)

type SubscribeReq struct {
	Type       string   `json:"type"`
	ProductIds []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

type GenericResponse struct {
	Type string `json:"type"`
}

type MatchResponse struct {
	Type string `json:"type"`
	TradeID int64 `json:"trade_id"`
	MakerOrderID string `json:"maker_order_id"`
	TakerOrderID string `json:"taker_order_id"`
	Side string `json:"side"`
	Size string `json:"size"`
	Price string `json:"price"`
	ProductID string `json:"product_id"`
	Sequence int64 `json:"sequence"`
	Time string `json:"time"`
}

type HeartbeatResponse struct {
	Type string `json:"type"`
	Sequence int64 `json:"sequence"`
	LastTradeID int64 `json:"last_trade_id"`
	ProductID string `json:"product_id"`
	Time string `json:"time"`
}

type ErrorResponse struct {
	Type string `json:"type"`
	Message string `json:"message"`
	Reason string `json:"reason"`
}

type ChannelResponse struct {
	Name string `json:"name"`
	ProductIDs []string `json:"product_ids"`
}

type SubscribeResponse struct {
	Type string                  `json:"type"`
	Channels []ChannelResponse `json:"channels"`
}

type TickerResponse struct {
	Type string `json:"type"`
	TradeID int64 `json:"trade_id"`
	Sequence int64 `json:"sequence"`
	Time string `json:"time"`
	ProductID string `json:"product_id"`
	Price string `json:"price"`
	Side string `json:"side"`
	Size string `json:"last_size"`
	BestBid string `json:"best_bid"`
	BestAsk string `json:"best_ask"`
	Volume24h string `json:"volume_24h"`
	Open24h string `json:"open_24h"`
	High24h string `json:"high_24h"`
	Low24h string `json:"low_24h"`
}

type SnapshotResponse struct {
	Type string `json:"type"`
	Time string `json:"time"`
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
}

type L2UpdateResponse struct {
	Type string `json:"type"`
	Time string `json:"time"`
	Price string `json:"price"`
	Changes [][]string `json:"changes"`
}

type WSParser struct {
	Endpoint string
	conn *websocket.Conn
	writer *csv.Writer
	fd *os.File
	header []string
	output string
	subtype string
	raw bool
}

func NewWSParser(endpoint string, output string, subtype string, raw bool) (*WSParser) {
	p := &WSParser{endpoint, nil, nil, nil, nil, output, subtype, raw}

	d := websocket.Dialer{TLSClientConfig: &tls.Config{RootCAs: nil, InsecureSkipVerify: true}, HandshakeTimeout: 10*time.Second}
	conn, _, err := d.Dial(p.Endpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("NewWSParser: connected to endpoint=%v subtype=%v\n", endpoint, subtype)
	p.conn = conn
	return p
}

func (p *WSParser) Subscribe(products []string) {
	var req interface{}
	switch p.subtype {
	case "trades":
		p.header = []string{"time", "trade_id", "type", "product_id", "last_size", "price", "side", "best_bid", "best_ask",
			"volume_24h", "volume_30d", "open_24h", "low_24h", "high_24h"}
		req = SubscribeReq{
			Type: "subscribe",
			ProductIds: products,
			Channels: []string{channelTicker, channelHeartbeat},
		}

	case "quotes_trades":
		p.header = []string{"time", "trade_id", "type", "product_id", "last_size", "price", "side", "best_bid", "best_ask",
			"volume_24h", "volume_30d", "open_24h", "low_24h", "high_24h"}
		req = SubscribeReq{
			Type: "subscribe",
			ProductIds: products,
			Channels: []string{channelTicker, channelLevel2, channelMatches, channelHeartbeat},
		}
	default:
		log.Fatalf("unknown subscription type=%v", p.subtype)
	}

	buf, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}

	err = p.conn.WriteMessage(websocket.TextMessage, buf)
	if err != nil {
		log.Fatal(err)
	}
}

func (p *WSParser) createOutputWriter() {
	if len(p.output)==0 {
		t := time.Now()
		y,m,d := t.Date()
		if p.raw {
			p.output = fmt.Sprintf("%04d%02d%02d.%s.json", y, m, d, p.subtype)
		} else {
			p.output = fmt.Sprintf("%04d%02d%02d.%s.csv", y, m, d, p.subtype)
		}
	}

	log.Printf("createOutputWriter: output=%v\n", p.output)
	fd, err := os.OpenFile(p.output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("cannot open output file=%v: %v", p.output, err)
	}

	if p.raw {
		p.fd = fd
		return
	}

	p.writer = csv.NewWriter(fd)
	offset, err := fd.Seek(0, io.SeekCurrent)
	if err != nil {
		log.Fatal("cannot ftell:", err)
	}
	if offset>0 {
		p.writer.Write(p.header)
	}
}

func (p *WSParser) parsePayload(msg []byte) {
	header := GenericResponse{}
	err := json.Unmarshal(msg, &header)
	if err != nil {
		log.Print("parsePayload: parse error during json unmarshal");
		return
	}

	switch header.Type {
	case "ticker":
		resp := TickerResponse{}
		err := json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("tickerResponse: unable to parse: ", err)
		}
		log.Printf("received ticker=%v\n", resp)

	case "match":
		resp := MatchResponse{}
		err := json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("matchResponse: unable to parse: ", err)
		}
		log.Printf("match: %v\n", resp)

	case "subscriptions":
		resp := SubscribeResponse{}
		err := json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("subscribeResponse: unable to parse: ", err)
		}
		log.Printf("subscribe: %v\n", resp.Channels)

	case "error":
		resp := ErrorResponse{}
		err := json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("errorResponse: unable to parse: ", err)
		}
		log.Printf("received error=%v reason=%v\n", resp.Message, resp.Reason)

	case "heartbeat":
		resp := HeartbeatResponse{}
		err := json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("heartbeatResponse: unable to parse: ", err)
		}
		log.Printf("received heartbeat=%v\n", resp)

	case "l2update":
		resp := L2UpdateResponse{}
		err := json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("l2updateResponse: unable to parse: ", err)
		}
		log.Printf("received l2update=%v\n", resp)

	case "snapshot":
		resp := SnapshotResponse{}
		err := json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("snapshotResponse: unable to parse: ", err)
		}
		log.Printf("received snapshot=%v\n", resp)

	default:
		log.Printf("unhandled message type=%v\n", header.Type)
	}
}

func (p *WSParser) commitRaw(ts time.Time, msg []byte) {
	header, err := ts.MarshalText()
	if err != nil {
		log.Fatal("commitRaw: cannot marshal timestamp")
	}
	header = append(header, ' ')
	p.fd.Write(header)

	msg = append(msg, '\n')
	p.fd.Write(msg)
}

func (p *WSParser) captureMessage(ts time.Time, message []byte) {
	if p.writer == nil && p.fd == nil {
		p.createOutputWriter()
	}

	if p.raw {
		p.commitRaw(ts, message)
	} else {
		p.parsePayload(message)
	}
}

func (p *WSParser) handleRead() {
	messageType, message, err := p.conn.ReadMessage()
	if err != nil {
		log.Fatal(err)
	}

	ts := time.Now()

	if messageType == websocket.PingMessage {
		p.conn.WriteMessage(websocket.PongMessage, []byte(time.Now().String()))
	}

	if messageType == websocket.TextMessage {
		p.captureMessage(ts, message)
	}

	if messageType == websocket.BinaryMessage {
		log.Print("handleRead: ignoring binary message")
	}
}

func (p *WSParser) Run() {
	defer func() {
		p.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		p.conn.Close()
	}()

	for {
		p.handleRead()
	}
}

func (p *WSParser) Close() {
	p.conn.Close()
}
