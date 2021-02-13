// Copyright (c) 2019 IoTeX
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iotexproject/iotex-address/address"
	"github.com/iotexproject/iotex-antenna-go/v2/iotex"
	"github.com/iotexproject/iotex-proto/golang/iotexapi"
	"go.uber.org/config"

	"github.com/iotexproject/ioTube/witness-service/db"
	"github.com/iotexproject/ioTube/witness-service/util"
	"github.com/iotexproject/ioTube/witness-service/witness"
)

// Configuration defines the configuration of the witness service
type Configuration struct {
	Chain                    string        `json:"chain" yaml:"chain"`
	ClientURL                string        `json:"clientURL" yaml:"clientURL"`
	PrivateKey               string        `json:"privateKey" yaml:"privateKey"`
	ValidatorContractAddress string        `json:"vialidatorContractAddress" yaml:"validatorContractAddress"`
	CashierContractAddress   string        `json:"cashierContractAddress" yaml:"cashierContractAddress"`
	StartBlockHeight         int           `json:"startBlockHeight" yaml:"startBlockHeight"`
	BatchSize                int           `json:"batchSize" yaml:"batchSize"`
	Interval                 time.Duration `json:"interval" yaml:"interval"`

	RelayerURL   string `json:"relayerURL" yaml:"relayerURL"`
	SlackWebHook string `json:"slackWebHook" yaml:"slackWebHook"`

	DatabaseURL        string `json:"databaseURL" yaml:"databaseURL"`
	TransferTableName  string `json:"transferTableName" yaml:"transferTableName"`
	TokenPairTableName string `json:"tokenPairTableName" yaml:"tokenPairTableName"`
}

var (
	defaultConfig = Configuration{
		Chain:                    "ethereum",
		Interval:                 time.Minute,
		RelayerURL:               "",
		StartBlockHeight:         9305000,
		BatchSize:                100,
		PrivateKey:               "",
		SlackWebHook:             "",
		ClientURL:                "",
		TransferTableName:        "",
		TokenPairTableName:       "",
		ValidatorContractAddress: "",
		CashierContractAddress:   "",
	}
)

var configFile = flag.String("config", "", "path of config file")

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "-config <filename>")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	opts := []config.YAMLOption{config.Static(defaultConfig), config.Expand(os.LookupEnv)}
	if *configFile != "" {
		opts = append(opts, config.File(*configFile))
	}
	yaml, err := config.NewYAML(opts...)
	if err != nil {
		log.Fatalln(err)
	}
	var cfg Configuration
	if err := yaml.Get(config.Root).Populate(&cfg); err != nil {
		log.Fatalln(err)
	}
	if height, ok := os.LookupEnv("WITNESS_START_BLOCK_HEIGHT"); ok {
		cfg.StartBlockHeight, err = strconv.Atoi(height)
		if err != nil {
			log.Fatalln(err)
		}
	}
	if relayerURL, ok := os.LookupEnv("WITNESS_RELAYER_URL"); ok {
		cfg.RelayerURL = relayerURL
	}
	if pk, ok := os.LookupEnv("WITNESS_PRIVATE_KEY"); ok {
		cfg.PrivateKey = pk
	}
	privateKey, err := crypto.HexToECDSA(cfg.PrivateKey)
	if err != nil {
		log.Fatalf("failed to decode private key %v\n", err)
	}
	// TODO: load more parameters from env
	if cfg.SlackWebHook != "" {
		util.SetSlackURL(cfg.SlackWebHook)
	}
	if chain, ok := os.LookupEnv("WITNESS_CHAIN"); ok {
		cfg.Chain = chain
	}
	if validatorAddr, ok := os.LookupEnv("WITNESS_VALIDATOR_CONTRACT_ADDRESS"); ok {
		cfg.ValidatorContractAddress = validatorAddr
	}
	if cashierAddr, ok := os.LookupEnv("WITNESS_CASHIER_CONTRACT_ADDRESS"); ok {
		cfg.CashierContractAddress = cashierAddr
	}
	if url, ok := os.LookupEnv("WITNESS_CLIENT_URL"); ok {
		cfg.ClientURL = url
	}
	var cashier witness.TokenCashier
	var validatorContractAddr common.Address
	switch cfg.Chain {
	case "iotex":
		cashierAddr, err := address.FromString(cfg.CashierContractAddress)
		if err != nil {
			log.Fatalf("failed to parse cashier address %v\n", err)
		}
		conn, err := iotex.NewDefaultGRPCConn(cfg.ClientURL)
		if err != nil {
			log.Fatalf("failed ot create grpc connection %v\n", err)
		}
		// defer conn.Close()
		if cashier, err = witness.NewTokenCashier(cashierAddr, iotex.NewReadOnlyClient(iotexapi.NewAPIServiceClient(conn))); err != nil {
			log.Fatalf("failed to create cashier %v\n", err)
		}
		validatorContractAddr = common.HexToAddress(cfg.ValidatorContractAddress)
	case "ethereum":
		ethClient, err := ethclient.DialContext(context.Background(), cfg.ClientURL)
		if err != nil {
			log.Fatal(err)
		}
		if cashier, err = witness.NewTokenCashierOnEthereum(common.HexToAddress(cfg.CashierContractAddress), ethClient, 12); err != nil {
			log.Fatalf("failed to create cashier %v\n", err)
		}
		addr, err := address.FromString(cfg.ValidatorContractAddress)
		if err != nil {
			log.Fatalf("failed to parse validator contract address %v\n", err)
		}
		validatorContractAddr = common.BytesToAddress(addr.Bytes())
	default:
		log.Fatalf("unknown chain name %s", cfg.Chain)
	}

	service, err := witness.NewService(
		cfg.RelayerURL,
		validatorContractAddr,
		cashier,
		witness.NewRecorder(
			db.NewStore("mysql", cfg.DatabaseURL),
			cfg.TransferTableName,
			cfg.TokenPairTableName,
		),
		privateKey,
		uint64(cfg.StartBlockHeight),
		uint16(cfg.BatchSize),
		cfg.Interval,
	)
	if err != nil {
		log.Fatalf("failed to create witness service: %v\n", err)
	}
	if err := service.Start(context.Background()); err != nil {
		log.Fatalf("failed to start witness service: %v\n", err)
	}
	defer service.Stop(context.Background())
	log.Println("Serving...")
	select {}
}