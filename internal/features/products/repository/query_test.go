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
		Name:         generator.CreateRandomString(10),
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

func deleteProductsTest(t *testing.T) {
	const query = `
	TRUNCATE TABLE
		transaction_histories,
		products;
	`

	_, err := pool.Exec(ctx, query)
	require.NoError(t, err)

}

func TestCreateProduct(t *testing.T) {
	name := generator.CreateRandomString(10)
	description := generator.CreateRandomString(50)
	price := int32(generator.RandomInt(5, 500))
	availability := int32(generator.RandomInt(0, 50))

	testCases := []struct {
		name   string
		params product.CreateProductParams
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
			err: false,
		}, {
			name: "success_partial_params",
			params: product.CreateProductParams{
				Name:         name + "2",
				Description:  "",
				Price:        price,
				Availability: availability,
			},
			err: false,
		}, {
			name: "failed_no_name",
			params: product.CreateProductParams{
				Description:  description,
				Price:        price,
				Availability: availability,
			},
			err: true,
		}, {
			name: "failed_minus_price",
			params: product.CreateProductParams{
				Name:         name,
				Description:  description,
				Price:        -1,
				Availability: availability,
			},
			err: true,
		}, {
			name: "failed_minus_availability",
			params: product.CreateProductParams{
				Name:         name,
				Description:  description,
				Price:        price,
				Availability: -1,
			},
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
	deleteProductsTest(t)

	var input []product.Product
	for i := 0; i < 6; i++ {
		_, temp := createProductTest(t)

		input = append(input, *temp)
	}
	testCases := []struct {
		desc   string
		arg    product.ListProductsParams
		length int
		idx    int
		err    bool
	}{
		{
			desc: "success_full",
			arg: product.ListProductsParams{
				Limit:  6,
				Offset: 0,
			},
			length: 6,
			idx:    0,
			err:    false,
		}, {
			desc: "success_partial",
			arg: product.ListProductsParams{
				Limit:  4,
				Offset: 3,
			},
			length: 3,
			idx:    3,
			err:    false,
		}, {
			desc: "success_empty",
			arg: product.ListProductsParams{
				Limit:  5,
				Offset: 1000,
			},
			length: 0,
			err:    false,
		}, {
			desc: "failed_minus_limit",
			arg: product.ListProductsParams{
				Limit:  -1,
				Offset: 0,
			},
			err: true,
		}, {
			desc: "failed_minus_offset",
			arg: product.ListProductsParams{
				Limit:  5,
				Offset: -50,
			},
			err: true,
		}, {
			desc: "failed_minus_arg",
			arg: product.ListProductsParams{
				Limit:  -1,
				Offset: -1,
			},
			err: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.ListProducts(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, tC.length, len(*res))
				for i := 0; i < len(*res); i++ {
					assert.Equal(t, input[tC.idx].Name, (*res)[i].Name)
					assert.Equal(t, input[tC.idx].Description, (*res)[i].Description)
					assert.Equal(t, input[tC.idx].Price, (*res)[i].Price)
					assert.Equal(t, input[tC.idx].Availability, (*res)[i].Availability)
					tC.idx++
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetTotalProducts(t *testing.T) {
	deleteProductsTest(t)

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

func TestUpdateProductAvailability(t *testing.T) {
	_, resProduct := createProductTest(t)
	t.Log("product:", resProduct)
	added := generator.RandomInt32(10, 50)
	substract := generator.RandomInt32(-10, -1)
	t.Log("added:", added)
	t.Log("substract:", substract)
	testCases := []struct {
		desc string
		arg  product.UpdateProductAvailabilityParams
		ans  product.Product
		err  bool
	}{
		{
			desc: "success_added",
			arg: product.UpdateProductAvailabilityParams{
				ID:           resProduct.ID,
				Availability: added,
			},
			ans: product.Product{
				ID:           resProduct.ID,
				Availability: resProduct.Availability + added,
			},
			err: false,
		}, {
			desc: "success_substract",
			arg: product.UpdateProductAvailabilityParams{
				ID:           resProduct.ID,
				Availability: substract,
			},
			ans: product.Product{
				ID:           resProduct.ID,
				Availability: resProduct.Availability + added + substract,
			},
			err: false,
		}, {
			desc: "failed_negative_availability",
			arg: product.UpdateProductAvailabilityParams{
				ID:           resProduct.ID,
				Availability: -(resProduct.Availability + added) * 2,
			},
			err: true,
		}, {
			desc: "failed_wrong_id",
			arg: product.UpdateProductAvailabilityParams{
				ID:           0,
				Availability: generator.RandomInt32(0, 50),
			},
			err: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res, err := repoTest.UpdateProductAvailability(ctx, tC.arg)
			if !tC.err {
				require.NoError(t, err)
				assert.Equal(t, tC.ans.Availability, res.Availability)
				t.Log("res:", res)
			} else {
				require.Error(t, err)
			}
		})
	}
}
