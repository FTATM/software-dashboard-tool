package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/FTATM/software-dashboard-tool/internal/model"
)

type S3FileHandler struct {
	s3Client model.S3Client
}

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// ⚡ NEW: Pass the pre-initialized client into the handler
func NewS3FileHandler(s3Client model.S3Client) *S3FileHandler {
	return &S3FileHandler{
		s3Client: s3Client,
	}
}

func (h *S3FileHandler) UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	var res Response

	// 1. Limit max body size (10MB)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	// 2. Parse form file
	file, _, err := r.FormFile("image")
	if err != nil {
		res.Message = "Invalid file or file size exceeds 10MB"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	defer file.Close()

	// 3. Sniff actual file content
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		res.Message = "Failed to inspect file"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	// 4. Detect true MIME type
	detectedType := http.DetectContentType(buffer[:n])
	safeExt, ok := allowedImageTypes[detectedType]
	if !ok {
		res.Message = "Only image files (PNG, JPG, GIF, WEBP) are allowed"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// 5. Seek back to start
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		res.Message = "Failed to process image stream"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	// 6. Generate secure filename
	newFileName := uuid.New().String() + safeExt

	// 7. ⚡ NEW: Delegate the actual upload to the client!
	err = h.s3Client.UploadImage(r.Context(), file, newFileName, detectedType)
	if err != nil {
		res.Message = "Failed to save image"
		respondJson(w, http.StatusInternalServerError, &res)
		return
	}

	publicUrl := fmt.Sprintf("/file/image/%s", newFileName)
	res.Data = map[string]any{
		"url": publicUrl,
	}
	respondJson(w, http.StatusOK, &res)
}

func (h *S3FileHandler) GetImageHandler(w http.ResponseWriter, r *http.Request) {
	var res Response
	filename := r.PathValue("filename")
	if filename == "" {
		res.Message = model.ErrInvalidBody.Error()
		respondJson(w, http.StatusBadRequest, &res)
		return
	}

	// ⚡ NEW: Delegate fetching to the client!
	fileStream, contentType, err := h.s3Client.GetImage(r.Context(), filename)
	if err != nil {
		res.Message = "Image not found"
		respondJson(w, http.StatusBadRequest, &res)
		return
	}
	defer fileStream.Close()

	// Set headers and stream the file directly to the client
	if contentType != nil {
		w.Header().Set("Content-Type", *contentType)
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	io.Copy(w, fileStream)
}
