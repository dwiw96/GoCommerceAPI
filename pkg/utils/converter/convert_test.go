package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertStrToDate(t *testing.T) {
	tests := []struct {
		input string
		ans   string
	}{
		{
			input: "1990-12-30",
			ans:   "1990-12-30",
		}, {
			input: "1994-1-3",
			ans:   "1994-01-03",
		}, {
			input: "2001-3-29",
			ans:   "2001-03-29",
		}, {
			input: "2004-12-1",
			ans:   "2004-12-01",
		},
	}

	for _, test := range tests {
		res := ConvertStrToDate(test.input)

		date := res.Format("2006-01-02")
		assert.Equal(t, test.ans, date)
		t.Log(res)
	}
}

func TestConvertStrToInt(t *testing.T) {
	tests := []struct {
		input string
		ans   int
	}{
		{
			input: "0",
			ans:   0,
		}, {
			input: "50",
			ans:   50,
		}, {
			input: "324",
			ans:   324,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			res, err := ConvertStrToInt(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.ans, res)
		})
	}
}
