package htmltopdf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/minio/minio-go/v7"
)

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

func MinioExport(ctx context.Context, minioClient *minio.Client, content []byte, savepath string) error {
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
	pdfg.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader(content)))
	if err := pdfg.CreateContext(ctx); err != nil {
		return err
	}
	// log.Runtime(ctx).Info(log.Map{"message": parseFilePath(savepath)})
	err = exec.CommandContext(ctx, "mkdir", "-p", "/tmp"+parseFilePath(savepath)).Run()
	if err != nil {
		return err
	}
	if err := pdfg.WriteFile("/tmp" + savepath); err != nil {
		return err
	}
	upInfo, err := minioClient.FPutObject(ctx,
		os.Getenv("MINIO_BUCKET"),
		savepath,
		"/tmp"+savepath,
		minio.PutObjectOptions{ContentType: "application/pdf"},
	)
	if err != nil {
		return err
	}
	fmt.Println("file uploaded: ", upInfo.ChecksumSHA256)

	return nil
}

func parseFilePath(savepath string) string {
	temp := strings.Split(savepath, "/")
	return strings.Join(temp[:len(temp)-1], "/")
}

func NFSMinioExport(ctx context.Context, minioClient *minio.Client, content []byte, savepath string, nfspath string) error {
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
	pdfg.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader(content)))
	if err := pdfg.CreateContext(ctx); err != nil {
		return err
	}
	err = exec.CommandContext(ctx, "mkdir", "-p", "/tmp"+parseFilePath(savepath)).Run()
	if err != nil {
		return err
	}
	if err := pdfg.WriteFile("/tmp" + savepath); err != nil {
		return err
	}
	upInfo, err := minioClient.FPutObject(ctx,
		os.Getenv("MINIO_BUCKET"),
		savepath,
		"/tmp"+savepath,
		minio.PutObjectOptions{ContentType: "application/pdf"},
	)
	if err != nil {
		return err
	}
	fmt.Println("file uploaded: ", upInfo.ChecksumSHA256)
	err = exec.CommandContext(ctx, "mv", "/tmp"+savepath, nfspath+savepath).Run()
	if err != nil {
		return err
	}
	fmt.Println("file moved to NFS")

	return nil
}

func NFSExport(ctx context.Context, content []byte, savepath string, nfspath string) error {
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
	pdfg.AddPage(wkhtmltopdf.NewPageReader(bytes.NewReader(content)))
	if err := pdfg.CreateContext(ctx); err != nil {
		return err
	}

	return pdfg.WriteFile(nfspath + savepath)
}
