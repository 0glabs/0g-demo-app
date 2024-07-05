# 0g-demo-app

This repo is to demonstrate a simple application that is built on top of 0G network.
This example involves three parts: contract, server, client

- A contract is responsible for recording the authority of clients (whether they are registered)
- A server is checking the authority of the request from a client and allows for file upload/download if it's registered
- A client should first interact with the contract to register himself and then upload/download files to the service provider (server).

1. Deploy a contract on [0G chain](https://0g.ai/).

In order to run, execute the following to install dependencies
```bash
cd server
npm install --save-dev hardhat
npm install
```

Then run the following code to compile the contract
```bash
npx hardhat compile
```

Before deploying, setup the environment variable with your own private key by creating a `.env` file and put
```bash
PRIVATE_KEY=<your_private_key>
```

Finally, run the deploy script to deploy the contract
```bash
npx run scripts/deploy.js --network zg
```

Record the deployed contract address.

2. Compile and run the server

```bash
go mod tidy
go build
```

Add the following field to the .env file
```bash
CONTRACT_ADDRESS=<deployed_contract_address>
```

then start service
```bash
./server
```

3. Run the client to register onchain

```bash
cd client
go mod tidy
go run main.go
```

You will receive a tx hash which is the transaction to set the key on chain.

4. Access the server and upload/download files

Run the following to upload the file

```bash
curl -X POST -F 'key=<your_key_in_client>' 'http://localhost:3333/upload'
```

And use the following to download the file. The data root can be found in the console when you upload the file.
```bash
curl -v -X POST -H 'Content-Type: application/json' -d '{"key":"<your_key_in_client>","root":"<data_root>"}'  'http://localhost:3333/download'
```

Refer to [this documentation](https://goethereumbook.org/smart-contract-read/) on how to generate .go code from .sol code.

