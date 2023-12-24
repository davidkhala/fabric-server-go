package client

import (
	"github.com/davidkhala/fabric-common/golang"
	"github.com/davidkhala/fabric-server-go/model"
	"github.com/davidkhala/goutils"
	"github.com/davidkhala/goutils/http"
	"github.com/kortschak/utter"
	"net/url"
	"testing"
)

var cryptoConfig = golang.CryptoConfig{
	MSPID:    "astriMSP",
	PrivKey:  golang.FindKeyFilesOrPanic("/home/david/delphi-fabric/config/ca-crypto-config/peerOrganizations/astri.org/users/Admin@astri.org/msp/keystore")[0],
	SignCert: "/home/david/delphi-fabric/config/ca-crypto-config/peerOrganizations/astri.org/users/Admin@astri.org/msp/signcerts/Admin@astri.org-cert.pem",
}

// client side cache
var txid string
var channel = "allchannel"
var endorsers = []model.Node{
	{
		Address:               "localhost:8051",
		TLSCARoot:             string(ReadPEMFile("/home/david/delphi-fabric/config/ca-crypto-config/peerOrganizations/icdd/tlsca/tlsca.icdd-cert.pem")),
		SslTargetNameOverride: "peer0.icdd",
	},
	{
		Address:               "localhost:7051",
		TLSCARoot:             string(ReadPEMFile("/home/david/delphi-fabric/config/ca-crypto-config/peerOrganizations/astri.org/peers/peer0.astri.org/tls/ca.crt")),
		SslTargetNameOverride: "peer0.astri.org",
	},
}

func postProposal(result model.CreateProposalResult, signer *golang.Crypto) {
	var signedBytes = GetProposalSigned(result.Proposal, signer)
	txid = result.Txid
	var transactionBytes = CommitProposalAndSign(result.Proposal, signedBytes, endorsers, *signer)
	var orderer = model.Node{
		Address:               "localhost:7050",
		TLSCARoot:             string(ReadPEMFile("/home/david/delphi-fabric/config/ca-crypto-config/ordererOrganizations/hyperledger/orderers/orderer0.hyperledger/tls/ca.crt")),
		SslTargetNameOverride: "orderer0.hyperledger",
	}
	var status = Commit(orderer, transactionBytes)
	utter.Dump(status)
	waitForTx(txid)
}
func waitForTx(txid string) {
	var eventer = EventerFrom(endorsers[0])
	var signer = InitOrPanic(cryptoConfig)
	var txStatus = eventer.WaitForTx(channel, txid, signer)
	utter.Dump(txStatus)
}
func TestTransaction(t *testing.T) {
	var signer = InitOrPanic(cryptoConfig)
	var chaincode = "contracts"
	var args = []string{"SmartContract:who"}

	// build http
	var body = url.Values{
		"creator":   {model.BytesPacked(signer.Creator)},
		"channel":   {channel},
		"chaincode": {chaincode},
		"args":      {string(goutils.ToJson(args))},
	}
	var _url = BuildURL("/fabric/create-proposal")
	var response = http.PostForm(_url, body, nil)
	var result = model.CreateProposalResult{}
	goutils.FromJson(response.BodyBytes(), &result)
	utter.Dump(result)
	postProposal(result, signer)
}
func TestQuery(t *testing.T) {
	var signer = InitOrPanic(cryptoConfig)
	var chaincode = "contracts"
	var args = []string{"SmartContract:who"}
	var body = url.Values{
		"creator":   {model.BytesPacked(signer.Creator)},
		"channel":   {channel},
		"chaincode": {chaincode},
		"args":      {string(goutils.ToJson(args))},
	}

	var _url = BuildURL("/fabric/create-proposal")
	var response = http.PostForm(_url, body, nil)
	var result = model.CreateProposalResult{}
	goutils.FromJson(response.BodyBytes(), &result)
	utter.Dump(result)

	// phase 2: QueryProposal
	var signedBytes = GetProposalSigned(result.Proposal, signer)
	var queryResult = QueryProposal(result.Proposal, signedBytes, endorsers)
	println(queryResult)
}
func TestCreateToken(t *testing.T) {
	var signer = InitOrPanic(cryptoConfig)

	var body = url.Values{
		"creator": {model.BytesPacked(signer.Creator)},
		"channel": {channel},
		"owner":   {"david"},
		"content": {"github.com/delphi-fabric"},
	}
	var _url = BuildURL("/ecosystem/createToken")
	var response = http.PostForm(_url, body, nil)
	var result = model.CreateTokenResult{}
	goutils.FromJson(response.BodyBytes(), &result)
	utter.Dump("Token:" + result.Token)
	postProposal(result.CreateProposalResult, signer)
}
