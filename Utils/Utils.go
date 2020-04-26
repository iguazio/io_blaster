package Utils

import (
	"errors"
	"math/rand"
	"time"
)

func CompareInterface(op string, a interface{}, b interface{}) (bool, error) {
	if op == "==" {
		return a == b, nil
	}

	var a_value float64
	var b_value float64

	if _, ok := a.(int64); ok {
		a_value = float64(a.(int64))
	} else {
		a_value = a.(float64)
	}

	if _, ok := b.(int64); ok {
		b_value = float64(b.(int64))
	} else {
		b_value = b.(float64)
	}

	switch op {
	case ">=":
		return a_value >= b_value, nil
	case "<=":
		return a_value >= b_value, nil
	case ">":
		return a_value >= b_value, nil
	case "<":
		return a_value >= b_value, nil
	default:
		return false, errors.New("Type assertion error")
	}
}

func GetSeededRandom() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GenerateRandomString(length int) string {
	var seededRand *rand.Rand = GetSeededRandom()
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	strBytes := make([]byte, length)
	for i := range strBytes {
		strBytes[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(strBytes)
}
