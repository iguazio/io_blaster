package Utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func CompareInterface(op string, a interface{}, b interface{}) (bool, error) {
	var a_value float64
	var b_value float64
	use_origin_value := false

	if _, ok := a.(int64); ok {
		a_value = float64(a.(int64))
	} else if _, ok := a.(float64); ok {
		a_value = a.(float64)
	} else {
		use_origin_value = true
	}

	if _, ok := b.(int64); ok {
		b_value = float64(b.(int64))
	} else if _, ok := b.(float64); ok {
		b_value = b.(float64)
	} else {
		use_origin_value = true
	}

	switch op {
	case "==":
		if use_origin_value {
			return a == b, nil
		} else {
			return a_value == b_value, nil
		}
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

func GenerateRandomBlob(length int) []byte {
	var seededRand *rand.Rand = GetSeededRandom()
	blob := make([]byte, length)
	for i := range blob {
		blob[i] = byte(seededRand.Intn(255))
	}

	return blob
}

func GenerateRandomBase64(src_blob_length int) string {
	blob := GenerateRandomBlob(src_blob_length)
	return base64.StdEncoding.EncodeToString(blob)
}

func ArrayFormat(format string, args []interface{}, arrayIndexesInArgs []int, arrayJoinString string) string {
	numberOfArrays := len(arrayIndexesInArgs)
	totalIterationsNeeded := 1
	currentArrayIndexes := make([]int, numberOfArrays)
	arrayLens := make([]int, numberOfArrays)
	for arrayIndex, arrayIndexInArgs := range arrayIndexesInArgs {
		arrayLens[arrayIndex] = len(args[arrayIndexInArgs].([]interface{}))
		totalIterationsNeeded *= arrayLens[arrayIndex]
	}

	argsForFormat := make([]interface{}, len(args))
	arrayIndexesInArgsIndex := arrayIndexesInArgs[0]
	for argIndex, arg := range args {
		if arrayIndexesInArgsIndex < numberOfArrays && argIndex == arrayIndexesInArgs[arrayIndexesInArgsIndex] {
			argsForFormat[argIndex] = arg.([]interface{})[0]
			arrayIndexesInArgsIndex++
		} else {
			argsForFormat[argIndex] = arg
		}
	}

	arrayFormatParts := make([]string, totalIterationsNeeded)
	for iterNum, currentArrayIndexLooping := 0, numberOfArrays-1; iterNum < totalIterationsNeeded; iterNum++ {
		for currentArrayIndexes[currentArrayIndexLooping] == arrayLens[currentArrayIndexLooping] {
			currentArrayIndexes[currentArrayIndexLooping] = 0
			argsForFormat[arrayIndexesInArgs[currentArrayIndexLooping]] = args[arrayIndexesInArgs[currentArrayIndexLooping]].([]interface{})[0]
			currentArrayIndexLooping--
			currentArrayIndexes[currentArrayIndexLooping]++

		}
		argsForFormat[arrayIndexesInArgs[currentArrayIndexLooping]] = args[arrayIndexesInArgs[currentArrayIndexLooping]].([]interface{})[currentArrayIndexes[currentArrayIndexLooping]]
		currentArrayIndexLooping = numberOfArrays - 1
		argsForFormat[arrayIndexesInArgs[currentArrayIndexLooping]] = args[arrayIndexesInArgs[currentArrayIndexLooping]].([]interface{})[currentArrayIndexes[currentArrayIndexLooping]]
		arrayFormatParts[iterNum] = fmt.Sprintf(format, argsForFormat...)
		currentArrayIndexes[currentArrayIndexLooping]++
	}

	return strings.Join(arrayFormatParts, arrayJoinString)
}
