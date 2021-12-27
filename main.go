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
		SubscriptionType string `short:"s" long:"subscription-type" description:"specify subscription type" default:"trades"`
		Output string `short:"o" long:"output" description:"override output filename"`
		Raw bool `short:"r" long:"raw" description:"capture raw json"`
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

	p := ws_parser.NewWSParser(opts.Endpoint, opts.Output, opts.SubscriptionType, opts.Raw)
	p.Subscribe(opts.Products)
	p.Run()
	os.Exit(0)
}
