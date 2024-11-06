package generator

import (
	"math/rand"
	"strconv"
	"strings"
)

const alphabet string = "abcdefghijklmnopqrstuvwxyz"

func RandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func RandomInt64(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func CreateRandomString(length int) string {
	var res strings.Builder

	alphabetLen := len(alphabet)
	for i := 0; i < length; i++ {
		char := alphabet[rand.Intn(alphabetLen)]
		res.WriteByte(char)
	}

	return res.String()
}

func CreateRandomEmail(name string) string {
	return name + "@mail.com"
}

func CreateRandomDate() string {
	year := strconv.Itoa(RandomInt(1900, 2007))
	month := strconv.Itoa(RandomInt(1, 12))
	day := strconv.Itoa(RandomInt(1, 30))

	return year + "-" + month + "-" + day
}

func CreateRandomGender() string {
	gender := []string{"male", "female"}

	numb := RandomInt(0, 1)
	return gender[numb]
}

func CreateRandomMaritalStatus() string {
	gender := []string{"single", "married", "divorced"}

	numb := RandomInt(0, 2)
	return gender[numb]
}

func CreateRandomMaritalStatusID() int64 {
	return RandomInt64(1, 3)

}
