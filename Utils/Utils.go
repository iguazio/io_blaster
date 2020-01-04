package Utils

import (
	"errors"
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
