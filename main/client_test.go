package main

import (
	"context"
	"github.com/davidkhala/fabric-common/golang"
	"github.com/davidkhala/fabric-server-go/client"
	"github.com/davidkhala/fabric-server-go/model"
	"github.com/davidkhala/goutils"
	"github.com/davidkhala/goutils/http"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/kortschak/utter"
	"github.com/stretchr/testify/assert"
	rawHttp "net/http"
	"net/url"
	"testing"
)

var cryptoConfig = golang.CryptoConfig{
	MSPID:    "astriMSP",
	PrivKey:  golang.FindKeyFilesOrPanic(goutils.HomeResolve("delphi-fabric/config/ca-crypto-config/peerOrganizations/astri.org/users/Admin@astri.org/msp/keystore"))[0],
	SignCert: goutils.HomeResolve("delphi-fabric/config/ca-crypto-config/peerOrganizations/astri.org/users/Admin@astri.org/msp/signcerts/Admin@astri.org-cert.pem"),
}

// client side cache
var channel = "allchannel"
var endorsers = []model.Node{
	{
		Address:               "localhost:8051",
		TLSCARoot:             string(client.ReadPEMFile(goutils.HomeResolve("delphi-fabric/config/ca-crypto-config/peerOrganizations/icdd/tlsca/tlsca.icdd-cert.pem"))),
		SslTargetNameOverride: "peer0.icdd",
	},
	{
		Address:               "localhost:7051",
		TLSCARoot:             string(client.ReadPEMFile(goutils.HomeResolve("delphi-fabric/config/ca-crypto-config/peerOrganizations/astri.org/peers/peer0.astri.org/tls/ca.crt"))),
		SslTargetNameOverride: "peer0.astri.org",
	},
}

func postProposal(t *testing.T, result model.CreateProposalResult, signer *golang.Crypto) {
	var signedBytes = client.GetProposalSigned(result.Proposal, signer)
	var transactionBytes = client.CommitProposalAndSign(result.Proposal, signedBytes, endorsers, *signer)
	var orderer = model.Node{
		Address:               "localhost:7050",
		TLSCARoot:             string(client.ReadPEMFile(goutils.HomeResolve("delphi-fabric/config/ca-crypto-config/ordererOrganizations/hyperledger/orderers/orderer0.hyperledger/tls/ca.crt"))),
		SslTargetNameOverride: "orderer0.hyperledger",
	}
	println("...before commit transactionBytes")
	var status = client.Commit(orderer, transactionBytes)
	assert.Equal(t, common.Status_SUCCESS.String(), status)
	waitForTx(result.Txid)
}
func waitForTx(txid string) {
	var eventer = client.EventerFrom(endorsers[0])
	var signer = golang.LoadCryptoFrom(cryptoConfig)
	var txStatus = eventer.WaitForTx(channel, txid, signer)
	goutils.AssertOK(txStatus == peer.TxValidationCode_VALID.String(), txid+" is invalid")
}
func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		main()
	}(ctx)
	m.Run()
	cancel()
}
func TestPing(t *testing.T) {
	var _url = client.BuildURL("/fabric/ping")
	var _path = goutils.HomeResolve("delphi-fabric/config/ca-crypto-config/peerOrganizations/icdd/tlsca/tlsca.icdd-cert.pem")
	var Certificate = model.BytesPacked(client.ReadPEMFile(_path))

	var body = url.Values{
		"address":                  {"localhost:8051"},
		"certificate":              {Certificate},
		"ssl-target-name-override": {"peer0.icdd"},
	}

	response, err := rawHttp.PostForm(_url, body)
	goutils.PanicError(err)
	utter.Dump(response.Status)
}
func TestTransaction(t *testing.T) {
	var signer = golang.LoadCryptoFrom(cryptoConfig)
	var chaincode = "contracts"
	var args = []string{"StupidContract:ping"}

	// build http
	var body = url.Values{
		"creator":   {model.BytesPacked(signer.Creator)},
		"channel":   {channel},
		"chaincode": {chaincode},
		"args":      {string(goutils.ToJson(args))},
	}
	var _url = client.BuildURL("/fabric/create-proposal")
	var response = http.PostForm(_url, body, nil)
	var result = model.CreateProposalResult{}
	goutils.FromJson(response.BodyBytes(), &result)
	utter.Dump(result)
	postProposal(t, result, signer)
}
func TestQuery(t *testing.T) {
	var signer = golang.LoadCryptoFrom(cryptoConfig)
	var chaincode = "contracts"
	var args = []string{"SmartContract:who"}
	var body = url.Values{
		"creator":   {model.BytesPacked(signer.Creator)},
		"channel":   {channel},
		"chaincode": {chaincode},
		"args":      {string(goutils.ToJson(args))},
	}

	var _url = client.BuildURL("/fabric/create-proposal")
	var response = http.PostForm(_url, body, nil)
	var result = model.CreateProposalResult{}
	goutils.FromJson(response.BodyBytes(), &result)
	utter.Dump(result)

	// phase 2: QueryProposal
	var signedBytes = client.GetProposalSigned(result.Proposal, signer)
	var queryResult = client.QueryProposal(result.Proposal, signedBytes, endorsers)
	println(queryResult)
}
