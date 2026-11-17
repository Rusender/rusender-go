package rusender

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// FileInfo is an account file.
type FileInfo struct {
	ID       int    `json:"id"`
	FileName string `json:"fileName"`
	// Size in bytes.
	Size      int       `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

// UploadFile is a file to upload. The server detects the file type from the contents.
type UploadFile struct {
	FileName string
	Content  io.Reader
}

// ListFilesParams are the parameters of [FilesService.List].
type ListFilesParams struct {
	PageParams
	// Search searches by name.
	Search string `json:"search,omitempty"`
}

// FilesService: account files, campaign attachments.
type FilesService struct{ t *transport }

// List lists account files, newest first, searchable by name.
// Scope: files.read.
func (s *FilesService) List(ctx context.Context, params *ListFilesParams, opts ...RequestOption) (*Page[FileInfo], error) {
	return listPage[FileInfo](ctx, s.t, request{method: "GET", path: "/v1/public/files", query: params}, opts)
}

// Upload uploads campaign attachments in one request: 1–20 files, 5 MB in total. The set is validated as a
// whole: if any file fails validation (type, size, count), none is saved and the method returns ErrValidation —
// the rejected file is in Details. Pass the returned IDs in campaign Attachments.
// Not retried on failure: a retry would create duplicate files.
// Scope: files.write.
func (s *FilesService) Upload(ctx context.Context, files []UploadFile, opts ...RequestOption) ([]FileInfo, error) {
	var buf bytes.Buffer
	form := multipart.NewWriter(&buf)
	for _, file := range files {
		part, err := form.CreateFormFile("file", file.FileName)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(part, file.Content); err != nil {
			return nil, fmt.Errorf("rusender: read %s: %w", file.FileName, err)
		}
	}
	if err := form.Close(); err != nil {
		return nil, err
	}

	result, err := call[struct {
		Files []FileInfo `json:"files"`
	}](ctx, s.t, request{
		method:        "POST",
		path:          "/v1/public/files",
		form:          &multipartBody{contentType: form.FormDataContentType(), data: buf.Bytes()},
		notIdempotent: true,
	}, opts)
	if err != nil {
		return nil, err
	}
	return result.Files, nil
}

// Get gets file metadata: name, size, upload date. Use Download for the contents.
// Scope: files.read.
func (s *FilesService) Get(ctx context.Context, fileID int, opts ...RequestOption) (*FileInfo, error) {
	return call[FileInfo](ctx, s.t, request{method: "GET", path: pathf("/v1/public/files/%s", fileID)}, opts)
}

// Download downloads the file contents. Returns the raw response: read or stream its Body and close it.
// Scope: files.read.
func (s *FilesService) Download(ctx context.Context, fileID int, opts ...RequestOption) (*http.Response, error) {
	return s.t.raw(ctx, request{method: "GET", path: pathf("/v1/public/files/%s/download", fileID)}, opts)
}
