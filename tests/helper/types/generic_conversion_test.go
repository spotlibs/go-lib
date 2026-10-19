package types_test

import (
	"testing"

	"github.com/spotlibs/go-lib/helper/types"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// IntToString
// =============================================================================

func TestIntToString(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{name: "positive integer", input: 123, expected: "123"},
		{name: "zero", input: 0, expected: "0"},
		{name: "negative integer", input: -456, expected: "-456"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := types.IntToString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// =============================================================================
// StringToInt
// =============================================================================

func TestStringToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{name: "valid positive", input: "42", expected: 42},
		{name: "valid zero", input: "0", expected: 0},
		{name: "valid negative", input: "-99", expected: -99},
		{name: "invalid string returns zero", input: "abc", expected: 0},
		{name: "empty string returns zero", input: "", expected: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := types.StringToInt[int64](tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// =============================================================================
// ValueOrDefault
// =============================================================================

func TestValueOrDefault(t *testing.T) {
	t.Run("string non-empty returns value", func(t *testing.T) {
		result := types.ValueOrDefault("hello", "default")
		assert.Equal(t, "hello", result)
	})

	t.Run("string empty returns default", func(t *testing.T) {
		result := types.ValueOrDefault("", "default")
		assert.Equal(t, "default", result)
	})

	t.Run("int non-zero returns value", func(t *testing.T) {
		result := types.ValueOrDefault(5, 99)
		assert.Equal(t, 5, result)
	})

	t.Run("int zero returns default", func(t *testing.T) {
		result := types.ValueOrDefault(0, 99)
		assert.Equal(t, 99, result)
	})
}

// =============================================================================
// AnyToBool
// =============================================================================

func TestAnyToBool(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{name: "true bool", input: true, expected: true},
		{name: "false bool", input: false, expected: false},
		{name: "string true", input: "true", expected: true},
		{name: "string false", input: "false", expected: false},
		{name: "int 1", input: 1, expected: true},
		{name: "int 0", input: 0, expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := types.AnyToBool(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// =============================================================================
// Contains
// =============================================================================

func TestContains(t *testing.T) {
	t.Run("found in int slice", func(t *testing.T) {
		assert.True(t, types.Contains([]int{1, 2, 3}, 2))
	})

	t.Run("not found in int slice", func(t *testing.T) {
		assert.False(t, types.Contains([]int{1, 2, 3}, 99))
	})

	t.Run("found in string slice", func(t *testing.T) {
		assert.True(t, types.Contains([]string{"a", "b", "c"}, "b"))
	})

	t.Run("not found in string slice", func(t *testing.T) {
		assert.False(t, types.Contains([]string{"a", "b", "c"}, "z"))
	})

	t.Run("empty slice returns false", func(t *testing.T) {
		assert.False(t, types.Contains([]int{}, 1))
	})
}

// =============================================================================
// TypeToPointerType
// =============================================================================

func TestTypeToPointerType(t *testing.T) {
	t.Run("string to pointer", func(t *testing.T) {
		input := "hello"
		ptr := types.TypeToPointerType(input)
		assert.NotNil(t, ptr)
		assert.Equal(t, input, *ptr)
	})

	t.Run("int to pointer", func(t *testing.T) {
		input := 42
		ptr := types.TypeToPointerType(input)
		assert.NotNil(t, ptr)
		assert.Equal(t, input, *ptr)
	})
}

// =============================================================================
// PointerTypeToType
// =============================================================================

func TestPointerTypeToType(t *testing.T) {
	t.Run("non-nil pointer returns value", func(t *testing.T) {
		val := "world"
		result := types.PointerTypeToType(&val)
		assert.Equal(t, "world", result)
	})

	t.Run("nil pointer returns zero value", func(t *testing.T) {
		var ptr *string
		result := types.PointerTypeToType(ptr)
		assert.Equal(t, "", result)
	})

	t.Run("nil int pointer returns zero", func(t *testing.T) {
		var ptr *int
		result := types.PointerTypeToType(ptr)
		assert.Equal(t, 0, result)
	})
}

// =============================================================================
// Unique
// =============================================================================

func TestUnique(t *testing.T) {
	t.Run("removes duplicates from int slice", func(t *testing.T) {
		result := types.Unique([]int{1, 2, 2, 3, 3, 3})
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("removes duplicates from string slice", func(t *testing.T) {
		result := types.Unique([]string{"a", "b", "a", "c"})
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})

	t.Run("no duplicates stays same", func(t *testing.T) {
		result := types.Unique([]int{1, 2, 3})
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("empty slice returns nil result", func(t *testing.T) {
		result := types.Unique([]int{})
		assert.Empty(t, result)
	})
}

// =============================================================================
// OffsetByType
// =============================================================================

func TestOffsetByType(t *testing.T) {
	tests := []struct {
		name     string
		page     uint
		limit    uint
		expected int64
	}{
		{name: "page 1 limit 10 gives offset 0", page: 1, limit: 10, expected: 0},
		{name: "page 2 limit 10 gives offset 10", page: 2, limit: 10, expected: 10},
		{name: "page 3 limit 5 gives offset 10", page: 3, limit: 5, expected: 10},
		{name: "page 1 limit 0 gives offset 0", page: 1, limit: 0, expected: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := types.OffsetByType[int64](tc.page, tc.limit)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// =============================================================================
// ExpectedNumber
// =============================================================================

func TestExpectedNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int64
	}{
		{name: "int", input: int(10), expected: 10},
		{name: "int8", input: int8(8), expected: 8},
		{name: "int16", input: int16(16), expected: 16},
		{name: "int32", input: int32(32), expected: 32},
		{name: "int64", input: int64(64), expected: 64},
		{name: "uint", input: uint(5), expected: 5},
		{name: "uint8", input: uint8(8), expected: 8},
		{name: "uint16", input: uint16(16), expected: 16},
		{name: "uint32", input: uint32(32), expected: 32},
		{name: "uint64", input: uint64(64), expected: 64},
		{name: "uintptr", input: uintptr(7), expected: 7},
		{name: "float32", input: float32(3.9), expected: 3},
		{name: "float64", input: float64(9.1), expected: 9},
		{name: "string valid", input: "55", expected: 55},
		{name: "string invalid returns zero", input: "abc", expected: 0},
		{name: "default case (bool) returns zero", input: true, expected: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := types.ExpectedNumber[int64](tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// =============================================================================
// MapToStruct
// =============================================================================

type sampleStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestMapToStruct(t *testing.T) {
	t.Run("map converts to struct successfully", func(t *testing.T) {
		input := map[string]any{
			"name": "Alice",
			"age":  30,
		}
		result, err := types.MapToStruct[sampleStruct](input)
		assert.NoError(t, err)
		assert.Equal(t, "Alice", result.Name)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("invalid input (unmarshalable) returns error", func(t *testing.T) {
		// json.Marshal cannot marshal a channel
		input := make(chan int)
		_, err := types.MapToStruct[sampleStruct](input)
		assert.Error(t, err)
	})
}

// =============================================================================
// JSONMarshal
// =============================================================================

func TestJSONMarshal(t *testing.T) {
	t.Run("marshal valid struct", func(t *testing.T) {
		input := sampleStruct{Name: "Bob", Age: 25}
		data, err := types.JSONMarshal(input)
		assert.NoError(t, err)
		assert.Contains(t, string(data), "Bob")
	})

	t.Run("marshal invalid type returns error", func(t *testing.T) {
		input := make(chan int)
		_, err := types.JSONMarshal(input)
		assert.Error(t, err)
	})
}

// =============================================================================
// JSONUnmarshal
// =============================================================================

func TestJSONUnmarshal(t *testing.T) {
	t.Run("unmarshal valid JSON", func(t *testing.T) {
		data := []byte(`{"name":"Carol","age":22}`)
		var result sampleStruct
		err := types.JSONUnmarshal(data, &result)
		assert.NoError(t, err)
		assert.Equal(t, "Carol", result.Name)
		assert.Equal(t, 22, result.Age)
	})

	t.Run("unmarshal invalid JSON returns error", func(t *testing.T) {
		data := []byte(`not valid json`)
		var result sampleStruct
		err := types.JSONUnmarshal(data, &result)
		assert.Error(t, err)
	})
}

// =============================================================================
// MapToStructOrDefault
// =============================================================================

func TestMapToStructOrDefault(t *testing.T) {
	t.Run("map converts to struct successfully", func(t *testing.T) {
		input := map[string]any{
			"name": "Alice",
			"age":  30,
		}
		result := types.MapToStructOrDefault[sampleStruct](input)
		assert.Equal(t, "Alice", result.Name)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("nil input returns zero value", func(t *testing.T) {
		result := types.MapToStructOrDefault[sampleStruct](nil)
		assert.Equal(t, sampleStruct{}, result)
	})

	t.Run("invalid input returns zero value", func(t *testing.T) {
		result := types.MapToStructOrDefault[sampleStruct](make(chan int))
		assert.Equal(t, sampleStruct{}, result)
	})

	t.Run("partial fields returns partial struct", func(t *testing.T) {
		input := map[string]any{"name": "Bob"}
		result := types.MapToStructOrDefault[sampleStruct](input)
		assert.Equal(t, "Bob", result.Name)
		assert.Equal(t, 0, result.Age)
	})
}
