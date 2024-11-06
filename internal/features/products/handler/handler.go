package handler

import (
	"context"

	product "github.com/dwiw96/vocagame-technical-test-backend/internal/features/products"
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

type productHandler struct {
	router   *gin.Engine
	service  product.IService
	validate *validator.Validate
	trans    ut.Translator
}

func NewProductHandler(router *gin.Engine, service product.IService, pool *pgxpool.Pool, client *redis.Client, ctx context.Context) {
	handler := &productHandler{
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

	router.POST("/api/v1/product/create", handler.createProduct)
	router.GET("/api/v1/product/get/:id", handler.getProduct)
	router.GET("/api/v1/product/list", handler.listProduct)
	router.PUT("/api/v1/product/update", handler.updateProduct)
	router.DELETE("/api/v1/product/delete/:id", handler.deleteProduct)
}

func translateError(trans ut.Translator, err error) (errTrans []string) {
	errs := err.(validator.ValidationErrors)
	a := (errs.Translate(trans))
	for _, val := range a {
		errTrans = append(errTrans, val)
	}

	return
}

func (h *productHandler) createProduct(c *gin.Context) {
	var request createProductReq

	err := c.BindJSON(&request)
	if err != nil {
		responses.ErrorJSON(c, 422, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	err = h.validate.Struct(request)
	if err != nil {
		errTranslated := translateError(h.trans, err)
		responses.ErrorJSON(c, 422, errTranslated, c.Request.RemoteAddr)
		return
	}

	serviceArg := product.CreateProductParams{
		Name:         request.Name,
		Description:  request.Description,
		Price:        request.Price,
		Availability: request.Availability,
	}

	res, code, err := h.service.CreateProduct(serviceArg)
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	response := responses.SuccessWithDataResponse(productResp(*res), code, "create new product success")
	c.IndentedJSON(code, response)
}

func (h *productHandler) getProduct(c *gin.Context) {
	productID := c.Param("id")

	res, code, err := h.service.GetProductByID(productID)
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	response := responses.SuccessWithDataResponse(productResp(*res), code, "success get product")
	c.IndentedJSON(code, response)
}

func (h *productHandler) listProduct(c *gin.Context) {
	page := c.DefaultQuery("page", "0")
	limit := c.DefaultQuery("limit", "10")

	res, currentPage, totalPages, code, err := h.service.ListProducts(page, limit)
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	response := responses.SuccessWithDataResponsePagination(res, currentPage, totalPages, "list of books")
	c.IndentedJSON(code, response)
}

func (h *productHandler) updateProduct(c *gin.Context) {
	var request updateProductReq

	err := c.BindJSON(&request)
	if err != nil {
		responses.ErrorJSON(c, 422, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	err = h.validate.Struct(request)
	if err != nil {
		errTranslated := translateError(h.trans, err)
		responses.ErrorJSON(c, 422, errTranslated, c.Request.RemoteAddr)
		return
	}

	res, code, err := h.service.UpdateProduct(product.UpdateProductParams(request))
	if err != nil {
		responses.ErrorJSON(c, code, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	response := responses.SuccessWithDataResponse(*res, code, "update success")
	c.IndentedJSON(code, response)
}

func (h *productHandler) deleteProduct(c *gin.Context) {
	productID := c.Param("id")

	err := h.service.DeleteProduct(productID)
	if err != nil {
		responses.ErrorJSON(c, 400, []string{err.Error()}, c.Request.RemoteAddr)
		return
	}

	response := responses.SuccessResponse("success deleted product")
	c.IndentedJSON(200, response)
}
