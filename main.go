package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func init() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Failed to load .env file")
	}
}

func main() {
	port := os.Getenv("PORT")

	e := echo.New()

	e.POST("/upload", handleFileUpload)
	e.Static("/images", "uploads")

	e.Logger.Fatal(e.Start(":" + port))
}

func handleFileUpload(c echo.Context) error {
	file, err := c.FormFile("file")

	// check if file key exists
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// check if file can be open
	src, err := file.Open()

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	defer src.Close()

	// read first 512 bytes to detect content type
	buf := make([]byte, 512)
	src.Read(buf)
	contentType := http.DetectContentType(buf)

	src.Seek(0, 0)

	// allow only PNG
	if contentType != "image/png" {
		log.Println("Unsupported file type:", contentType)
		return c.String(http.StatusUnsupportedMediaType, "only PNG allowed")
	}

	// create directory for uploaded file
	os.MkdirAll("uploads", os.ModePerm)

	dstPath := filepath.Join("uploads", file.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	defer dst.Close()
	url := fmt.Sprintf("/images/%s", file.Filename)

	return c.JSON(http.StatusOK, echo.Map{
		"image_path": url,
	})
}
