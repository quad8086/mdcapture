package ws_parser

import (
	"os"
	"io"
	"fmt"
	"log"
	"time"
	"strconv"
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
			Channels: []string{channelTicker, channelLevel2, channelHeartbeat},
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

type Datum map[string]string

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

func (p *WSParser) commitRaw(msg []byte) {
	msg = append(msg, '\n')
	p.fd.Write(msg)
}

func (p *WSParser) commitDatum(d Datum) {
	var record []string
	for _, k := range p.header {
		record = append(record, d[k])
	}
	p.writer.Write(record)
	p.writer.Flush()
}

func toDatum(m map[string]interface{}) Datum {
	res := make(map[string]string)
	for k,v := range m {
		switch vv := v.(type) {
		case string:
			res[k] = vv
		case float64:
			res[k] = strconv.FormatFloat(vv, 'f', 2, 64)
		}
	}
	return res
}

func (p *WSParser) parsePayload(msg []byte) Datum {
	var v map[string]interface{}
	err := json.Unmarshal(msg, &v)
	if err != nil {
		log.Print("parsePayload: parse error during json unmarshal");
		return nil
	}

	var type_ interface{}
	var ok bool
	if type_, ok = v["type"]; !ok {
		log.Print("json payload doesnt have type field")
		return nil
	}

	switch type_ {
	case "ticker":
		return toDatum(v)
	case "subscriptions":
		channels := v["channels"].([]interface{})
		for _, channel := range channels {
			ch := channel.(map[string]interface{})
			products := ch["product_ids"].([]interface{})
			if len(products)==0 {
				log.Fatal("subscribe: no products returned")
			}
			log.Printf("subscribe: channel=%v products=%v\n", ch["name"], products)
		}

	case "heartbeat":
		log.Printf("heartbeat")
	default:
		fmt.Printf("unhandled message type=%v\n", type_)
	}

	return nil
}

func (p *WSParser) captureMessage(message []byte) {
	if p.writer == nil && p.fd == nil {
		p.createOutputWriter()
	}

	if p.raw {
		p.commitRaw(message)
		return
	}

	res := p.parsePayload(message)
	if res != nil {
		p.commitDatum(res)
	}
}

func (p *WSParser) handleRead() {
	messageType, message, err := p.conn.ReadMessage()
	if err != nil {
		log.Fatal(err)
	}

	if messageType == websocket.PingMessage {
		p.conn.WriteMessage(websocket.PongMessage, []byte(time.Now().String()))
	}

	if messageType == websocket.TextMessage {
		p.captureMessage(message)
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
