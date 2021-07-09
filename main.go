// Copyright (c) 2021 Curvegrid Inc.

package main

import (
	"context"
	"log"
	"math/big"
	"time"

	"github.com/curvegrid/gofig"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func SubscribeTest(client *ethclient.Client, timeout time.Duration, address common.Address) {
	// setup timeout
	ctx, cancelTimeout := context.WithTimeout(context.Background(), timeout)

	// setup filter query
	query := ethereum.FilterQuery{
		Addresses: []common.Address{
			address,
		},
	}

	// setup logs
	ethLogs := make(chan types.Log)

	// attempt to subscribe
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

func GetLogsTest(client *ethclient.Client, timeout time.Duration, address common.Address, limit int64) {
	offset := int64(0)

	for {
		// setup filter query
		query := ethereum.FilterQuery{
			Addresses: []common.Address{
				address,
			},
			FromBlock: big.NewInt(offset),
			ToBlock:   big.NewInt(offset + limit),
		}

		log.Printf("Query: %+v", query)

		// setup timeout
		ctx, cancelTimeout := context.WithTimeout(context.Background(), timeout)

		// attempt to get logs
		ethLogs, err := client.FilterLogs(ctx, query)
		if err != nil {
			log.Printf("ERROR: %+v", err)
			cancelTimeout()
			continue
		}

		// we successfully got the logs, so we can cancel the timeout
		cancelTimeout()

		log.Printf("Logs received (%d): %+v", len(ethLogs), ethLogs)

		offset += limit
	}
}

type Config struct {
	Endpoint  string         `desc:"websocket endpoint to connect to"`
	Address   string         `desc:"Ethereum address to subscribe to events for"`
	Timeout   gofig.Duration `desc:"subscribe timeout"`
	Subscribe bool           `desc:"run the subscribe test"`
	GetLogs   bool           `desc:"run the get logs test"`
	Limit     int64          `desc:"maximum number of logs to retrieve at once"`
}

func main() {
	// config from environment (ES_ENDPOINT) or config file
	cfg := Config{
		Address: "0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c", // WBNB ERC-20 token
		Timeout: gofig.Duration(60 * time.Second),
		Limit:   2000,
	}
	gofig.SetEnvPrefix("ES")
	gofig.SetConfigFileFlag("c", "config file")
	gofig.AddConfigFile("config")
	gofig.Parse(&cfg)

	// setup subscription
	client, err := ethclient.DialContext(context.Background(), cfg.Endpoint)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected")

	if cfg.Subscribe && cfg.GetLogs {
		log.Fatal("only supports running a single test")
	}

	if cfg.Subscribe {
		SubscribeTest(client, time.Duration(cfg.Timeout), common.HexToAddress(cfg.Address))
	} else if cfg.GetLogs {
		GetLogsTest(client, time.Duration(cfg.Timeout), common.HexToAddress(cfg.Address), cfg.Limit)
	} else {
		log.Print("nothing to do")
	}
}
