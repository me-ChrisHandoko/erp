// Package fileupload provides secure file upload utilities with validation
package fileupload

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/lucsky/cuid"
)

const (
	// MaxImageSize defines maximum allowed image size (2MB)
	MaxImageSize = 2 * 1024 * 1024 // 2MB in bytes
)

var (
	// ErrInvalidFileType is returned when file type is not allowed
	ErrInvalidFileType = errors.New("invalid file type - only JPG and PNG are allowed")
	// ErrFileTooLarge is returned when file exceeds size limit
	ErrFileTooLarge = errors.New("file size exceeds 2MB limit")
	// ErrInvalidMagicBytes is returned when magic bytes don't match extension
	ErrInvalidMagicBytes = errors.New("file magic bytes don't match declared type")
	// ErrEmptyFile is returned when file is empty
	ErrEmptyFile = errors.New("file is empty")
)

// ImageMetadata contains information about uploaded image
type ImageMetadata struct {
	OriginalFilename string    `json:"originalFilename"`
	Filename         string    `json:"filename"`
	Size             int64     `json:"size"`
	Format           string    `json:"format"`
	MimeType         string    `json:"mimeType"`
	UploadedAt       time.Time `json:"uploadedAt"`
}

// AllowedImageType represents allowed image types
type AllowedImageType struct {
	Extension  string
	MimeType   string
	MagicBytes []byte
}

var (
	// AllowedImageTypes defines allowed image types with their magic bytes
	// SVG is explicitly EXCLUDED to prevent XSS attacks
	AllowedImageTypes = []AllowedImageType{
		{
			Extension:  ".jpg",
			MimeType:   "image/jpeg",
			MagicBytes: []byte{0xFF, 0xD8, 0xFF}, // JPEG magic bytes
		},
		{
			Extension:  ".jpeg",
			MimeType:   "image/jpeg",
			MagicBytes: []byte{0xFF, 0xD8, 0xFF}, // JPEG magic bytes
		},
		{
			Extension:  ".png",
			MimeType:   "image/png",
			MagicBytes: []byte{0x89, 0x50, 0x4E, 0x47}, // PNG magic bytes
		},
	}
)

// ValidateImageUpload validates uploaded image file for security
// This function performs multiple security checks:
// 1. File size validation (max 2MB)
// 2. File type validation (only JPG/PNG allowed, SVG BLOCKED)
// 3. Magic bytes validation (prevent file type spoofing)
func ValidateImageUpload(fileHeader *multipart.FileHeader) (*ImageMetadata, error) {
	// Check if file exists
	if fileHeader == nil {
		return nil, ErrEmptyFile
	}

	// Validate file size
	if fileHeader.Size == 0 {
		return nil, ErrEmptyFile
	}

	if fileHeader.Size > MaxImageSize {
		return nil, fmt.Errorf("%w: %d bytes (max: %d bytes)", ErrFileTooLarge, fileHeader.Size, MaxImageSize)
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))

	// Check if extension is allowed
	allowedType, found := findAllowedType(ext)
	if !found {
		return nil, fmt.Errorf("%w: %s (allowed: jpg, jpeg, png)", ErrInvalidFileType, ext)
	}

	// Open file to read magic bytes
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Read first bytes for magic byte validation
	magicBytes := make([]byte, 512) // Read first 512 bytes for detection
	n, err := file.Read(magicBytes)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}
	magicBytes = magicBytes[:n]

	// Validate magic bytes match declared type
	if !validateMagicBytes(magicBytes, allowedType.MagicBytes) {
		return nil, fmt.Errorf("%w: declared as %s but magic bytes don't match", ErrInvalidMagicBytes, ext)
	}

	// Generate unique filename
	uniqueFilename := generateUniqueFilename(fileHeader.Filename)

	// Create metadata
	metadata := &ImageMetadata{
		OriginalFilename: fileHeader.Filename,
		Filename:         uniqueFilename,
		Size:             fileHeader.Size,
		Format:           strings.TrimPrefix(ext, "."),
		MimeType:         allowedType.MimeType,
		UploadedAt:       time.Now(),
	}

	return metadata, nil
}

// ReadImageContent reads the entire image file content for storage
// Call this after ValidateImageUpload succeeds
func ReadImageContent(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read entire file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return content, nil
}

// findAllowedType checks if extension is in allowed types
func findAllowedType(ext string) (AllowedImageType, bool) {
	for _, allowedType := range AllowedImageTypes {
		if allowedType.Extension == ext {
			return allowedType, true
		}
	}
	return AllowedImageType{}, false
}

// validateMagicBytes checks if file's magic bytes match expected type
func validateMagicBytes(fileBytes, expectedMagic []byte) bool {
	if len(fileBytes) < len(expectedMagic) {
		return false
	}

	// Check if file starts with expected magic bytes
	return bytes.HasPrefix(fileBytes, expectedMagic)
}

// generateUniqueFilename creates unique filename using CUID
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	uniqueID := cuid.New()
	return fmt.Sprintf("%s%s", uniqueID, ext)
}

// GetAllowedExtensions returns human-readable list of allowed extensions
func GetAllowedExtensions() []string {
	extensions := make([]string, 0, len(AllowedImageTypes))
	seen := make(map[string]bool)

	for _, allowedType := range AllowedImageTypes {
		ext := strings.TrimPrefix(allowedType.Extension, ".")
		if !seen[ext] {
			extensions = append(extensions, ext)
			seen[ext] = true
		}
	}

	return extensions
}
