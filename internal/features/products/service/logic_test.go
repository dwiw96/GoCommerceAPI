package service

import (
	"context"
	"os"
	"testing"

	product "github.com/dwiw96/vocagame-technical-test-backend/internal/features/products"
	repo "github.com/dwiw96/vocagame-technical-test-backend/internal/features/products/repository"
	testUtils "github.com/dwiw96/vocagame-technical-test-backend/testutils"

	converter "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/converter"
	generator "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/generator"
	errorHandler "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/responses"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	serviceTest product.IService
	ctx         context.Context
	pool        *pgxpool.Pool
)

func TestMain(m *testing.M) {
	pool = testUtils.GetPool()
	defer pool.Close()
	ctx = testUtils.GetContext()
	defer ctx.Done()

	schemaCleanup := testUtils.SetupDB("test_service_product")

	repoTest := repo.NewProductRepository(pool)
	serviceTest = NewProductService(ctx, repoTest)

	exitTest := m.Run()

	schemaCleanup()

	os.Exit(exitTest)
}

func createProductTest(t *testing.T) (input product.CreateProductParams, res *product.Product) {
	arg := product.CreateProductParams{
		Name:         generator.CreateRandomString(7),
		Description:  generator.CreateRandomString(50),
		Price:        int32(generator.RandomInt(5, 500)),
		Availability: int32(generator.RandomInt(0, 50)),
	}

	res, code, err := serviceTest.CreateProduct(arg)
	require.NoError(t, err)
	assert.Equal(t, arg.Name, res.Name)
	assert.Equal(t, arg.Description, res.Description)
	assert.Equal(t, arg.Price, res.Price)
	assert.Equal(t, arg.Availability, res.Availability)
	assert.Equal(t, errorHandler.CodeSuccessCreate, code)

	return arg, res
}

func TestCreateProduct(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	name := generator.CreateRandomString(7)

	testCases := []struct {
		desc   string
		params product.CreateProductParams
		code   int
		err    bool
	}{
		{
			desc: "success_all_params",
			params: product.CreateProductParams{
				Name:         name,
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			code: errorHandler.CodeSuccessCreate,
			err:  false,
		}, {
			desc: "success_partial_params",
			params: product.CreateProductParams{
				Name:         generator.CreateRandomString(7),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			code: errorHandler.CodeSuccessCreate,
			err:  false,
		}, {
			desc: "failed_duplicate_name",
			params: product.CreateProductParams{
				Name:         name,
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			code: errorHandler.CodeFailedDuplicated,
			err:  true,
		}, {
			desc: "failed_no_name",
			params: product.CreateProductParams{
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			code: errorHandler.CodeFailedUser,
			err:  true,
		}, {
			desc: "failed_minus_price",
			params: product.CreateProductParams{
				Name:         generator.CreateRandomString(7),
				Description:  generator.CreateRandomString(50),
				Price:        -1,
				Availability: int32(generator.RandomInt(0, 50)),
			},
			code: errorHandler.CodeFailedUser,
			err:  true,
		}, {
			desc: "failed_minus_availability",
			params: product.CreateProductParams{
				Name:         generator.CreateRandomString(7),
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: -1,
			},
			code: errorHandler.CodeFailedUser,
			err:  true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.CreateProduct(tC.params)
			assert.Equal(t, tC.code, code)
			if !tC.err {
				require.NoError(t, err)
				assert.NotZero(t, res.ID)
				assert.Equal(t, tC.params.Name, res.Name)
				assert.Equal(t, tC.params.Description, res.Description)
				assert.Equal(t, tC.params.Price, res.Price)
				assert.Equal(t, tC.params.Availability, res.Availability)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetProductByID(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	_, resProduct := createProductTest(t)
	testCases := []struct {
		desc string
		id   int32
		ans  product.Product
		err  bool
	}{
		{
			desc: "success",
			id:   resProduct.ID,
			ans: product.Product{
				ID:           resProduct.ID,
				Name:         resProduct.Name,
				Description:  resProduct.Description,
				Price:        resProduct.Price,
				Availability: resProduct.Availability,
			},
			err: false,
		}, {
			desc: "failed_wrong_id",
			id:   0,
			err:  true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.GetProductByID(converter.ConvertInt32ToString(tC.id))
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, tC.ans.ID, res.ID)
				assert.Equal(t, tC.ans.Name, res.Name)
				assert.Equal(t, tC.ans.Description, res.Description)
				assert.Equal(t, tC.ans.Price, res.Price)
				assert.Equal(t, tC.ans.Availability, res.Availability)
				assert.Equal(t, 200, code)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestListProducts(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	length := 10
	var input []product.Product
	for i := 0; i < length; i++ {
		_, temp := createProductTest(t)

		input = append(input, *temp)
	}
	testCases := []struct {
		desc       string
		page       int32
		limit      int32
		ans        []product.Product
		length     int
		totalPages int
		code       int
		err        bool
	}{
		{
			desc:       "success_full",
			page:       1,
			limit:      10,
			length:     10,
			totalPages: 1,
			code:       errorHandler.CodeSuccess,
			err:        false,
		}, {
			desc:       "success_partial",
			page:       3,
			limit:      4,
			length:     2,
			totalPages: 3,
			code:       errorHandler.CodeSuccess,
			err:        false,
		}, {
			desc:       "success_empty",
			page:       100,
			limit:      100,
			length:     0,
			totalPages: 1,
			code:       errorHandler.CodeSuccess,
			err:        false,
		}, {
			desc:       "success_no_page",
			limit:      3,
			length:     3,
			totalPages: 4,
			code:       errorHandler.CodeSuccess,
			err:        false,
		}, {
			desc:       "success_no_limit",
			page:       0,
			length:     10,
			totalPages: 1,
			code:       errorHandler.CodeSuccess,
			err:        false,
		}, {
			desc:       "success_minus_arg",
			page:       -1,
			limit:      -5,
			length:     10,
			totalPages: 1,
			code:       errorHandler.CodeSuccess,
			err:        false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, _, totalPages, code, err := serviceTest.ListProducts(converter.ConvertInt32ToString(tC.page), converter.ConvertInt32ToString(tC.limit))
			if !tC.err {
				require.NoError(t, err)
				if res != nil {
					assert.Equal(t, tC.length, len(*res))
				}
				assert.Equal(t, tC.totalPages, totalPages)
				assert.Equal(t, tC.code, code)

				switch tC.desc {
				case "success_full":
					assert.Equal(t, input, *res)
				case "success_partial":
					for i := 0; i < 2; i++ {
						assert.Equal(t, input[i+8], (*res)[i])
					}
				case "success_empty":
					assert.Empty(t, res)
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestUpdateProduct(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	_, resProduct := createProductTest(t)
	_, resProduct2 := createProductTest(t)
	name := generator.CreateRandomString(7)
	testCases := []struct {
		desc string
		arg  product.UpdateProductParams
		code int
		err  bool
	}{
		{
			desc: "success_all_arg",
			arg: product.UpdateProductParams{
				ID:           resProduct.ID,
				Name:         name,
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			code: errorHandler.CodeSuccess,
			err:  false,
		}, {
			desc: "success_null_3_arg",
			arg: product.UpdateProductParams{
				ID:   resProduct.ID,
				Name: generator.CreateRandomString(7),
			},
			code: errorHandler.CodeSuccess,
			err:  false,
		}, {
			desc: "success_no_change",
			arg: product.UpdateProductParams{
				ID:           resProduct.ID,
				Name:         resProduct.Name,
				Description:  resProduct.Description,
				Price:        resProduct.Price,
				Availability: resProduct.Availability,
			},
			code: errorHandler.CodeSuccess,
			err:  false,
		}, {
			desc: "failed_null_name",
			arg: product.UpdateProductParams{
				ID:           resProduct.ID,
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			code: errorHandler.CodeFailedUser,
			err:  true,
		}, {
			desc: "failed_duplicate_name",
			arg: product.UpdateProductParams{
				ID:           resProduct.ID,
				Name:         resProduct2.Name,
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			code: errorHandler.CodeFailedDuplicated,
			err:  true,
		}, {
			desc: "failed_wrong_id",
			arg: product.UpdateProductParams{
				ID:           0,
				Name:         generator.CreateRandomString(7),
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			code: errorHandler.CodeFailedUser,
			err:  true,
		}, {
			desc: "failed_invalid price",
			arg: product.UpdateProductParams{
				ID:           resProduct.ID,
				Name:         generator.CreateRandomString(7),
				Description:  generator.CreateRandomString(50),
				Price:        -5,
				Availability: int32(generator.RandomInt(0, 50)),
			},
			code: errorHandler.CodeFailedUser,
			err:  true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, code, err := serviceTest.UpdateProduct(tC.arg)
			assert.Equal(t, tC.code, code)
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, resProduct.ID, res.ID)
				switch tC.desc {
				case "success_all_arg":
					assert.Equal(t, tC.arg.Name, res.Name)
					assert.Equal(t, tC.arg.Description, res.Description)
					assert.Equal(t, tC.arg.Price, res.Price)
					assert.Equal(t, tC.arg.Availability, res.Availability)
				case "success_null_name":
					assert.Equal(t, resProduct.Name, res.Name)
					assert.Equal(t, tC.arg.Description, res.Description)
					assert.Equal(t, tC.arg.Price, res.Price)
					assert.Equal(t, tC.arg.Availability, res.Availability)
				case "success_null_3_arg":
					assert.Equal(t, tC.arg.Name, res.Name)
					assert.Equal(t, "", res.Description)
					assert.Equal(t, int32(0), res.Price)
					assert.Equal(t, int32(0), res.Availability)
				case "success_no_change":
					assert.Equal(t, resProduct.Name, res.Name)
					assert.Equal(t, resProduct.Description, res.Description)
					assert.Equal(t, resProduct.Price, res.Price)
					assert.Equal(t, resProduct.Availability, res.Availability)
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	err := testUtils.DeleteSchemaTestData(pool)
	require.NoError(t, err)

	_, resProduct := createProductTest(t)
	testCases := []struct {
		desc string
		id   int32
		err  bool
	}{
		{
			desc: "success",
			id:   resProduct.ID,
			err:  false,
		}, {
			desc: "failed",
			id:   0,
			err:  true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := serviceTest.DeleteProduct(converter.ConvertInt32ToString(tC.id))
			if !tC.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
