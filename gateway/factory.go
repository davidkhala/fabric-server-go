package gateway

import (
	"bytes"
	"github.com/davidkhala/fabric-common/golang/proposal"
	"github.com/davidkhala/fabric-server-go/model"
	"github.com/davidkhala/goutils"
	"github.com/davidkhala/protoutil"
	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/pkg/errors"
	"net/http"
)

// CreateProposal
// @Router /fabric/create-proposal [post]
// @Produce json
// @Accept x-www-form-urlencoded
// @Param creator formData string true "Hex-encoded creator bytes"
// @Param channel formData string true "Fabric channel name"
// @Param chaincode formData string true "Fabric chaincode name"
// @Param args formData string true "Fabric chaincode calling args, string array as JSON"
// @Param transient formData string false "JSON format, like map[string]string"
// @Success 200 {object} model.CreateProposalResult
func CreateProposal(c *gin.Context) {

	creator := model.BytesFromForm(c, "creator")
	channel := c.PostForm("channel")
	chaincode := c.PostForm("chaincode")
	var args []string
	goutils.FromJson([]byte(c.PostForm("args")), &args)
	var rawTransient = c.PostForm("transient")
	var transientBytes = map[string][]byte{}

	if rawTransient != "" {
		var transient = map[string]string{}
		goutils.FromJson([]byte(rawTransient), &transient)
		for key, value := range transient {
			transientBytes[key] = []byte(value)
		}
	} else {
		transientBytes = nil
	}

	createdProposal, txid, err := proposal.CreateProposal(
		creator,
		channel,
		chaincode,
		args,
		transientBytes,
		proposal.WithType(peer.ChaincodeSpec_GOLANG),
	)

	goutils.PanicError(err)

	c.JSON(http.StatusOK, model.CreateProposalResult{
		Proposal: model.BytesPacked(protoutil.MarshalOrPanic(createdProposal)),
		Txid:     txid,
	})
}

func CreateUnSignedTx(proposal *peer.Proposal, responses []*peer.ProposalResponse) ([]byte, error) {
	if len(responses) == 0 {
		return nil, errors.Errorf("at least one proposal response is required")
	}

	// the original header
	hdr, err := protoutil.UnmarshalHeader(proposal.Header)
	if err != nil {
		return nil, err
	}

	// the original payload
	pPayload, err := protoutil.UnmarshalChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, err
	}

	endorsements := make([]*peer.Endorsement, 0)

	// ensure that all actions are bitwise equal and that they are successful
	var a1 []byte
	for n, r := range responses {
		if n == 0 {
			a1 = r.Payload
			if r.Response.Status < 200 || r.Response.Status >= 400 {
				return nil, errors.Errorf("proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
			}
		}
		if bytes.Compare(a1, r.Payload) != 0 {
			return nil, errors.Errorf("ProposalResponsePayloads from Peers do not match")
		}
		endorsements = append(endorsements, r.Endorsement)
	}
	// create ChaincodeEndorsedAction
	cea := &peer.ChaincodeEndorsedAction{ProposalResponsePayload: a1, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := protoutil.GetBytesProposalPayloadForTx(pPayload) //, hdrExt.PayloadVisibility
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	c := &peer.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := protoutil.GetBytesChaincodeActionPayload(c)
	if err != nil {
		return nil, err
	}

	// create a transaction
	txAction := &peer.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	txActions := make([]*peer.TransactionAction, 1)
	txActions[0] = txAction
	tx := &peer.Transaction{Actions: txActions}
	// serialize the tx
	txBytes, err := protoutil.GetBytesTransaction(tx)
	if err != nil {
		return nil, err
	}

	// create the payload
	payload := &common.Payload{Header: hdr, Data: txBytes}
	bytesPayload, err := protoutil.GetBytesPayload(payload)
	if err != nil {
		return nil, err
	}
	return bytesPayload, nil

}
