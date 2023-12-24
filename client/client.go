package client

import (
	"context"
	"fmt"
	"github.com/davidkhala/fabric-common/golang"
	"github.com/davidkhala/fabric-server-go/model"
	"github.com/davidkhala/goutils"
	"github.com/davidkhala/protoutil"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/protobuf/proto"
)

func InitOrPanic(config golang.CryptoConfig) *golang.Crypto {
	cryptoObject, err := golang.LoadCryptoFrom(config)
	goutils.PanicError(err)
	return cryptoObject
}

func ReadPEMFile(file string) []byte {
	byteSlice, err := goutils.ReadFile(file)
	goutils.PanicError(err)
	return byteSlice
}
func GetProposalSigned(proposal string, signer *golang.Crypto) (signedBytes []byte) {
	var bytes = model.BytesFromString(proposal)
	var signature, err = signer.Sign(bytes)
	goutils.PanicError(err)

	var signed = peer.SignedProposal{
		ProposalBytes: bytes,
		Signature:     signature,
	}
	signedBytes, err = proto.Marshal(&signed)
	goutils.PanicError(err)
	return
}
func CommitProposalAndSign(proposal string, signedBytes []byte, endorsers []model.Node, signer golang.Crypto) []byte {
	_, payload := Propose(proposal, signedBytes, endorsers)
	// sign the payload
	sig, err := signer.Sign(payload)
	goutils.PanicError(err)
	// here's the envelope
	var envelop = common.Envelope{Payload: payload, Signature: sig}
	return protoutil.MarshalOrPanic(&envelop)
}
func QueryProposal(proposal string, signedBytes []byte, endorsers []model.Node) (result string) {
	parsedResult, _ := Propose(proposal, signedBytes, endorsers)
	var proposalResponse *peer.ProposalResponse

	if len(parsedResult) == 0 {
		panic("no proposalResponses found")
	}
	for _, proposalResponse = range parsedResult {
		if proposalResponse.Response.Status != 200 {
			panic(proposalResponse.Response.Message)
		}
		var currentResult = model.ShimResultFrom(proposalResponse).Payload
		if result != "" && result != currentResult {
			panic(fmt.Sprintf("expect result aligning to %s, but got %s", result, currentResult))
		} else {
			result = currentResult
		}
	}
	return

}

type GetTransactionByIDResult struct {
	Transaction *common.Payload

	Validation string
}

func (GetTransactionByIDResult) FromString(str string) GetTransactionByIDResult {
	var as = peer.ProcessedTransaction{}
	err := proto.Unmarshal([]byte(str), &as)
	goutils.PanicError(err)
	var result = GetTransactionByIDResult{}
	result.Transaction = protoutil.UnmarshalPayloadOrPanic(as.TransactionEnvelope.Payload)
	result.Validation = peer.TxValidationCode_name[as.ValidationCode]
	return result
}

type Eventer struct {
	golang.Eventer
}

// TODO support multiple eventer
func EventerFrom(node model.Node) Eventer {

	var node_translated = golang.Node{
		Addr:                  node.Address,
		TLSCARootByte:         model.BytesFromString(node.TLSCARoot),
		SslTargetNameOverride: node.SslTargetNameOverride,
	}
	grpcClient, err := node_translated.AsGRPCClient()
	goutils.PanicError(err)
	return Eventer{golang.EventerFrom(context.Background(), grpcClient)}
}

func (e Eventer) WaitForTx(channel, txid string, signer *golang.Crypto) (txStatus string) {
	var seek = e.AsTransactionListener(txid)
	signedEvent, err := seek.SignBy(channel, signer)
	goutils.PanicError(err)
	_, err = e.SendRecv(signedEvent)
	goutils.PanicError(err)
	return fmt.Sprint(e.ReceiptData)
}
