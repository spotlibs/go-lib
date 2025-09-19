package htmltopdf_test

import (
	"context"
	"testing"

	htmltopdf "github.com/spotlibs/go-lib/helper/html_to_pdf"
	"github.com/stretchr/testify/assert"
)

func TestConvertFromURL(t *testing.T) {
	err := htmltopdf.ConvertFromURL(
		context.Background(),
		"https://www.lipsum.com/",
		"/var/www/html/pdf/testt2.pdf",
	)
	assert.NoError(t, err)
}
