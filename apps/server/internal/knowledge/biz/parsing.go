package biz

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
)

const maxDocumentBytes = 5 << 20

func normalizeSourceType(sourceType string) string {
	value := strings.ToLower(strings.TrimSpace(sourceType))
	switch value {
	case "", "text", "plain", "txt":
		return "text"
	case "md", "markdown":
		return "markdown"
	case "html", "htm":
		return "html"
	case "doc", "docx":
		return "doc"
	case "pdf":
		return "pdf"
	case "url", "link":
		return "url"
	default:
		return value
	}
}

func prepareContentNormalized(sourceType, content string) string {
	switch sourceType {
	case "markdown":
		content = stripMarkdown(content)
	case "html":
		content = stripHTMLTags(content)
	}
	return cleanContent(content)
}

func parseContent(ctx context.Context, sourceType, raw string) (string, error) {
	switch sourceType {
	case "url":
		return fetchURLText(ctx, raw)
	case "doc":
		return parseDocxBase64(raw)
	default:
		// pdf / text / markdown / html fall back to raw text
		return raw, nil
	}
}

func fetchURLText(ctx context.Context, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.BadRequest("DOC_URL_EMPTY", "document url missing")
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.BadRequest("DOC_URL_INVALID", "document url invalid")
	}
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.BadRequest("DOC_URL_FETCH_FAILED", "document url fetch failed")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxDocumentBytes))
	if err != nil {
		return "", err
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	text := string(body)
	if strings.Contains(contentType, "text/html") {
		text = stripHTMLTags(text)
	}
	return text, nil
}

func parseDocxBase64(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.BadRequest("DOCX_CONTENT_EMPTY", "docx content missing")
	}
	payload, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", errors.BadRequest("DOCX_BASE64_INVALID", "docx content must be base64")
	}
	readerAt := bytes.NewReader(payload)
	zr, err := zip.NewReader(readerAt, int64(len(payload)))
	if err != nil {
		return "", err
	}
	var xmlFile *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			xmlFile = f
			break
		}
	}
	if xmlFile == nil {
		return "", errors.BadRequest("DOCX_XML_MISSING", "docx document xml missing")
	}
	rc, err := xmlFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	xmlBytes, err := io.ReadAll(io.LimitReader(rc, maxDocumentBytes))
	if err != nil {
		return "", err
	}
	decoder := xml.NewDecoder(bytes.NewReader(xmlBytes))
	var builder strings.Builder
	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "t" {
			continue
		}
		var text string
		if err := decoder.DecodeElement(&text, &start); err != nil {
			return "", err
		}
		if text != "" {
			builder.WriteString(text)
			builder.WriteString(" ")
		}
	}
	return builder.String(), nil
}

func cleanContent(input string) string {
	if input == "" {
		return ""
	}
	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		fields := strings.Fields(line)
		out = append(out, strings.Join(fields, " "))
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func stripMarkdown(input string) string {
	if input == "" {
		return ""
	}
	lines := strings.Split(input, "\n")
	out := make([]string, 0, len(lines))
	inCode := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			out = append(out, line)
			continue
		}
		line = strings.TrimLeft(line, " \t")
		if strings.HasPrefix(line, ">") {
			line = strings.TrimSpace(strings.TrimPrefix(line, ">"))
		}
		if strings.HasPrefix(line, "#") {
			line = strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "+ ") {
			line = strings.TrimSpace(line[2:])
		}
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "__", "")
		line = strings.ReplaceAll(line, "`", "")
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func stripHTMLTags(input string) string {
	if input == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(input))
	inTag := false
	for _, r := range input {
		switch r {
		case '<':
			inTag = true
			b.WriteRune(' ')
		case '>':
			inTag = false
			b.WriteRune(' ')
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}
