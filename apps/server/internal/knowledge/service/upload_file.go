package service

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	v1 "github.com/ZTH7/RagoDesk/apps/server/api/knowledge/v1"
	biz "github.com/ZTH7/RagoDesk/apps/server/internal/knowledge/biz"
	"github.com/go-kratos/kratos/v2/errors"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const maxUploadBytes = 10 << 20

type uploadDocumentFileItem struct {
	Document *v1.Document        `json:"document"`
	Version  *v1.DocumentVersion `json:"version"`
}

type uploadDocumentFileResponse struct {
	Items []uploadDocumentFileItem `json:"items"`
}

func (s *KnowledgeService) UploadDocumentFile(ctx khttp.Context) error {
	reqCtx := ctx.Request().Context()
	if err := requireTenantContext(reqCtx); err != nil {
		return err
	}
	if err := s.iamUC.RequirePermission(reqCtx, biz.PermissionDocumentUpload); err != nil {
		return err
	}

	if err := ctx.Request().ParseMultipartForm(maxUploadBytes); err != nil {
		return errors.BadRequest("DOC_UPLOAD_INVALID", "invalid multipart form")
	}

	kbID := strings.TrimSpace(ctx.Request().FormValue("kb_id"))
	title := strings.TrimSpace(ctx.Request().FormValue("title"))
	sourceType := strings.TrimSpace(ctx.Request().FormValue("source_type"))

	files := ctx.Request().MultipartForm.File["files"]
	if len(files) == 0 {
		files = ctx.Request().MultipartForm.File["file"]
	}
	if len(files) == 0 {
		return errors.BadRequest("DOC_UPLOAD_EMPTY", "no files provided")
	}

	if len(files) > 1 {
		title = ""
	}

	resp := uploadDocumentFileResponse{Items: make([]uploadDocumentFileItem, 0, len(files))}
	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			return err
		}
		payload, err := io.ReadAll(io.LimitReader(file, maxUploadBytes))
		_ = file.Close()
		if err != nil {
			return err
		}
		if len(payload) == 0 {
			return errors.BadRequest("DOC_CONTENT_EMPTY", "document content empty")
		}

		inferredType := sourceType
		if inferredType == "" {
			inferredType = inferSourceTypeFromFilename(fh.Filename)
		}
		contentType := strings.TrimSpace(fh.Header.Get("Content-Type"))
		if contentType == "" {
			contentType = detectContentType(fh.Filename, payload)
		}

		doc, ver, err := s.uc.UploadDocumentFile(reqCtx, kbID, title, inferredType, fh.Filename, payload, contentType)
		if err != nil {
			return err
		}
		resp.Items = append(resp.Items, uploadDocumentFileItem{
			Document: toDocument(doc),
			Version:  toDocumentVersion(ver),
		})
	}

	return ctx.Result(http.StatusOK, resp)
}

func inferSourceTypeFromFilename(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".pdf":
		return "pdf"
	case ".doc":
		return "doc"
	case ".docx":
		return "docx"
	case ".md", ".markdown":
		return "markdown"
	case ".html", ".htm":
		return "html"
	case ".txt":
		return "text"
	default:
		return "text"
	}
}

func detectContentType(filename string, payload []byte) string {
	if ext := filepath.Ext(filename); ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			return mimeType
		}
	}
	return http.DetectContentType(payload)
}
