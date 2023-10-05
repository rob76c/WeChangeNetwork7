# WeChangeNetwork7


##Create a KIND cluster:
chmod +x network
./network kind
./network cluster init


##Launch the network, create a channel, and deploy the basic-asset-transfer smart contract:
./network up
./network channel create
./network chaincode deploy asset-transfer-basic ../asset-transfer-basic/chaincode-external


##Invoke and query chaincode:
./network chaincode invoke asset-transfer-basic '{"Args":["InitLedger"]}'
./network chaincode query  asset-transfer-basic '{"Args":["ReadAsset","asset1"]}'


##Access the blockchain with a REST API:
./network rest-easy


##Shut down the test network:
./network down 


##Tear down the cluster (KIND):
./network unkind
