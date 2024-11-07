package handler

type walletUrlParam struct {
	UserID int32 `uri:"user_id" validate:"required,number"`
}

type updateWalletReq struct {
	Amount int32 `json:"amount" validate:"required,number"`
}
