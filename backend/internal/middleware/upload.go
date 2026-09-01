// File upload handling: enforces a MIME allow-list and size limit, reading the
// file fully into memory (nothing is written to local disk — matches the
// original multer.memoryStorage() configuration).
package middleware

import (
	"io"

	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/apperror"
)

type UploadedFile struct {
	Buffer      []byte
	Filename    string
	ContentType string
	Size        int64
}

const uploadedFileContextKey = "uploadedFile"

// RequireSingleImage parses a single multipart file field, enforcing
// maxSizeBytes and the allowed MIME list, and stores it on the context.
// fieldName's absence is NOT an error here — handlers decide whether the file
// is mandatory (mirrors evidence uploads, which are optional).
func RequireSingleImage(fieldName string, maxSizeBytes int64, allowedMIME []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSizeBytes+1<<20 { // small slack for multipart overhead
			_ = c.Error(apperror.PayloadTooLarge("Image exceeds size limit"))
			c.Abort()
			return
		}

		fileHeader, err := c.FormFile(fieldName)
		if err != nil {
			// No file provided — let the handler decide if that's an error.
			c.Next()
			return
		}

		if fileHeader.Size > maxSizeBytes {
			_ = c.Error(apperror.PayloadTooLarge("Image exceeds size limit"))
			c.Abort()
			return
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if !containsMIME(allowedMIME, contentType) {
			_ = c.Error(apperror.UnsupportedMediaType("Unsupported image type: " + contentType))
			c.Abort()
			return
		}

		f, err := fileHeader.Open()
		if err != nil {
			_ = c.Error(apperror.BadRequest("Could not read uploaded file"))
			c.Abort()
			return
		}
		defer f.Close()

		buf, err := io.ReadAll(f)
		if err != nil {
			_ = c.Error(apperror.BadRequest("Could not read uploaded file"))
			c.Abort()
			return
		}

		c.Set(uploadedFileContextKey, UploadedFile{
			Buffer:      buf,
			Filename:    fileHeader.Filename,
			ContentType: contentType,
			Size:        fileHeader.Size,
		})
		c.Next()
	}
}

func UploadedFileFromContext(c *gin.Context) (UploadedFile, bool) {
	v, ok := c.Get(uploadedFileContextKey)
	if !ok {
		return UploadedFile{}, false
	}
	f, ok := v.(UploadedFile)
	return f, ok
}

func containsMIME(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
