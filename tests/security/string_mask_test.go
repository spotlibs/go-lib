package security_test

import (
	"os"
	"testing"

	"github.com/spotlibs/go-lib/security"
	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	os.Setenv("SECURITY_KEY", "abcdefgh12345678")
	chipertext, err := security.Encrypt("beautiful soup")
	assert.NoError(t, err)
	assert.True(t, len(chipertext) > 16)
}

func TestDecrypt(t *testing.T) {
	os.Setenv("SECURITY_KEY", "abcdefgh12345678")
	chipertext, err := security.Encrypt("beautiful soup")
	assert.NoError(t, err)
	plaintext, err := security.Decrypt(chipertext)
	assert.NoError(t, err)
	assert.Equal(t, "beautiful soup", plaintext)
}
