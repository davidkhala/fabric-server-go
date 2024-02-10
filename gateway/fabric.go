package gateway

import (
	"github.com/davidkhala/fabric-common/golang"
	"github.com/davidkhala/fabric-server-go/model"
	"github.com/davidkhala/goutils"
	"github.com/davidkhala/goutils/crypto"
	"github.com/davidkhala/goutils/grpc"
	"github.com/davidkhala/protoutil"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
	"net/http"
)

// PingFabric
// @Router /fabric/ping [post]
// @Produce text/plain
// @Accept x-www-form-urlencoded
// @Param address formData string true "endpoint like grpc(s)://\<fqdn\> or \<fqdn\>"
// @Param certificate formData string true "Certificate in PEM format. should be in hex format after translation to solve linebreak issue"
// @Param ssl-target-name-override formData string true "pseudo endpoint \<fqdn\>"
// @Success 200 {string} string pong
// @Failure 400 {string} string Bad request
func PingFabric(c *gin.Context) {

	address := c.PostForm("address")
	certificatePEM := model.BytesFromForm(c, "certificate")

	certificate, err := crypto.ParseCertPem(certificatePEM)
	if err != nil {
		c.String(http.StatusBadRequest, "Bad request: [certificate]")
		return
	}
	var param = grpc.Params{
		SslTargetNameOverride: c.DefaultPostForm("ssl-target-name-override", golang.ToAddress(address)),
		Certificate:           certificate,
		WaitForReady:          true,
	}
	_, err = golang.Ping(address, param)

	if err != nil {
		c.String(http.StatusServiceUnavailable, "ServiceUnavailable")
		return
	}
	c.String(http.StatusOK, "pong")
}

// ProcessProposal
// @Router /fabric/transact/process-proposal [post]
// @Produce json
// @Accept x-www-form-urlencoded
// @Param endorsers formData string true "json data to specify endorsers"
// @Param signed-proposal formData string true "Hex-encoded and serialized signed-proposal protobuf"
// @Param proposal formData string true "Hex-encoded and serialized proposal protobuf"
// @Success 200 {object} model.ProposalResponseResult
func ProcessProposal(c *gin.Context) {
	endorsers := c.PostForm("endorsers")
	signedBytes := model.BytesFromForm(c, "signed-proposal")
	proposalBytes := model.BytesFromForm(c, "proposal")
	var signed = peer.SignedProposal{}
	err := proto.Unmarshal(signedBytes, &signed)
	goutils.PanicError(err)
	var proposalResponses []*peer.ProposalResponse
	var proposalResponseAsStrings []string
	var endorserNodes []model.Node
	goutils.FromJson([]byte(endorsers), &endorserNodes)
	for _, node := range endorserNodes {
		var nodeTranslated = golang.Node{
			Addr:                  node.Address,
			TLSCARootByte:         model.BytesFromString(node.TLSCARoot),
			SslTargetNameOverride: node.SslTargetNameOverride,
		}
		grpcClient := nodeTranslated.AsGRPCClientOrPanic() // FIXME multiple error type

		endorserClient := golang.EndorserFrom(c, grpcClient)
		proposalResponse, _err := endorserClient.ProcessProposal(&signed)
		goutils.PanicError(_err)
		proposalResponses = append(proposalResponses, proposalResponse)

		proposalResponseAsStrings = append(proposalResponseAsStrings, model.BytesPacked(protoutil.MarshalOrPanic(proposalResponse)))
	}

	// prepare unsigned tx
	proposal, err := protoutil.UnmarshalProposal(proposalBytes)
	goutils.PanicError(err)
	payloadBytes, err := CreateUnSignedTx(proposal, proposalResponses)
	var result = model.ProposalResponseResult{
		ProposalResponses: proposalResponseAsStrings,
		Payload:           model.BytesPacked(payloadBytes),
	}
	c.JSON(http.StatusOK, result)
}

// Commit
// @Router /fabric/transact/commit [post]
// @Produce json
// @Accept x-www-form-urlencoded
// @Param orderer formData string true "json data to specify orderer"
// @Param transaction formData string true "serialized signed proposalResponses as envelop protobuf with hex format"
// @Success 200 {object} model.TxResult
func Commit(c *gin.Context) {
	var orderer = c.PostForm("orderer")
	var transaction = model.BytesFromForm(c, "transaction")
	var envelop = &common.Envelope{}
	err := proto.Unmarshal(transaction, envelop)
	goutils.PanicError(err)
	var ordererNode model.Node
	goutils.FromJson([]byte(orderer), &ordererNode)

	var nodeTranslated = golang.Node{
		Addr:                  ordererNode.Address,
		TLSCARootByte:         model.BytesFromString(ordererNode.TLSCARoot),
		SslTargetNameOverride: ordererNode.SslTargetNameOverride,
	}
	ordererGrpc := nodeTranslated.AsGRPCClientOrPanic() // FIXME multiple error type

	var committer = golang.Committer{
		AtomicBroadcastClient: golang.CommitterFrom(ordererGrpc),
	}
	err = committer.Setup(c)
	goutils.PanicError(err)

	txResult, err := committer.SendRecv(envelop)
	goutils.PanicError(err)
	c.JSON(http.StatusOK, txResult)
}
