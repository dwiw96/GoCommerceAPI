package repository

import (
	"context"
	"os"
	"testing"

	cfg "github.com/dwiw96/vocagame-technical-test-backend/config"
	product "github.com/dwiw96/vocagame-technical-test-backend/internal/features/products"
	pg "github.com/dwiw96/vocagame-technical-test-backend/pkg/driver/postgresql"
	generator "github.com/dwiw96/vocagame-technical-test-backend/pkg/utils/generator"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	repoTest product.IRepository
	ctx      context.Context
	pool     *pgxpool.Pool
)

func TestMain(m *testing.M) {
	env := cfg.GetEnvConfig()
	pool = pg.ConnectToPg(env)
	defer pool.Close()
	ctx = context.Background()
	defer ctx.Done()

	repoTest = NewProductRepository(pool)

	os.Exit(m.Run())
}

func createProductTest(t *testing.T) (input product.CreateProductParams, res *product.Product) {
	arg := product.CreateProductParams{
		Name:         generator.CreateRandomString(7),
		Description:  generator.CreateRandomString(50),
		Price:        int32(generator.RandomInt(5, 500)),
		Availability: int32(generator.RandomInt(0, 50)),
	}

	res, err := repoTest.CreateProduct(ctx, arg)
	require.NoError(t, err)
	assert.Equal(t, arg.Name, res.Name)
	assert.Equal(t, arg.Description, res.Description)
	assert.Equal(t, arg.Price, res.Price)
	assert.Equal(t, arg.Availability, res.Availability)

	return arg, res
}

func TestCreateProduct(t *testing.T) {
	ctx := context.Background()

	name := generator.CreateRandomString(7)
	description := generator.CreateRandomString(50)
	price := int32(generator.RandomInt(5, 500))
	availability := int32(generator.RandomInt(0, 50))

	testCases := []struct {
		name   string
		params product.CreateProductParams
		ctx    context.Context
		err    bool
	}{
		{
			name: "success_all_params",
			params: product.CreateProductParams{
				Name:         name,
				Description:  description,
				Price:        price,
				Availability: availability,
			},
			ctx: ctx,
			err: false,
		}, {
			name: "success_partial_params",
			params: product.CreateProductParams{
				Name:         name,
				Description:  "",
				Price:        price,
				Availability: availability,
			},
			ctx: ctx,
			err: false,
		}, {
			name: "failed_no_name",
			params: product.CreateProductParams{
				Description:  description,
				Price:        price,
				Availability: availability,
			},
			ctx: ctx,
			err: true,
		}, {
			name: "failed_minus_price",
			params: product.CreateProductParams{
				Name:         name,
				Description:  description,
				Price:        -1,
				Availability: availability,
			},
			ctx: ctx,
			err: true,
		}, {
			name: "failed_minus_availability",
			params: product.CreateProductParams{
				Name:         name,
				Description:  description,
				Price:        price,
				Availability: -1,
			},
			ctx: ctx,
			err: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := repoTest.CreateProduct(ctx, test.params)
			if !test.err {
				require.NoError(t, err)
				assert.NotZero(t, res.ID)
				assert.Equal(t, test.params.Name, res.Name)
				assert.Equal(t, test.params.Description, res.Description)
				assert.Equal(t, test.params.Price, res.Price)
				assert.Equal(t, test.params.Availability, res.Availability)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetProductByID(t *testing.T) {
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
			res, err := repoTest.GetProductByID(ctx, tC.id)
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, tC.ans.ID, res.ID)
				assert.Equal(t, tC.ans.Name, res.Name)
				assert.Equal(t, tC.ans.Description, res.Description)
				assert.Equal(t, tC.ans.Price, res.Price)
				assert.Equal(t, tC.ans.Availability, res.Availability)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestListProducts(t *testing.T) {
	deleteQuery := `DELETE FROM products;`
	_, err := pool.Exec(ctx, deleteQuery)
	require.NoError(t, err)

	var input []product.Product
	for i := 0; i < 6; i++ {
		_, temp := createProductTest(t)

		input = append(input, *temp)
	}
	testCases := []struct {
		desc  string
		input product.ListProductsParams
		ans   []product.Product
		err   bool
	}{
		{
			desc: "success_full",
			input: product.ListProductsParams{
				Limit:  6,
				Offset: 0,
			},
			err: false,
		}, {
			desc: "success_partial",
			input: product.ListProductsParams{
				Limit:  4,
				Offset: 3,
			},
			err: false,
		}, {
			desc: "success_empty",
			input: product.ListProductsParams{
				Limit:  5,
				Offset: 1000,
			},
			err: false,
		}, {
			desc: "failed_minus_limit",
			input: product.ListProductsParams{
				Limit:  -1,
				Offset: 0,
			},
			err: true,
		}, {
			desc: "failed_minus_offset",
			input: product.ListProductsParams{
				Limit:  5,
				Offset: -50,
			},
			err: true,
		}, {
			desc: "failed_minus_arg",
			input: product.ListProductsParams{
				Limit:  -1,
				Offset: -1,
			},
			err: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.ListProducts(ctx, tC.input)
			if !tC.err {
				require.NoError(t, err)
				switch tC.desc {
				case "success_full":
					assert.Equal(t, 6, len(*res))
					assert.Equal(t, input, *res)
				case "success_partial":
					assert.Equal(t, 3, len(*res))
					for i := 0; i < 3; i++ {
						assert.Equal(t, input[i+3], (*res)[i])
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

func TestGetTotalProducts(t *testing.T) {
	deleteQuery := `DELETE FROM products;`
	_, err := pool.Exec(ctx, deleteQuery)
	require.NoError(t, err)

	length := 6
	for i := 0; i < length; i++ {
		createProductTest(t)
	}

	res, err := repoTest.GetTotalProducts(ctx)
	require.NoError(t, err)
	assert.Equal(t, length, res)
}

func TestUpdateProduct(t *testing.T) {
	_, resProduct := createProductTest(t)
	testCases := []struct {
		desc string
		arg  product.UpdateProductParams
		err  bool
	}{
		{
			desc: "success_all_arg",
			arg: product.UpdateProductParams{
				ID:           resProduct.ID,
				Name:         generator.CreateRandomString(7),
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			err: false,
		}, {
			desc: "failed_null_name",
			arg: product.UpdateProductParams{
				ID:           resProduct.ID,
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			err: true,
		}, {
			desc: "success_null_3_arg",
			arg: product.UpdateProductParams{
				ID:   resProduct.ID,
				Name: generator.CreateRandomString(7),
			},
			err: false,
		}, {
			desc: "success_no_change",
			arg: product.UpdateProductParams{
				ID:           resProduct.ID,
				Name:         resProduct.Name,
				Description:  resProduct.Description,
				Price:        resProduct.Price,
				Availability: resProduct.Availability,
			},
			err: false,
		}, {
			desc: "failed_wrong_id",
			arg: product.UpdateProductParams{
				ID:           0,
				Name:         generator.CreateRandomString(7),
				Description:  generator.CreateRandomString(50),
				Price:        int32(generator.RandomInt(5, 500)),
				Availability: int32(generator.RandomInt(0, 50)),
			},
			err: true,
		}, {
			desc: "failed_invalid price",
			arg: product.UpdateProductParams{
				ID:           resProduct.ID,
				Name:         generator.CreateRandomString(7),
				Description:  generator.CreateRandomString(50),
				Price:        -5,
				Availability: int32(generator.RandomInt(0, 50)),
			},
			err: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.UpdateProduct(ctx, tC.arg)
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
			err := repoTest.DeleteProduct(ctx, tC.id)
			if !tC.err {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
