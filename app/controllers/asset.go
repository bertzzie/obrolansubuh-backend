package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/revel/revel"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"
)

type Asset struct {
	GormController
}

type UploadedFile struct {
	Url         string `json:"url"`
	ContentType string `json:"type"`
	ContentSize int    `json:"size"`
	DeleteURL   string `json:"deleteUrl"`
	DeleteType  string `json:"deleteType"`
}

type FileUploadResponse struct {
	Files []UploadedFile `json:"files"`
}

type FailedUpload struct {
	Name  string `json:"name"`
	Size  int    `json:"size"`
	Error string `json:"error"`
}

type FileUploadError struct {
	Files []FailedUpload `json:"files"`
}

func (c Asset) ImageUpload(image []byte) revel.Result {
	fileType := c.Params.Files["image"][0].Header["Content-Type"]
	fileName := c.Params.Files["image"][0].Filename

	hostname := revel.Config.StringDefault("server.hostname", "http://localhost:9000")
	uploadDr := revel.Config.StringDefault("upload.image.location", "/public/upload/")

	hashName := hashFileName(fileName, c.Session["user"])

	fullName := uploadDr + hashName

	if err := saveFile(image, revel.BasePath+fullName); err != nil {
		revel.ERROR.Printf("[LGFATAL] Failed to write uploaded file %s by %s. Error: %s",
			revel.BasePath+fullName,
			c.Session["user"],
			err)

		failedUpload := FailedUpload{
			Name:  fullName,
			Size:  len(image),
			Error: c.Message("errors.upload.image"),
		}

		FUR := FileUploadError{
			Files: []FailedUpload{failedUpload},
		}

		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(FUR)
	} else {
		fileInfo := UploadedFile{
			Url:         hostname + fullName,
			ContentType: fileType[0],
			ContentSize: len(image),
			DeleteURL:   hostname + "/post/image/delete/" + fileName,
			DeleteType:  "DELETE",
		}

		FUR := FileUploadResponse{
			Files: []UploadedFile{fileInfo},
		}

		return c.RenderJson(FUR)
	}
}

func saveFile(file []byte, destination string) error {
	err := ioutil.WriteFile(destination, file, 0644) // Permission: -rw-r--r--

	if err != nil {
		return err
	}

	return nil
}

// Hash function to change filename so we always have unique
// filename for uploaded file.
//
// Hash function is a simple
func hashFileName(filename string, username string) string {
	ext := filepath.Ext(filename)

	fullname := filename + "_" + username + "_" + time.Now().Format("20060102150405")
	sum := md5.Sum([]byte(fullname))
	hash := hex.EncodeToString(sum[:])

	result := hash + ext

	return result

}
