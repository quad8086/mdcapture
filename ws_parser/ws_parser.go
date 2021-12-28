package ws_parser

import (
	"os"
	"log"
	"time"
	"strconv"
	"encoding/csv"
	"encoding/json"
	"crypto/tls"
	"github.com/gorilla/websocket"
)

type WSParser struct {
	Endpoint string
	conn *websocket.Conn
	writer *csv.Writer
	raw_fd *os.File
	output string
	subtype string
	raw bool
	committer *Committer
	headers map[string][]string
}

func NewWSParser(endpoint string, output string, subtype string, raw bool) (*WSParser) {
	c := NewCommitter()
	p := &WSParser{endpoint, nil, nil, nil, output, subtype, raw, c, make(map[string][]string)}

	c.RegisterTable("ticker", []string{"type", "recv_ts", "time", "product_id", "sequence", "qty", "price", "side", "trade_id",
		"best_bid", "best_ask", "open_24h", "low_24h", "high_24h", "volume_24h", "volume_30d"})
	c.RegisterTable("match", []string{"type", "recv_ts", "time", "product_id", "trade_id", "side", "size", "price"})
	c.RegisterTable("level", []string{"type", "recv_ts", "time", "product_id", "side", "price", "qty"})

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
		req = SubscribeReq{
			Type: "subscribe",
			ProductIds: products,
			Channels: []string{channelTicker, channelHeartbeat},
		}

	case "quotes_trades":
		req = SubscribeReq{
			Type: "subscribe",
			ProductIds: products,
			Channels: []string{channelTicker, channelLevel2, channelHeartbeat},
		}
	default:
		log.Fatalf("unknown subscription type=%v\n", p.subtype)
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

func (p *WSParser) parsePayload(ts time.Time, msg []byte) {
	header := BaseResponse{}
	err := json.Unmarshal(msg, &header)
	if err != nil {
		log.Print("parsePayload: parse error during json unmarshal");
		return
	}

	recv_ts, _ := ts.MarshalText()
	s_recv_ts := string(recv_ts)

	switch header.Type {
	case "ticker":
		resp := TickerResponse{}
		err = json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("response: unable to parse: ", err)
		}
		record := []string{resp.Type, s_recv_ts, resp.Time, resp.ProductID, strconv.FormatInt(resp.Sequence, 10), resp.Qty, resp.Price, resp.Side, strconv.FormatInt(resp.TradeID, 10),
			resp.BestBid, resp.BestAsk, resp.Open24h, resp.Low24h, resp.High24h, resp.Volume24h, resp.Volume30d}
		p.committer.CommitRecord(ts, header.Type, record)

	case "l2update":
		resp := L2UpdateResponse{}
		err = json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("response: unable to parse: ", err)
		}
		record := []string{resp.Type, s_recv_ts, resp.Time, resp.ProductID, resp.Changes[0][0], resp.Changes[0][1], resp.Changes[0][2]}
		p.committer.CommitRecord(ts, "level", record)

	case "snapshot":
		resp := SnapshotResponse{}
		err = json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("response: unable to parse: ", err)
		}
		for _,chg := range resp.Bids {
			record := []string{resp.Type, s_recv_ts, "", resp.ProductID, "buy", chg[0], chg[1]}
			p.committer.CommitRecord(ts, "level", record)
		}
		for _,chg := range resp.Asks {
			record := []string{resp.Type, s_recv_ts, "", resp.ProductID, "sell", chg[0], chg[1]}
			p.committer.CommitRecord(ts, "level", record)
		}

	case "match":
		resp := MatchResponse{}
		err = json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("response: unable to parse: ", err)
		}
		log.Printf("received type=%v response=%v\n", resp.Type, resp)

	case "subscriptions":
		resp := SubscribeResponse{}
		err = json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("response: unable to parse: ", err)
		}
		log.Printf("received type=%v response=%v\n", resp.Type, resp)

	case "error":
		resp := ErrorResponse{}
		err = json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("response: unable to parse: ", err)
		}
		log.Printf("received type=%v response=%v\n", resp.Type, resp)

	case "heartbeat":
		resp := HeartbeatResponse{}
		err = json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("response: unable to parse: ", err)
		}
		log.Printf("received type=%v response=%v\n", resp.Type, resp)

	default:
		log.Printf("unhandled message type=%v\n", header.Type)
		return
	}
}

func (p *WSParser) captureMessage(ts time.Time, message []byte) {
	if p.raw {
		p.committer.commitRawPayload(ts, message)
	} else {
		p.parsePayload(ts, message)
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
		p.Close()
	}()

	for {
		p.handleRead()
	}
}

func (p *WSParser) Close() {
	p.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	p.conn.Close()
	p.committer.Close()
}
