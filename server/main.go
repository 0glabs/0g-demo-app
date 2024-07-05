package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/0glabs/0g-storage-client/common/blockchain"
	"github.com/0glabs/0g-storage-client/contract"
	"github.com/0glabs/0g-storage-client/core"
	"github.com/0glabs/0g-storage-client/node"
	"github.com/0glabs/0g-storage-client/transfer"

	"server/store"
)

type DownloadArgs struct {
	Key  string `json:"key"`
	Root string `json:"root"`
}

type uploadArgs struct {
	file string
	tags string

	url      string
	contract string
	key      string

	force    bool
	taskSize uint
}

func getContract() *store.Store {

	client, err := ethclient.Dial("https://rpc-testnet.0g.ai")
	if err != nil {
		log.Fatal(err)
	}

	privateKey, err := crypto.HexToECDSA(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(16600))
	if err != nil {
		log.Fatal(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address := common.HexToAddress(os.Getenv("CONTRACT_ADDRESS"))
	instance, err := store.NewStore(address, client)
	if err != nil {
		log.Fatal(err)
	}
	return instance
}

func uploadFile(client *node.Client) {

	args := uploadArgs{
		file:     "test.txt",
		tags:     "0x",
		url:      "https://rpc-testnet.0g.ai",
		contract: "0x8873cc79c5b3b5666535C825205C9a128B1D75F1",
		key:      os.Getenv("PRIVATE_KEY"),
		force:    false,
		taskSize: 10,
	}
	w3client := blockchain.MustNewWeb3(args.url, args.key)
	defer w3client.Close()
	contractAddr := common.HexToAddress(args.contract)
	flow, err := contract.NewFlowContract(contractAddr, w3client)
	if err != nil {
		log.Fatal(err)
	}

	uploader, err := transfer.NewUploader(flow, []*node.Client{client})
	if err != nil {
		log.Fatal(err)
	}

	opt := transfer.UploadOption{
		Tags:     hexutil.MustDecode(args.tags),
		Force:    args.force,
		TaskSize: args.taskSize,
	}

	file, err := core.Open(args.file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if err := uploader.Upload(file, opt); err != nil {
		log.Fatal(err)
	}

	log.Println("Upload success")
}

func downloadFile(client *node.Client, root string) {
	downloader, err := transfer.NewDownloader(client)
	if err != nil {
		log.Fatal(err)
	}

	err = downloader.Download(root, "test.txt-downloaded", false)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Download success")
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ip := "https://rpc-storage.0g.ai"
	client, err := node.NewClient(ip)
	if err != nil {
		fmt.Println("Error creating client:", err)
	}

	instance := getContract()

	http.HandleFunc("/upload", uploadWrapper(client, instance))
	http.HandleFunc("/download", downloadWrapper(client, instance))

	err = http.ListenAndServe(":3333", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func uploadWrapper(client *node.Client, instance *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PostFormValue("key")
		if allowed, err := instance.GetItem(nil, key); err == nil {
			if allowed {
				uploadFile(client)
			} else {
				log.Println("Key not allowed")
			}
		}
	}
}

func downloadWrapper(client *node.Client, instance *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var d DownloadArgs

		// Decode the JSON body
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&d); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if allowed, err := instance.GetItem(nil, d.Key); err == nil {
			if allowed {
				downloadFile(client, d.Root)
			} else {
				log.Println("Key not allowed", d.Key)
			}
		}
	}
}
