package main

import (
	"os"
	"log"
	"github.com/jessevdk/go-flags"
	"mdcapture/ws_parser"
)

func main() {
	var opts struct {
		Endpoint string `short:"E" long:"endpoint" description:"override endpoint" default:"wss://ws-feed-public.sandbox.exchange.coinbase.com"`
		Products []string `short:"p" long:"products" description:"define products list"`
		SubscriptionType string `short:"s" long:"subscription-type" description:"subscription type (trades|quotes_trades)" default:"trades"`
		Raw bool `short:"r" long:"raw" description:"capture raw json"`
		Directory string `short:"D" long:"directory" description:"specify output directory; templates {y}, {m}, {d}, {ymd}"`
		Status bool `short:"S" long:"status" description:"periodically output status"`
	}

	log.SetPrefix("mdcapture ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lmsgprefix)

	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatal("options error")
	}

	if len(opts.Products)==0 {
		log.Fatal("please specify products")
	}

	p := ws_parser.NewWSParser(opts.Endpoint, opts.SubscriptionType, opts.Raw, opts.Directory, opts.Status)
	p.Subscribe(opts.Products)
	p.Run()
	os.Exit(0)
}
