package ws_parser

import (
	"log"
	"time"
	"strconv"
	"encoding/csv"
	"encoding/json"
	"crypto/tls"
	"github.com/gorilla/websocket"
)

type WSParser struct {
	endpoint string
	conn *websocket.Conn
	writer *csv.Writer
	channels []string
	raw bool
	committer *Committer
	headers map[string][]string
	start_ts time.Time
	last_ts time.Time
	show_status bool
	products []string
}

func NewWSParser(endpoint string, subtype string, raw bool, directory string, show_status bool, products []string) (*WSParser) {
	if len(products) == 0 {
		log.Fatal("NewWSParser: no products specified")
	}

	// figure out channels
	var channels []string
	switch subtype {
	case "trades":
		channels = []string{channelTicker}

	case "quotes_trades":
		channels = []string{channelTicker, channelMatches, channelLevel2}

	default:
		log.Fatalf("unknown subscription type=%v\n", subtype)
	}

	c := NewCommitter(directory)
	p := &WSParser{endpoint, nil, nil, channels, raw, c, make(map[string][]string), time.Now(), time.Time{}, show_status, products}

	c.RegisterTable("ticker", []string{"type", "recv_ts", "time", "product_id", "sequence", "qty", "price", "side", "trade_id",
		"best_bid", "best_ask", "open_24h", "low_24h", "high_24h", "volume_24h", "volume_30d"})
	c.RegisterTable("level", []string{"type", "recv_ts", "time", "product_id", "side", "price", "qty"})
	c.RegisterTable("match", []string{"type", "recv_ts", "time", "product_id", "trade_id", "side", "qty", "price", "sequence",
		"maker_id", "taker_id"})

	return p
}

func (p *WSParser) ConnectSubscribe() {
	if p.conn != nil {
		log.Printf("WSParser: connect: already connected\n")
		return
	}

	// connect
	d := websocket.Dialer{TLSClientConfig: &tls.Config{RootCAs: nil, InsecureSkipVerify: true}, HandshakeTimeout: 10*time.Second}
	conn, _, err := d.Dial(p.endpoint, nil)
	if err != nil {
		log.Printf("ConnectSubscribe: error connecting: %v\n", err);
		return
	}

	log.Printf("NewWSParser: now connected to endpoint=%v channels=%v\n", p.endpoint, p.channels)
	p.conn = conn

	// subscribe
	var req SubscribeReq
	req = SubscribeReq{
		Type: "subscribe",
		ProductIds: p.products,
		Channels: p.channels,
	}

	log.Printf("subscribe products=%v channels=%v\n", p.products, p.channels)
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

	const layout = "20060102-150405.000000"
	s_recv_ts := ts.Format(layout)

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
	case "last_match":
		resp := MatchResponse{}
		err = json.Unmarshal(msg, &resp)
		if err != nil {
			log.Fatal("response: unable to parse: ", err)
		}
		record := []string{resp.Type, s_recv_ts, resp.Time, resp.ProductID, strconv.FormatInt(resp.TradeID, 10),
			resp.Side, resp.Size, resp.Price,
			strconv.FormatInt(resp.Sequence, 10), resp.MakerOrderID, resp.TakerOrderID}
		p.committer.CommitRecord(ts, "match", record)

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
		log.Printf("handleRead: error: %v\n", err)
		p.Disconnect()
	}

	recv_ts := time.Now()

	if messageType == websocket.PingMessage {
		p.conn.WriteMessage(websocket.PongMessage, []byte(recv_ts.String()))
	}

	if messageType == websocket.TextMessage {
		p.captureMessage(recv_ts, message)
	}

	if messageType == websocket.BinaryMessage {
		log.Print("handleRead: ignoring binary message")
	}
}

func (p *WSParser) Run() {
	defer func() {
		p.Close()
	}()

	var report_ts time.Time = time.Now()
	var connection_attempts int64 = 0
	for {
		p.last_ts = time.Now()
		if p.last_ts.Day() != p.start_ts.Day() {
			log.Printf("committer: exiting normally on day change\n")
			break
		}

		if p.conn == nil {
			time.Sleep(time.Duration(connection_attempts)*time.Second)
			p.ConnectSubscribe()
			connection_attempts++
			continue
		}

		if p.show_status && p.last_ts.Sub(report_ts).Seconds() > 5 {
			log.Printf("committer: %v\n", p.committer.Status())
			report_ts = p.last_ts
		}

		p.handleRead()
	}
}

func (p *WSParser) Disconnect() {
	if p.conn == nil {
		return
	}

	log.Printf("disconnecting\n")
	//p.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	p.conn.Close()
	p.conn = nil
}

func (p *WSParser) Close() {
	p.Disconnect()
	p.committer.Close()
}
