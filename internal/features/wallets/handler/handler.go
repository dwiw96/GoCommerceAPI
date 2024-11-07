package handler

import (
	"context"

	auth "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth"
	wallets "github.com/dwiw96/vocagame-technical-test-backend/internal/features/wallets"
	mid "github.com/dwiw96/vocagame-technical-test-backend/pkg/middleware"
	responses "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/responses"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
)

type walletsHandler struct {
	router   *gin.Engine
	service  wallets.IService
	validate *validator.Validate
	trans    ut.Translator
}

func NewWalletsHandler(router *gin.Engine, service wallets.IService, pool *pgxpool.Pool, client *redis.Client, ctx context.Context) {
	handler := &walletsHandler{
		router:   router,
		service:  service,
		validate: validator.New(),
	}

	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(handler.validate, trans)
	handler.trans = trans

	router.Use(mid.AuthMiddleware(ctx, pool, client))

	router.POST("/api/v1/wallets", handler.createWallet)
	router.GET("/api/v1/wallets/:user_id", handler.getWallet)
	router.PUT("/api/v1/wallets/:user_id/deposit", handler.depositToWallet)
	router.PUT("/api/v1/wallets/:user_id/withdraw", handler.withdrawFromWallet)
}

func translateError(trans ut.Translator, err error) (errTrans []string) {
	errs := err.(validator.ValidationErrors)
	a := (errs.Translate(trans))
	for _, val := range a {
		errTrans = append(errTrans, val)
	}

	return
}

func (h *walletsHandler) createWallet(c *gin.Context) {
	authPayload, isExists := c.Keys["payloadKey"].(*auth.JwtPayload)

	if !isExists {
		responses.ErrorJSON(c, 401, []string{"token is wrong"}, c.Request.RemoteAddr)
		return
	}

	arg := wallets.CreateWalletParams{
		UserID:  authPayload.UserID,
		Balance: 0,
	}
	res, code, err := h.service.CreateWallet(arg)
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	respBody := toWalletResp(res)

	response := responses.SuccessWithDataResponse(respBody, code, "create new wallet success")
	c.IndentedJSON(code, response)
}

func (h *walletsHandler) getWallet(c *gin.Context) {
	var urlParam walletUrlParam
	if err := c.ShouldBindUri(&urlParam); err != nil {
		if err := h.validate.Struct(urlParam); err != nil {
			errTranslated := translateError(h.trans, err)
			responses.ErrorJSON(c, 422, errTranslated, c.Request.RemoteAddr)
			return
		}

		responses.ErrorJSON(c, 422, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	res, code, err := h.service.GetWalletByUserID(urlParam.UserID)
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	respBody := toWalletResp(res)

	response := responses.SuccessWithDataResponse(respBody, code, "get wallet success")
	c.IndentedJSON(code, response)
}

func (h *walletsHandler) depositToWallet(c *gin.Context) {
	var urlParam walletUrlParam
	if err := c.ShouldBindUri(&urlParam); err != nil {
		if err := h.validate.Struct(urlParam); err != nil {
			errTranslated := translateError(h.trans, err)
			responses.ErrorJSON(c, 422, errTranslated, c.Request.RemoteAddr)
			return
		}

		responses.ErrorJSON(c, 422, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	var reqBody updateWalletReq
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		if err := h.validate.Struct(reqBody); err != nil {
			errTranslated := translateError(h.trans, err)
			responses.ErrorJSON(c, 422, errTranslated, c.Request.RemoteAddr)
			return
		}

		responses.ErrorJSON(c, 422, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	arg := wallets.UpdateWalletParams{
		Amount: reqBody.Amount,
		UserID: urlParam.UserID,
	}

	res, code, err := h.service.DepositToWallet(arg)
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	respBody := toWalletResp(res)

	response := responses.SuccessWithDataResponse(respBody, code, "deposit to wallet success")
	c.IndentedJSON(code, response)
}

func (h *walletsHandler) withdrawFromWallet(c *gin.Context) {
	var urlParam walletUrlParam
	if err := c.ShouldBindUri(&urlParam); err != nil {
		if err := h.validate.Struct(urlParam); err != nil {
			errTranslated := translateError(h.trans, err)
			responses.ErrorJSON(c, 422, errTranslated, c.Request.RemoteAddr)
			return
		}

		responses.ErrorJSON(c, 422, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	var reqBody updateWalletReq
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		if err := h.validate.Struct(reqBody); err != nil {
			errTranslated := translateError(h.trans, err)
			responses.ErrorJSON(c, 422, errTranslated, c.Request.RemoteAddr)
			return
		}

		responses.ErrorJSON(c, 422, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	arg := wallets.UpdateWalletParams{
		Amount: reqBody.Amount,
		UserID: urlParam.UserID,
	}

	res, code, err := h.service.WithdrawFromWallet(arg)
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	respBody := toWalletResp(res)

	response := responses.SuccessWithDataResponse(respBody, code, "withdraw from wallet success")
	c.IndentedJSON(code, response)
}
