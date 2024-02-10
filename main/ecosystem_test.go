package main

import (
	"github.com/davidkhala/fabric-common/golang"
	"github.com/davidkhala/fabric-server-go/client"
	"github.com/davidkhala/fabric-server-go/model"
	"github.com/davidkhala/goutils"
	"github.com/davidkhala/goutils/http"
	"net/url"
	"testing"
)

func TestCreateToken(t *testing.T) {
	var signer = golang.LoadCryptoFrom(cryptoConfig)

	var body = url.Values{
		"creator": {model.BytesPacked(signer.Creator)},
		"channel": {channel},
		"owner":   {"david"},
	}
	var _url = client.BuildURL("/ecosystem/create-token")
	var response = http.PostForm(_url, body, nil)
	var result = model.CreateTokenResult{}
	goutils.FromJson(response.BodyBytes(), &result)
	println("Token:" + result.Token)
	postProposal(t, result.CreateProposalResult, signer)
}
