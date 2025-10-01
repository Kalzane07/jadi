package utils

import (
	"html"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

const (
	MaxUploadSize = 10 * 1024 * 1024 // 10 MB
)

// SanitizeInput performs a 3-step sanitization: URL Decode, HTML Unescape, and HTML tag stripping.
func SanitizeInput(input string) string {
	p := bluemonday.StrictPolicy()
	urlDecoded, _ := url.QueryUnescape(input)
	htmlUnescaped := html.UnescapeString(urlDecoded)
	return p.Sanitize(htmlUnescaped)
}

// ValidatePDFUpload checks if the uploaded file is a valid PDF and within the size limit.
func ValidatePDFUpload(c *gin.Context, file *multipart.FileHeader) bool {
	// Check size
	if file.Size > MaxUploadSize {
		return false
	}

	// Check content type
	src, err := file.Open()
	if err != nil {
		return false
	}
	defer src.Close()

	buffer := make([]byte, 512)
	if _, err := src.Read(buffer); err != nil {
		return false
	}

	return http.DetectContentType(buffer) == "application/pdf"
}
