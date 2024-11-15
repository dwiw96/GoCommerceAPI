package handler

import (
	"context"

	auth "github.com/dwiw96/vocagame-technical-test-backend/internal/features/auth"
	transactions "github.com/dwiw96/vocagame-technical-test-backend/internal/features/transactions"
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

type transactionsHandler struct {
	router   *gin.Engine
	service  transactions.IService
	validate *validator.Validate
	trans    ut.Translator
}

func NewTransactionsHandler(router *gin.Engine, service transactions.IService, pool *pgxpool.Pool, client *redis.Client, ctx context.Context) {
	handler := &transactionsHandler{
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

	router.POST("/api/v1/transactions", handler.transaction)
}

func translateError(trans ut.Translator, err error) (errTrans []string) {
	errs := err.(validator.ValidationErrors)
	a := (errs.Translate(trans))
	for _, val := range a {
		errTrans = append(errTrans, val)
	}

	return
}

func (h *transactionsHandler) transaction(c *gin.Context) {
	authPayload, isExists := c.Keys["payloadKey"].(*auth.JwtPayload)
	if !isExists {
		responses.ErrorJSON(c, 401, []string{"token is wrong"}, c.Request.RemoteAddr)
		return
	}

	var reqBody transactionReq
	err := c.BindJSON(&reqBody)
	if err != nil {
		responses.ErrorJSON(c, 422, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}
	err = h.validate.Struct(&reqBody)
	if err != nil {
		errTranslated := translateError(h.trans, err)
		responses.ErrorJSON(c, 422, errTranslated, c.Request.RemoteAddr)
		return
	}

	var (
		res  *transactions.TransactionHistory
		code int
	)

	transactionsArg := toTransactionstArg(authPayload.UserID, reqBody)
	switch reqBody.TransactionType {
	case "purchase":
		res, code, err = h.service.PurchaseProduct(transactionsArg)
	case "deposit":
		res, code, err = h.service.Deposit(transactionsArg)
	}

	if err != nil && res != nil {
		respBody := toTransactionResp(res)
		response := responses.ErrorWithDataResponse(respBody, code, err.Error(), "failed")
		c.IndentedJSON(code, response)
		return
	}
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	respBody := toTransactionResp(res)
	response := responses.SuccessWithDataResponse(respBody, code, "success")
	c.IndentedJSON(code, response)
}
