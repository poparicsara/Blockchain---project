/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func main() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			fmt.Printf("Failed to populate wallet contents: %s\n", err)
			os.Exit(1)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		os.Exit(1)
	}

	contract := network.GetContract("cars")

	// result, err := contract.EvaluateTransaction("GetAllCars")
	// if err != nil {
	// 	log.Fatalf("Failed to evaluate transaction: %v", err)
	// }
	// log.Println(string(result))

	// result, err = contract.SubmitTransaction("createCar", "CAR10", "VW", "Polo", "Grey", "Mary", "3")
	// if err != nil {
	// 	fmt.Printf("Failed to submit transaction: %s\n", err)
	// 	os.Exit(1)
	// }
	// fmt.Println(string(result))

	fmt.Println("-------------- GET CAR BY ID --------------")
	result, err := contract.EvaluateTransaction("getCarById", "5")
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(result))

	fmt.Println("-------------- GET OWNER BY ID --------------")
	result, err = contract.EvaluateTransaction("getOwnerById", "OWNER1")
	if err != nil {
		fmt.Printf("Failed to evaluate transaction: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(result))

	fmt.Println("-------------- GET CARS BY COLOR --------------")
	result, err = contract.EvaluateTransaction("getCarsByColor", "blue")
	if err != nil {
		fmt.Println(fmt.Errorf("failed to evaluate transaction: %w", err))
		return
	}
	fmt.Println(string(result))

	// fmt.Println("-------------- REPAIR CAR --------------")
	// _, err = contract.SubmitTransaction("repairCar", "1")
	// if err != nil {
	// 	fmt.Println(fmt.Errorf("failed to submit transaction: %w", err))
	// 	return
	// }

	// _, err = contract.SubmitTransaction("changeCarColor", "CAR5", "blue")
	// if err != nil {
	// 	fmt.Printf("Failed to submit transaction: %s\n", err)
	// 	os.Exit(1)
	// }
}

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		"..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	err = wallet.Put("appUser", identity)
	if err != nil {
		return err
	}
	return nil
}