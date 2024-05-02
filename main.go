package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fatih/color"

	"github.com/ethereum/go-ethereum/common"
	"github.com/joho/godotenv"
	pancakeFactory "github.com/nikola43/liquiditysniperbot/contracts/IPancakeFactory"
	pancakePair "github.com/nikola43/liquiditysniperbot/contracts/IPancakePair"
	"github.com/nikola43/web3golanghelper/web3helper"
)

type Reserve struct {
	Reserve0           *big.Int
	Reserve1           *big.Int
	BlockTimestampLast uint32
}

func main() {

	//yellow := color.New(color.FgYellow).SprintFunc()
	// red := color.New(color.FgRed).SprintFunc()
	// green := color.New(color.FgGreen).SprintFunc()
	// cyan := color.New(color.FgCyan).SprintFunc()

	// read .env variables
	RPC_URL, WS_URL, WETH_ADDRESS, FACTORY_ADDRESS, TOKEN_ADDRESS, PK, BUY_AMOUNT, ROUTER_ADDRESS := readEnvVariables()

	web3GolangHelper := initWeb3(RPC_URL, WS_URL)
	fromAddress := GeneratePublicAddressFromPrivateKey(PK)

	// convert buy amount to float

	// infinite loop
	for {
		// get pair address
		lpPairAddress := getPair(web3GolangHelper, WETH_ADDRESS, FACTORY_ADDRESS, TOKEN_ADDRESS)
		fmt.Println("LP Pair Address: " + lpPairAddress)

		if lpPairAddress != "0x0000000000000000000000000000000000000000" {
			reserves := getReserves(web3GolangHelper, lpPairAddress)

			fmt.Println("Reserve0: " + reserves.Reserve0.String())
			fmt.Println("Reserve1: " + reserves.Reserve1.String())

			// check if reserves is greater than 0
			if reserves.Reserve0.Cmp(big.NewInt(0)) > 0 && reserves.Reserve1.Cmp(big.NewInt(0)) > 0 {
				buyAmount, err := strconv.ParseFloat(BUY_AMOUNT, 32)
				if err != nil {
					fmt.Println(err)
				}
				web3GolangHelper.BuyV3(ROUTER_ADDRESS, WETH_ADDRESS, PK, fromAddress, TOKEN_ADDRESS, buyAmount)
				fmt.Println("Press enter to exit")
				fmt.Scanln()
				os.Exit(0)
			}
		}
		// sleep 1 second
		time.Sleep(1 * time.Millisecond)
	}

}

func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

// function for read .env variables
func readEnvVariables() (string, string, string, string, string, string, string, string) {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	red := color.New(color.FgRed).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	RPC_URL := os.Getenv("RPC_URL")
	WS_URL := os.Getenv("WS_URL")
	WETH_ADDRESS := os.Getenv("WETH_ADDRESS")
	FACTORY_ADDRESS := os.Getenv("FACTORY_ADDRESS")
	TOKEN_ADDRESS := os.Getenv("TOKEN_ADDRESS")
	PK := os.Getenv("PK")
	BUY_AMOUNT := os.Getenv("BUY_AMOUNT")
	ROUTER_ADDRESS := os.Getenv("ROUTER_ADDRESS")

	if RPC_URL == "" {
		fmt.Printf("%s\n", red("Error: RPC_URL is not set in .env"))
		os.Exit(1)
	}
	if WS_URL == "" {
		fmt.Printf("%s\n", red("Error: WS_URL is not set in .env"))
		os.Exit(1)
	}
	if WETH_ADDRESS == "" {
		fmt.Printf("%s\n", red("Error: WETH_ADDRESS is not set in .env"))
		os.Exit(1)
	}
	if FACTORY_ADDRESS == "" {
		fmt.Printf("%s\n", red("Error: FACTORY_ADDRESS is not set in .env"))
		os.Exit(1)
	}
	if TOKEN_ADDRESS == "" {
		fmt.Printf("%s\n", red("Error: TOKEN_ADDRESS is not set in .env"))
		os.Exit(1)
	}
	if PK == "" {
		fmt.Printf("%s\n", red("Error: PK is not set in .env"))
		os.Exit(1)
	}
	if BUY_AMOUNT == "" {
		fmt.Printf("%s\n", red("Error: BUY_AMOUNT is not set in .env"))
		os.Exit(1)
	}
	if ROUTER_ADDRESS == "" {
		fmt.Printf("%s\n", red("Error: ROUTER_ADDRESS is not set in .env"))
		os.Exit(1)
	}

	fmt.Printf("%s\n\n", cyan("CONFIG"))
	fmt.Printf("%s: %s\n", cyan("RPC_URL"), yellow(RPC_URL))
	fmt.Printf("%s: %s\n", cyan("WS_URL"), yellow(WS_URL))
	fmt.Printf("%s: %s\n", cyan("WETH_ADDRESS"), yellow(WETH_ADDRESS))
	fmt.Printf("%s: %s\n", cyan("FACTORY_ADDRESS"), yellow(FACTORY_ADDRESS))
	fmt.Printf("%s: %s\n", cyan("TOKEN_ADDRESS"), yellow(TOKEN_ADDRESS))
	fmt.Printf("%s: %s\n", cyan("BUY_AMOUNT"), yellow(BUY_AMOUNT))
	fmt.Printf("%s: %s\n", cyan("ROUTER_ADDRESS"), yellow(ROUTER_ADDRESS))

	return RPC_URL, WS_URL, WETH_ADDRESS, FACTORY_ADDRESS, TOKEN_ADDRESS, PK, BUY_AMOUNT, ROUTER_ADDRESS
}

func initWeb3(rpcUrl, wsUrl string) *web3helper.Web3GolangHelper {
	web3GolangHelper := web3helper.NewWeb3GolangHelper(rpcUrl, wsUrl)

	chainID, err := web3GolangHelper.HttpClient().NetworkID(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Chain Id: " + chainID.String())
	return web3GolangHelper
}

func getReserves(web3GolangHelper *web3helper.Web3GolangHelper, pairAddress string) Reserve {

	pairInstance, instanceErr := pancakePair.NewPancake(common.HexToAddress(pairAddress), web3GolangHelper.HttpClient())
	if instanceErr != nil {
		fmt.Println(instanceErr)
	}

	reserves, getReservesErr := pairInstance.GetReserves(nil)
	if getReservesErr != nil {
		fmt.Println(getReservesErr)
	}

	return reserves
}

func getPair(web3GolangHelper *web3helper.Web3GolangHelper, wethAddress, factoryAddress, tokenAddress string) string {

	factoryInstance, instanceErr := pancakeFactory.NewPancake(common.HexToAddress(factoryAddress), web3GolangHelper.HttpClient())
	if instanceErr != nil {
		fmt.Println(instanceErr)
	}

	lpPairAddress, getPairErr := factoryInstance.GetPair(nil, common.HexToAddress(wethAddress), common.HexToAddress(tokenAddress))
	if getPairErr != nil {
		fmt.Println(getPairErr)
	}

	return lpPairAddress.Hex()

}

func GeneratePublicAddressFromPrivateKey(plainPrivateKey string) common.Address {
	privateKey, err := crypto.HexToECDSA(plainPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return fromAddress
}
