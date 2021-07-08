// Copyright (c) 2021 Curvegrid Inc.

package main

import (
	"context"
	"log"
	"time"

	"github.com/curvegrid/gofig"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Config struct {
	Endpoint string         `desc:"websocket endpoint to connect to"`
	Address  string         `desc:"Ethereum address to subscribe to events for"`
	Timeout  gofig.Duration `desc:"subscribe timeout"`
}

func main() {
	// config from environment (ES_ENDPOINT) or config file
	cfg := Config{
		Address: "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c", // WBNB ERC-20 token
		Timeout: gofig.Duration(5 * time.Second),
	}
	gofig.SetEnvPrefix("ES")
	gofig.SetConfigFileFlag("c", "config file")
	gofig.AddConfigFile("config")
	gofig.Parse(&cfg)

	// setup timeout
	ctx, cancelTimeout := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout))

	// setup subscription
	client, err := ethclient.DialContext(ctx, cfg.Endpoint)
	if err != nil {
		log.Fatal(err)
	}

	// setup filter query
	query := ethereum.FilterQuery{
		Addresses: []common.Address{
			common.HexToAddress(cfg.Address),
		},
	}

	// setup logs
	ethLogs := make(chan types.Log)

	// attempt to subscribe
	// HANGS HERE!
	subscription, err := client.SubscribeFilterLogs(ctx, query, ethLogs)
	if err != nil {
		log.Fatal(err)
	}
	defer subscription.Unsubscribe()

	// we successfully subscribed, so we can cancel the timeout
	cancelTimeout()

	log.Print("Subscription successful, waiting for logs")

	// subscription was successful, read logs and any errors from the channels
	// CTRL+C to quit
	for {
		select {
		case err := <-subscription.Err():
			// error returned from the subscription
			// for example, websocket connection was closed by the remote side
			log.Fatal(err)
		case ethLog := <-ethLogs:
			log.Printf("Msg received: %+v", ethLog)
		}
	}
}
