package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomInt(t *testing.T) {
	tests := []struct {
		min int
		max int
	}{
		{
			min: 1,
			max: 5,
		}, {
			min: 5,
			max: 10,
		}, {
			min: 5,
			max: 15,
		}, {
			min: 70,
			max: 120,
		}, {
			min: 0,
			max: 100,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			res := RandomInt(test.min, test.max)
			assert.GreaterOrEqual(t, res, test.min)
			assert.LessOrEqual(t, res, test.max)
		})
	}
}

func TestCreateRandomName(t *testing.T) {
	length := []int{3, 5, 7, 4, 2}
	for _, v := range length {
		res := CreateRandomString(v)
		require.NotEmpty(t, res)
		assert.Equal(t, v, len(res))
	}
}

func TestCreateRandomEmail(t *testing.T) {
	names := []string{"john", "doe", "test"}

	for _, v := range names {
		res := CreateRandomEmail(v)
		assert.Equal(t, res, v+"@mail.com")
	}
}

func TestCreateRandomDate(t *testing.T) {
	for i := 0; i < 5; i++ {
		res := CreateRandomDate()
		require.NotEmpty(t, res)
		assert.GreaterOrEqual(t, len(res), 8)
		assert.LessOrEqual(t, len(res), 10)
	}
}

func TestCreateRandomGender(t *testing.T) {
	for i := 0; i < 5; i++ {
		res := CreateRandomGender()
		require.NotEmpty(t, res)
		if res != "male" && res != "female" {
			t.Error(res)
		}
	}
}

func TestCreateRandomMaritalStatus(t *testing.T) {
	for i := 0; i < 5; i++ {
		res := CreateRandomMaritalStatus()
		require.NotEmpty(t, res)
		if res != "single" && res != "married" && res != "divorced" {
			t.Error(res)
		}
	}
}

func TestCreateRandomMaritalStatusID(t *testing.T) {
	for i := 0; i < 5; i++ {
		res := CreateRandomMaritalStatusID()
		require.GreaterOrEqual(t, res, int64(1))
		require.LessOrEqual(t, res, int64(3))
	}
}
