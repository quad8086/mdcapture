package ws_parser

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

type BaseResponse struct {
	Type string `json:"type"`
}

type MatchResponse struct {
	Type string `json:"type"`
	Time string `json:"time"`
	Sequence int64 `json:"sequence"`
	TradeID int64 `json:"trade_id"`
	MakerOrderID string `json:"maker_order_id"`
	TakerOrderID string `json:"taker_order_id"`
	Side string `json:"side"`
	Size string `json:"size"`
	Price string `json:"price"`
	ProductID string `json:"product_id"`
}

type HeartbeatResponse struct {
	Type string `json:"type"`
	Time string `json:"time"`
	Sequence int64 `json:"sequence"`
	LastTradeID int64 `json:"last_trade_id"`
	ProductID string `json:"product_id"`
}

type ErrorResponse struct {
	BaseResponse
	Message string `json:"message"`
	Reason string `json:"reason"`
}

type ChannelResponse struct {
	Name string `json:"name"`
	ProductIDs []string `json:"product_ids"`
}

type SubscribeResponse struct {
	BaseResponse
	Channels []ChannelResponse `json:"channels"`
}

type TickerResponse struct {
	Type string `json:"type"`
	Time string `json:"time"`
	ProductID string `json:"product_id"`
	TradeID int64 `json:"trade_id"`
	Sequence int64 `json:"sequence"`
	Price string `json:"price"`
	Side string `json:"side"`
	Qty string `json:"last_size"`
	BestBid string `json:"best_bid"`
	BestAsk string `json:"best_ask"`
	Open24h string `json:"open_24h"`
	High24h string `json:"high_24h"`
	Low24h string `json:"low_24h"`
	Volume24h string `json:"volume_24h"`
	Volume30d string `json:"volume_30d"`
}

type SnapshotResponse struct {
	Type string `json:"type"`
	ProductID string `json:"product_id"`
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
}

type L2UpdateResponse struct {
	Type string `json:"type"`
	Time string `json:"time"`
	ProductID string `json:"product_id"`
	Changes [][]string `json:"changes"`
}
