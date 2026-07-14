package types_test

import (
	"testing"

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
