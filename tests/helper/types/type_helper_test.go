package types_test

import (
	"testing"

	"github.com/goravel/framework/support/carbon"
	"github.com/spotlibs/go-lib/helper/types"
	"github.com/stretchr/testify/assert"
)

func TestPtr(t *testing.T) {
	s := "hello world"
	r := types.Ptr(s)
	assert.Equal(t, s, *r)
	x := 24
	y := types.Ptr(x)
	assert.Equal(t, x, *y)
}

// =============================================================================
// PtrFormat
// =============================================================================

func TestPtrFormat(t *testing.T) {
	t.Run("non-nil pointer calls formatter and returns result", func(t *testing.T) {
		val := "hello"
		result := types.PtrFormat(&val, func(s string) string {
			return s + "_formatted"
		})
		assert.NotNil(t, result)
		assert.Equal(t, "hello_formatted", *result)
	})

	t.Run("nil pointer returns nil", func(t *testing.T) {
		var val *string
		result := types.PtrFormat(val, func(s string) string {
			return s + "_formatted"
		})
		assert.Nil(t, result)
	})

	t.Run("int pointer formats correctly", func(t *testing.T) {
		val := 42
		result := types.PtrFormat(&val, func(n int) string {
			return "number_42"
		})
		assert.NotNil(t, result)
		assert.Equal(t, "number_42", *result)
	})

	t.Run("formatter returning empty string still returns non-nil pointer", func(t *testing.T) {
		val := "anything"
		result := types.PtrFormat(&val, func(s string) string {
			return ""
		})
		assert.NotNil(t, result)
		assert.Equal(t, "", *result)
	})
}

// =============================================================================
// PtrDateTimeString
// =============================================================================

func TestPtrDateTimeString(t *testing.T) {
	t.Run("nil pointer returns nil", func(t *testing.T) {
		result := types.PtrDateTimeString(nil)
		assert.Nil(t, result)
	})

	t.Run("zero date returns nil", func(t *testing.T) {
		zero := &carbon.DateTime{}
		result := types.PtrDateTimeString(zero)
		assert.Nil(t, result)
	})

	t.Run("valid date returns formatted string", func(t *testing.T) {
		dt := &carbon.DateTime{Carbon: carbon.Parse("2024-08-10 09:30:00")}
		result := types.PtrDateTimeString(dt)
		assert.NotNil(t, result)
		assert.Equal(t, "2024-08-10 09:30:00", *result)
	})
}
