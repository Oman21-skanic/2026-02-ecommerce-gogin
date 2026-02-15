package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/example/ecommerce-api/internal/config"
)

type UploadsHandler struct {
	cfg *config.Config
}

func NewUploadsHandler(cfg *config.Config) *UploadsHandler {
	return &UploadsHandler{cfg: cfg}
}

func (h *UploadsHandler) UploadProductThumbnail(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file diperlukan"})
		return
	}

	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ukuran file maksimal 2MB"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "format file harus jpg, jpeg, png, atau webp"})
		return
	}

	if err := os.MkdirAll("uploads", 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal membuat folder upload"})
		return
	}

	filename := uuid.NewString() + ext
	path := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal menyimpan file"})
		return
	}

	baseURL := strings.TrimRight(h.cfg.BaseURL, "/")
	c.JSON(http.StatusOK, gin.H{
		"url": baseURL + "/uploads/" + filename,
	})
}
