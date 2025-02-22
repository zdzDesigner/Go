package lang

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

/**
 * @ToInt 转换为int
 * @param {arg:interface}
 * @return: int
 */
func ToInt(arg any) (d int) {

	if arg != nil {
		var tmp = reflect.ValueOf(arg).Interface()

		switch v := tmp.(type) {
		case json.Number:
			n, err := v.Float64()
			if err != nil {
				d = 0
			} else {
				d = int(n)
			}
		case string:
			d, _ = strconv.Atoi(v)
		case float64:
			d = int(v)
		case float32:
			d = int(v)
		case int:
			d = v
		}
	}
	return d
}

// ToFloat64 ..
func ToFloat64(arg any) (d float64) {
	if arg != nil {
		var tmp = reflect.ValueOf(arg).Interface()

		switch v := tmp.(type) {
		case json.Number:
			n, err := v.Float64()
			if err != nil {
				d = 0
			}
			d = n
		case string:
			d, _ = strconv.ParseFloat(v, 64)
		case float32:
			d = float64(v)
		case int:
			d = float64(v)
		case float64:
			d = v
		}
	}
	return d
}

// Float64Digit .. 保留小数点位数
func Float64Digit(source float64, digit int) float64 {
	value, _ := strconv.ParseFloat(fmt.Sprintf(fmt.Sprintf("%%.%df", digit), source), 64)
	return value
}

func ToString(arg any) (str string) {

	if arg != nil {
		var tmp = reflect.ValueOf(arg).Interface()
		switch v := tmp.(type) {
		case int:
			str = strconv.Itoa(v)
		case int8:
		case int16:
		case int32:
			str = strconv.FormatInt(int64(v), 10)
		case int64:
			str = strconv.FormatInt(v, 10)
		case string:
			str = v
		case float32:
			str = strconv.FormatFloat(float64(v), 'f', -1, 32)
		case float64:
			str = strconv.FormatFloat(v, 'f', -1, 64)
		case fmt.Stringer:
			str = v.String()
		case reflect.Value:
			str = ToString(v.Interface())

		}
	}
	return str
}


func IsEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return v.Bool() == false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}

