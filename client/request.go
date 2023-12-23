package client

// request.go is designed for http request towards server side
import (
	"github.com/davidkhala/fabric-server-go/model"
	"github.com/davidkhala/goutils"
	"github.com/davidkhala/goutils/http"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"net/url"
	"os"
)

func BuildURL(route string) string {
	baseUrl, found := os.LookupEnv("BASE_URL")
	if !found {
		baseUrl = "http://localhost:8080"
	}
	return baseUrl + route
}

func Propose(proposal string, signedBytes []byte, endorsers []model.Node) (proposalResponses []*peer.ProposalResponse, payload []byte) {
	// Send out
	var _url = BuildURL("/fabric/transact/process-proposal")

	for index, endorser := range endorsers {
		endorsers[index].TLSCARoot = model.BytesPacked([]byte(endorser.TLSCARoot))
	}
	var endorsersInString = string(goutils.ToJson(endorsers))

	var body = url.Values{
		"endorsers":       {endorsersInString},
		"signed-proposal": {model.BytesPacked(signedBytes)},
		"proposal":        {proposal},
	}
	var response = http.PostForm(_url, body, nil) // TODO change to https
	var result = model.ProposalResponseResult{}
	return result.ParseOrPanic(response.BodyBytes())
}

func Commit(_orderer model.Node, transactionBytes []byte) string {
	_orderer.TLSCARoot = model.BytesPacked([]byte(_orderer.TLSCARoot))
	var body = url.Values{
		"orderer":     {string(goutils.ToJson(_orderer))},
		"transaction": {model.BytesPacked(transactionBytes)},
	}
	var _url = BuildURL("/fabric/transact/commit")
	var response = http.PostForm(_url, body, nil)
	var txResult = &orderer.BroadcastResponse{}
	goutils.FromJson(response.BodyBytes(), txResult)
	return common.Status_name[int32(txResult.Status)]
}
