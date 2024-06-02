package logic

import (
	"github.com/gin-gonic/gin"
	"ichat-go/config"
	"ichat-go/errs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func fullPath(path string) string {
	return filepath.Join(config.App.UploadDir, path)
}

func FileUpload(c *gin.Context) string {
	file, err := c.FormFile("file")
	if err != nil {
		panic(errs.NewAppError(errs.CodeBadRequest, "文件参数错误"))
	}
	if file.Size > 1024*1024*10 {
		panic(errs.NewAppError(errs.CodeBadRequest, "文件大小不能超过10M"))
	}
	ext := filepath.Ext(file.Filename)
	savePath := ""
	for {
		savePath = time.Now().Format("2006-01-02T150405.000") + ext
		if !fileExists(fullPath(savePath)) {
			break
		}
		time.Sleep(time.Millisecond)
	}
	//time.Sleep(time.Second * 3)
	if err := c.SaveUploadedFile(file, fullPath(savePath)); err != nil {
		panic(errs.SaveFileFailed)
	}
	return "file/" + savePath
}

func checkValid(path string) string {
	root, _ := filepath.Abs(fullPath(""))
	abs, _ := filepath.Abs(fullPath(path))
	if !strings.HasPrefix(abs, root) {
		panic(errs.Forbidden)
	}
	return abs
}

func FileDownload(path string, c *gin.Context) {
	abs := checkValid(path)
	c.File(abs)
	c.Done()
}
