package fileupload

import (
	"bytes"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create mock file upload
func createMockFileUpload(filename string, content []byte) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", filename)
	part.Write(content)
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(10 << 20) // 10MB max

	if files, ok := form.File["file"]; ok && len(files) > 0 {
		return files[0]
	}
	return nil
}

func TestValidateImageUpload_ValidJPEG(t *testing.T) {
	// JPEG magic bytes
	jpegContent := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	fileHeader := createMockFileUpload("test.jpg", jpegContent)

	metadata, err := ValidateImageUpload(fileHeader)

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, "test.jpg", metadata.OriginalFilename)
	assert.Equal(t, "jpg", metadata.Format)
	assert.Equal(t, "image/jpeg", metadata.MimeType)
	assert.True(t, len(metadata.Filename) > 0)
	assert.NotEqual(t, "test.jpg", metadata.Filename) // Should be unique
}

func TestValidateImageUpload_ValidPNG(t *testing.T) {
	// PNG magic bytes
	pngContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	fileHeader := createMockFileUpload("logo.png", pngContent)

	metadata, err := ValidateImageUpload(fileHeader)

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, "logo.png", metadata.OriginalFilename)
	assert.Equal(t, "png", metadata.Format)
	assert.Equal(t, "image/png", metadata.MimeType)
}

func TestValidateImageUpload_RejectSVG(t *testing.T) {
	// SVG content (XML-based, contains potential XSS vector)
	svgContent := []byte(`<?xml version="1.0"?>
<svg xmlns="http://www.w3.org/2000/svg">
  <script>alert('XSS')</script>
</svg>`)
	fileHeader := createMockFileUpload("logo.svg", svgContent)

	metadata, err := ValidateImageUpload(fileHeader)

	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.ErrorIs(t, err, ErrInvalidFileType)
	assert.Contains(t, err.Error(), ".svg")
}

func TestValidateImageUpload_RejectGIF(t *testing.T) {
	// GIF magic bytes (also not allowed)
	gifContent := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}
	fileHeader := createMockFileUpload("image.gif", gifContent)

	metadata, err := ValidateImageUpload(fileHeader)

	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.ErrorIs(t, err, ErrInvalidFileType)
}

func TestValidateImageUpload_FileTooLarge(t *testing.T) {
	// Create file larger than 2MB
	largeContent := make([]byte, MaxImageSize+1024) // 2MB + 1KB
	// Add JPEG magic bytes
	copy(largeContent[0:3], []byte{0xFF, 0xD8, 0xFF})

	fileHeader := createMockFileUpload("large.jpg", largeContent)

	metadata, err := ValidateImageUpload(fileHeader)

	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.ErrorIs(t, err, ErrFileTooLarge)
	assert.Contains(t, err.Error(), "2MB")
}

func TestValidateImageUpload_EmptyFile(t *testing.T) {
	fileHeader := createMockFileUpload("empty.jpg", []byte{})

	metadata, err := ValidateImageUpload(fileHeader)

	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.ErrorIs(t, err, ErrEmptyFile)
}

func TestValidateImageUpload_NilFileHeader(t *testing.T) {
	metadata, err := ValidateImageUpload(nil)

	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.ErrorIs(t, err, ErrEmptyFile)
}

func TestValidateImageUpload_InvalidMagicBytes_JPEGExtensionPNGContent(t *testing.T) {
	// File claims to be JPEG but has PNG magic bytes (file type spoofing)
	pngContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	fileHeader := createMockFileUpload("fake.jpg", pngContent)

	metadata, err := ValidateImageUpload(fileHeader)

	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.ErrorIs(t, err, ErrInvalidMagicBytes)
	assert.Contains(t, err.Error(), "magic bytes don't match")
}

func TestValidateImageUpload_InvalidMagicBytes_PNGExtensionJPEGContent(t *testing.T) {
	// File claims to be PNG but has JPEG magic bytes
	jpegContent := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	fileHeader := createMockFileUpload("fake.png", jpegContent)

	metadata, err := ValidateImageUpload(fileHeader)

	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.ErrorIs(t, err, ErrInvalidMagicBytes)
}

func TestValidateImageUpload_SVGDisguisedAsJPEG(t *testing.T) {
	// Malicious SVG renamed to .jpg (XSS attack attempt)
	svgContent := []byte(`<svg onload="alert('XSS')"></svg>`)
	fileHeader := createMockFileUpload("malicious.jpg", svgContent)

	metadata, err := ValidateImageUpload(fileHeader)

	// Should fail because magic bytes don't match JPEG
	assert.Error(t, err)
	assert.Nil(t, metadata)
	assert.ErrorIs(t, err, ErrInvalidMagicBytes)
}

func TestValidateImageUpload_CaseInsensitiveExtension(t *testing.T) {
	// Test uppercase extensions
	jpegContent := []byte{0xFF, 0xD8, 0xFF, 0xE0}

	tests := []struct {
		filename string
	}{
		{"test.JPG"},
		{"test.Jpg"},
		{"test.PNG"},
		{"test.Png"},
		{"test.JPEG"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			fileHeader := createMockFileUpload(tt.filename, jpegContent)

			// Should handle case-insensitive extensions
			// Note: This will fail magic bytes check for PNG extensions with JPEG content
			// but should not fail on extension validation
			_, err := ValidateImageUpload(fileHeader)

			// Error should be about magic bytes, not invalid file type
			if err != nil {
				assert.NotErrorIs(t, err, ErrInvalidFileType)
			}
		})
	}
}

func TestValidateImageUpload_UniqueFilenames(t *testing.T) {
	jpegContent := []byte{0xFF, 0xD8, 0xFF, 0xE0}

	fileHeader1 := createMockFileUpload("logo.jpg", jpegContent)
	fileHeader2 := createMockFileUpload("logo.jpg", jpegContent)

	metadata1, err1 := ValidateImageUpload(fileHeader1)
	metadata2, err2 := ValidateImageUpload(fileHeader2)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Generated filenames should be unique even for same original filename
	assert.NotEqual(t, metadata1.Filename, metadata2.Filename)
	assert.Equal(t, metadata1.OriginalFilename, metadata2.OriginalFilename)
}

func TestReadImageContent(t *testing.T) {
	jpegContent := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	fileHeader := createMockFileUpload("test.jpg", jpegContent)

	content, err := ReadImageContent(fileHeader)

	assert.NoError(t, err)
	assert.NotNil(t, content)
	assert.Equal(t, len(jpegContent), len(content))
	assert.Equal(t, jpegContent, content)
}

func TestGetAllowedExtensions(t *testing.T) {
	extensions := GetAllowedExtensions()

	assert.NotEmpty(t, extensions)
	assert.Contains(t, extensions, "jpg")
	assert.Contains(t, extensions, "png")
	assert.NotContains(t, extensions, "svg")
	assert.NotContains(t, extensions, "gif")
}

func TestValidateMagicBytes(t *testing.T) {
	tests := []struct {
		name           string
		fileBytes      []byte
		expectedMagic  []byte
		shouldMatch    bool
	}{
		{
			name:          "JPEG magic bytes match",
			fileBytes:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10},
			expectedMagic: []byte{0xFF, 0xD8, 0xFF},
			shouldMatch:   true,
		},
		{
			name:          "PNG magic bytes match",
			fileBytes:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A},
			expectedMagic: []byte{0x89, 0x50, 0x4E, 0x47},
			shouldMatch:   true,
		},
		{
			name:          "Magic bytes mismatch",
			fileBytes:     []byte{0x89, 0x50, 0x4E, 0x47},
			expectedMagic: []byte{0xFF, 0xD8, 0xFF},
			shouldMatch:   false,
		},
		{
			name:          "File too short",
			fileBytes:     []byte{0xFF, 0xD8},
			expectedMagic: []byte{0xFF, 0xD8, 0xFF},
			shouldMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateMagicBytes(tt.fileBytes, tt.expectedMagic)
			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

func TestGenerateUniqueFilename(t *testing.T) {
	filename1 := generateUniqueFilename("logo.jpg")
	filename2 := generateUniqueFilename("logo.jpg")

	assert.NotEqual(t, filename1, filename2)
	assert.True(t, strings.HasSuffix(filename1, ".jpg"))
	assert.True(t, strings.HasSuffix(filename2, ".jpg"))
	assert.True(t, len(filename1) > 4) // More than just ".jpg"
}

// Benchmark tests
func BenchmarkValidateImageUpload_JPEG(b *testing.B) {
	jpegContent := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	fileHeader := createMockFileUpload("test.jpg", jpegContent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateImageUpload(fileHeader)
	}
}

func BenchmarkValidateImageUpload_PNG(b *testing.B) {
	pngContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	fileHeader := createMockFileUpload("test.png", pngContent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateImageUpload(fileHeader)
	}
}
