package types

import (
	"encoding/json"
	"strconv"

	"github.com/spf13/cast"
)

type Int interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr
}

// IntToString converts an integer value to a string.
func IntToString[T Int](i T) string {
	s := strconv.FormatInt(int64(i), 10)
	return s
}

// StringToInt converts a string to an integer value.
func StringToInt[T Int](s string) T {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return T(i)
}

// ValueOrDefault returns the provided value if it is not empty, otherwise returns the defaultValue.
func ValueOrDefault[T comparable](value, defaultValue T) T {
	var emptyValue T

	if value == emptyValue {
		return defaultValue
	}

	return value
}

// AnyToBool converts any value to a boolean using the cast library.
func AnyToBool(value any) bool {
	return cast.ToBool(value)
}

// Contains checks if a target value exists in a given list.
func Contains[T comparable](list []T, target T) bool {
	for _, element := range list {
		if target == element {
			return true
		}
	}

	return false
}

// TypeToPointerType converts a value to its pointer type.
func TypeToPointerType[T any](input T) *T {
	return &input
}

// PointerTypeToType converts a pointer value to its original type.
func PointerTypeToType[T any](input *T) T {
	var emptyValue T

	if input == nil {
		return emptyValue
	}

	return *input
}

// Unique returns unique value in as slice
func Unique[T comparable](elements []T) (result []T) {
	encountered := map[T]bool{}
	for idx := range elements {
		if _, ok := encountered[elements[idx]]; ok {
			continue
		}
		encountered[elements[idx]] = true
		result = append(result, elements[idx])
	}

	return result
}

// OffsetByType to get offset from page and limit, min value for page = 1
func OffsetByType[T Int](page, limit uint) T {
	offset := (page - 1) * limit
	if offset < 0 {
		return 0
	}

	return T(offset)
}

func ExpectedNumber[T Int](v any) T {
	var result T
	switch value := v.(type) {
	case int:
		result = T(value)
	case int8:
		result = T(value)
	case int16:
		result = T(value)
	case int32:
		result = T(value)
	case int64:
		result = T(value)
	case uint:
		result = T(value)
	case uint8:
		result = T(value)
	case uint16:
		result = T(value)
	case uint32:
		result = T(value)
	case uint64:
		result = T(value)
	case uintptr:
		result = T(value)
	case float32:
		result = T(value)
	case float64:
		result = T(value)
	case string:
		result = T(StringToInt[int64](value))
	default:
		result = 0
	}

	return T(result)
}

func MapToStruct[T any](value any) (item T, err error) {
	data, err := JSONMarshal(value)
	if err != nil {
		return item, err
	}

	if err := JSONUnmarshal(data, &item); err != nil {
		return item, err
	}

	return item, nil
}

// MapToStructOrDefault converts any value to T by marshalling to JSON then
// unmarshalling into T. Returns the zero value of T if conversion fails.
func MapToStructOrDefault[T any](value any) T {
	item, _ := MapToStruct[T](value)
	return item
}

func JSONUnmarshal(data []byte, v interface{}) error {
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return err
	}
	return nil
}

func JSONMarshal(data interface{}) ([]byte, error) {
	result, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return result, nil
}
