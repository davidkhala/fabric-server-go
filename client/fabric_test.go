package client

import (
	"github.com/davidkhala/fabric-server-go/model"
	"github.com/davidkhala/goutils"
	"github.com/kortschak/utter"
	rawHttp "net/http"
	"net/url"
	"testing"
)

func TestPing(t *testing.T) {
	var _url = BuildURL("/fabric/ping")
	var _path = goutils.HomeResolve("delphi-fabric/config/ca-crypto-config/peerOrganizations/icdd/tlsca/tlsca.icdd-cert.pem")
	var Certificate = model.BytesPacked(ReadPEMFile(_path))

	var body = url.Values{
		"address":                  {"localhost:8051"},
		"certificate":              {Certificate},
		"ssl-target-name-override": {"peer0.icdd"},
	}

	response, err := rawHttp.PostForm(_url, body)
	goutils.PanicError(err)
	utter.Dump(response.Status)
}
