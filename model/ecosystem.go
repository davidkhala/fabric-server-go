package model

type CreateTokenResult struct {
	CreateProposalResult
	Token string `json:"token"`
}
