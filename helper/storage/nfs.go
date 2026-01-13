package storage

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"os/exec"
	"strings"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/spotlibs/go-lib/log"
)

func (h *nfsHelper) Upload(ctx context.Context, file filesystem.File, dirpath string) error {
	if h.err != nil {
		return h.err
	}
	info, err := os.Stat(dirpath)
	if err != nil || !info.IsDir() {
		err = os.MkdirAll(dirpath, 0664)
		if err != nil {
			log.Runtime(ctx).Error(log.Map{
				"message": "Failed to create directory for NFS upload",
				"dirpath": dirpath,
				"error":   err.Error(),
			})
			return err
		}
	}
	_, err = file.Store(dirpath)
	if err != nil {
		return err
	}
	return nil
}
func (h *nfsHelper) Move(ctx context.Context, srcPath string, destPath string) error {
	if h.err != nil {
		return h.err
	}
	if err := h.Copy(ctx, srcPath, destPath); err != nil {
		return err
	}
	return h.Delete(ctx, srcPath)
}
func (h *nfsHelper) Copy(ctx context.Context, srcPath string, destPath string) error {
	if h.err != nil {
		return h.err
	}
	_, err := os.Stat(srcPath)
	if err != nil {
		log.Runtime(ctx).Warning(log.Map{
			"message": "Source file not found for NFS copy",
			"srcPath": srcPath,
			"error":   err.Error(),
		})
		return err
	}
	temp := strings.Split(destPath, "/")
	if len(temp) > 1 {
		dirpath := strings.Join(temp[0:len(temp)-1], "/")
		info, err := os.Stat(dirpath)
		if err != nil || !info.IsDir() {
			err = os.MkdirAll(dirpath, 0664)
			if err != nil {
				log.Runtime(ctx).Error(log.Map{
					"message": "Failed to create directory for NFS copy",
					"dirpath": dirpath,
					"error":   err.Error(),
				})
				return err
			}
		}
	}
	err = exec.CommandContext(ctx, "cp", srcPath, destPath).Run()
	if err != nil {
		log.Runtime(ctx).Error(log.Map{
			"message":     "Failed to copy file in NFS",
			"source":      srcPath,
			"destination": destPath,
			"error":       err.Error(),
		})
		return err
	}
	return nil
}
func (h *nfsHelper) Delete(ctx context.Context, filepath string) error {
	if h.err != nil {
		return h.err
	}
	err := exec.CommandContext(ctx, "rm", filepath).Run()
	if err != nil {
		log.Runtime(ctx).Warning(log.Map{
			"message":  "Failed to delete file from NFS",
			"filepath": filepath,
			"error":    err.Error(),
		})
		return err
	}
	return nil
}
func (h *nfsHelper) Securelink(ctx context.Context, filepath string) (string, error) {
	if h.err != nil {
		return "", h.err
	}
	_, err := os.Stat(filepath)
	if err != nil {
		log.Runtime(ctx).Warning(log.Map{
			"message":  "File not found for secure link generation",
			"filepath": filepath,
			"error":    err.Error(),
		})
		return "", err
	}
	// Generate a secure link for the file
	linkname := randomString(40)
	err = exec.CommandContext(
		ctx,
		"ln", "-s", filepath, "/var/www/html/public/securelink/"+linkname+getFileExtension(filepath),
	).Run()
	if err != nil {
		return "", err
	}
	return linkname, nil
}

func (h *nfsHelper) SecurelinkFolder(ctx context.Context, dirpath string) (string, error) {
	if h.err != nil {
		return "", h.err
	}
	info, err := os.Stat(dirpath)
	if err != nil {
		log.Runtime(ctx).Warning(log.Map{
			"message":  "Directory not found for secure link generation",
			"filepath": dirpath,
			"error":    err.Error(),
		})
		return "", err
	} else if !info.IsDir() {
		err := errors.New("Directory not found for secure link generation")
		log.Runtime(ctx).Warning(log.Map{
			"message":  "Directory not found for secure link generation",
			"filepath": dirpath,
			"error":    err.Error(),
		})
		return "", err
	}
	// Generate a secure link for the file
	linkname := randomString(40)
	err = exec.CommandContext(ctx, "ln", "-sf", dirpath, "/var/www/html/public/securelink/"+linkname).Run()
	if err != nil {
		return "", err
	}
	return linkname, nil
}

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func getFileExtension(filepath string) string {
	tmp := strings.Split(filepath, "/")
	filename := tmp[len(tmp)-1]
	tmp = strings.Split(filename, ".")
	if len(tmp) > 0 {
		fileExtension := tmp[len(tmp)-1]
		return fileExtension
	}
	return "" // if only the file does not have any extension
}
