package helper_test

import (
	"testing"

	"github.com/spotlibs/go-lib/helper"
	"github.com/stretchr/testify/assert"
)

func TestNewStringSet(t *testing.T) {
	set := helper.NewStringSet()
	assert.NotNil(t, set)
	assert.Equal(t, 0, set.Size())
}

func TestNewStringSetFromSlice(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	set := helper.NewStringSetFromSlice(items)

	assert.Equal(t, 3, set.Size())
	assert.True(t, set.Exists("apple"))
	assert.True(t, set.Exists("banana"))
	assert.True(t, set.Exists("cherry"))
}

func TestNewStringSetFromSlice_Empty(t *testing.T) {
	set := helper.NewStringSetFromSlice([]string{})
	assert.Equal(t, 0, set.Size())
}

func TestNewStringSetFromSlice_Duplicates(t *testing.T) {
	items := []string{"apple", "apple", "banana"}
	set := helper.NewStringSetFromSlice(items)

	assert.Equal(t, 2, set.Size())
	assert.True(t, set.Exists("apple"))
	assert.True(t, set.Exists("banana"))
}

func TestStringSet_Add(t *testing.T) {
	set := helper.NewStringSet()

	set.Add("test")
	assert.Equal(t, 1, set.Size())
	assert.True(t, set.Exists("test"))
}

func TestStringSet_Add_Multiple(t *testing.T) {
	set := helper.NewStringSet()

	set.Add("first")
	set.Add("second")
	set.Add("third")

	assert.Equal(t, 3, set.Size())
	assert.True(t, set.Exists("first"))
	assert.True(t, set.Exists("second"))
	assert.True(t, set.Exists("third"))
}

func TestStringSet_Add_Duplicate(t *testing.T) {
	set := helper.NewStringSet()

	set.Add("duplicate")
	set.Add("duplicate")

	assert.Equal(t, 1, set.Size())
}

func TestStringSet_Exists(t *testing.T) {
	set := helper.NewStringSet()
	set.Add("exists")

	assert.True(t, set.Exists("exists"))
	assert.False(t, set.Exists("notexists"))
}

func TestStringSet_Exists_EmptyString(t *testing.T) {
	set := helper.NewStringSet()
	set.Add("")

	assert.True(t, set.Exists(""))
}

func TestStringSet_Size(t *testing.T) {
	set := helper.NewStringSet()
	assert.Equal(t, 0, set.Size())

	set.Add("one")
	assert.Equal(t, 1, set.Size())

	set.Add("two")
	assert.Equal(t, 2, set.Size())

	set.Add("one") // duplicate
	assert.Equal(t, 2, set.Size())
}

func TestStringSet_ChineseCharacters(t *testing.T) {
	set := helper.NewStringSet()
	set.Add("你好")
	set.Add("世界")

	assert.Equal(t, 2, set.Size())
	assert.True(t, set.Exists("你好"))
	assert.True(t, set.Exists("世界"))
	assert.False(t, set.Exists("测试"))
}
