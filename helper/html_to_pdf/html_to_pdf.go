package htmltopdf

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/minio/minio-go/v7"
)

type PDFOptions struct {
	MarginTop    uint
	MarginBottom uint
	MarginLeft   uint
	MarginRight  uint
	PaddingLeft  uint
	HeaderPath   string
	FooterPath   string
}

func ConvertFromURL(ctx context.Context, url, savepath string) error {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Dpi.Set(600)
	pdfg.NoCollate.Set(false)
	pdfg.MarginTop.Set(0)
	pdfg.MarginRight.Set(0)
	pdfg.MarginBottom.Set(0)
	pdfg.MarginLeft.Set(0)
	page := wkhtmltopdf.NewPage(url)
	page.DisableSmartShrinking.Set(true)
	pdfg.AddPage(page)
	if err := pdfg.CreateContext(ctx); err != nil {
		return err
	}

	return pdfg.WriteFile(savepath)
}

func MinioExport(ctx context.Context, minioClient *minio.Client, content []byte, savepath string, options ...PDFOptions) error {
	var option PDFOptions
	if len(options) > 0 {
		option = options[0]
	}
	setDefaultMargin(&option)
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Dpi.Set(600)
	pdfg.NoCollate.Set(false)
	pdfg.MarginTop.Set(option.MarginTop)
	pdfg.MarginRight.Set(option.MarginRight)
	pdfg.MarginBottom.Set(option.MarginBottom)
	pdfg.MarginLeft.Set(option.MarginLeft)
	page := wkhtmltopdf.NewPageReader(bytes.NewReader(content))
	page.HeaderHTML.Set(option.HeaderPath)
	page.FooterHTML.Set(option.FooterPath)
	pdfg.AddPage(page)
	if err := pdfg.CreateContext(ctx); err != nil {
		return err
	}
	upInfo, err := minioClient.PutObject(ctx,
		os.Getenv("MINIO_BUCKET"),
		savepath,
		bytes.NewBuffer(pdfg.Bytes()),
		int64(len(pdfg.Bytes())),
		minio.PutObjectOptions{ContentType: "application/pdf"},
	)
	if err != nil {
		return err
	}
	fmt.Println("file uploaded: ", upInfo.ChecksumSHA256)

	return nil
}

func NFSMinioExport(ctx context.Context, minioClient *minio.Client, content []byte, savepath string, nfspath string, options ...PDFOptions) error {
	var option PDFOptions
	if len(options) > 0 {
		option = options[0]
	}
	setDefaultMargin(&option)
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Dpi.Set(600)
	pdfg.NoCollate.Set(false)
	pdfg.MarginTop.Set(option.MarginTop)
	pdfg.MarginRight.Set(option.MarginRight)
	pdfg.MarginBottom.Set(option.MarginBottom)
	pdfg.MarginLeft.Set(option.MarginLeft)
	page := wkhtmltopdf.NewPageReader(bytes.NewReader(content))
	page.HeaderHTML.Set(option.HeaderPath)
	page.FooterHTML.Set(option.FooterPath)
	pdfg.AddPage(page)
	if err := pdfg.CreateContext(ctx); err != nil {
		return err
	}
	upInfo, err := minioClient.PutObject(ctx,
		os.Getenv("MINIO_BUCKET"),
		savepath,
		bytes.NewBuffer(pdfg.Bytes()),
		int64(len(pdfg.Bytes())),
		minio.PutObjectOptions{ContentType: "application/pdf"},
	)
	if err != nil {
		return err
	}
	fmt.Println("file uploaded: ", upInfo.ChecksumSHA256)
	if err := pdfg.WriteFile(nfspath + savepath); err != nil {
		return err
	}
	fmt.Println("file written to NFS")

	return nil
}

func NFSExport(ctx context.Context, content []byte, savepath string, nfspath string, options ...PDFOptions) error {
	var option PDFOptions
	if len(options) > 0 {
		option = options[0]
	}
	setDefaultMargin(&option)
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Dpi.Set(600)
	pdfg.NoCollate.Set(false)
	pdfg.MarginTop.Set(option.MarginTop)
	pdfg.MarginRight.Set(option.MarginRight)
	pdfg.MarginBottom.Set(option.MarginBottom)
	pdfg.MarginLeft.Set(option.MarginLeft)
	page := wkhtmltopdf.NewPageReader(bytes.NewReader(content))
	page.HeaderHTML.Set(option.HeaderPath)
	page.FooterHTML.Set(option.FooterPath)
	pdfg.AddPage(page)
	if err := pdfg.CreateContext(ctx); err != nil {
		return err
	}

	return pdfg.WriteFile(nfspath + savepath)
}

func setDefaultMargin(option *PDFOptions) {
	option.MarginBottom = 10
	option.MarginTop = 10
	option.MarginLeft = 10
	option.MarginRight = 10
}
