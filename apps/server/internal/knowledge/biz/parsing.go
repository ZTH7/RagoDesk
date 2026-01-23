package biz

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"

	"github.com/go-kratos/kratos/v2/errors"
	"rsc.io/pdf"
)

const maxDocumentBytes = 5 << 20

// CleaningStrategy normalizes parsed content.
type CleaningStrategy interface {
	Normalize(sourceType string, doc ParsedDocument) ParsedDocument
}

var numberedHeadingRe = regexp.MustCompile(`^\s*\d+(?:\.\d+)*[)\.]?\s+(.+)$`)

// DefaultCleaningStrategy applies built-in normalization rules.
type DefaultCleaningStrategy struct{}

func (DefaultCleaningStrategy) Normalize(sourceType string, doc ParsedDocument) ParsedDocument {
	return normalizeParsedDocument(sourceType, doc)
}

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
		if strings.Contains(content, "<") {
			content = stripHTMLTags(content)
		}
	}
	return cleanContent(content)
}

func parseDocument(ctx context.Context, sourceType string, raw []byte, meta DocumentMeta) (ParsedDocument, error) {
	doc := ParsedDocument{Meta: meta}
	switch sourceType {
	case "url":
		text, err := fetchURLText(ctx, strings.TrimSpace(string(raw)))
		if err != nil {
			return ParsedDocument{}, err
		}
		blocks, title := splitTextBlocks(text)
		if doc.Meta.Title == "" {
			doc.Meta.Title = title
		}
		doc.Blocks = blocks
		return doc, nil
	case "doc":
		text, err := parseDocBytes(raw)
		if err != nil {
			return ParsedDocument{}, err
		}
		blocks, title := splitTextBlocks(text)
		if doc.Meta.Title == "" {
			doc.Meta.Title = title
		}
		doc.Blocks = blocks
		return doc, nil
	case "docx":
		text, err := parseDocxBytes(raw)
		if err != nil {
			return ParsedDocument{}, err
		}
		blocks, title := splitTextBlocks(text)
		if doc.Meta.Title == "" {
			doc.Meta.Title = title
		}
		doc.Blocks = blocks
		return doc, nil
	case "pdf":
		blocks, err := parsePDFBlocks(raw)
		if err != nil {
			return ParsedDocument{}, err
		}
		doc.Blocks = blocks
		return doc, nil
	case "markdown":
		blocks, title := parseMarkdownBlocks(string(raw))
		if doc.Meta.Title == "" {
			doc.Meta.Title = title
		}
		doc.Blocks = blocks
		return doc, nil
	case "html":
		text := stripHTMLTags(string(raw))
		blocks, title := splitTextBlocks(text)
		if doc.Meta.Title == "" {
			doc.Meta.Title = title
		}
		doc.Blocks = blocks
		return doc, nil
	default:
		// text fall back to raw text
		blocks, title := splitTextBlocks(string(raw))
		if doc.Meta.Title == "" {
			doc.Meta.Title = title
		}
		doc.Blocks = blocks
		return doc, nil
	}
}

func normalizeParsedDocument(sourceType string, doc ParsedDocument) ParsedDocument {
	if len(doc.Blocks) == 0 {
		return doc
	}
	cleaned := make([]DocumentBlock, 0, len(doc.Blocks))
	for _, block := range doc.Blocks {
		text := prepareContentNormalized(sourceType, block.Text)
		if strings.TrimSpace(text) == "" {
			continue
		}
		block.Text = text
		cleaned = append(cleaned, block)
	}
	doc.Blocks = cleaned
	return doc
}

func enrichParsedDocument(doc ParsedDocument) ParsedDocument {
	if doc.Meta.Title == "" {
		doc.Meta.Title = inferTitleFromBlocks(doc.Blocks)
	}
	for i, block := range doc.Blocks {
		if block.PageNo > 0 && strings.TrimSpace(block.Section) == "" {
			doc.Blocks[i].Section = fmt.Sprintf("Page %d", block.PageNo)
		}
	}
	return doc
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

func parseDocBytes(payload []byte) (string, error) {
	if len(payload) == 0 {
		return "", errors.BadRequest("DOC_CONTENT_EMPTY", "doc content missing")
	}
	if looksLikeZip(payload) {
		return parseDocxBytes(payload)
	}
	text := extractDocBinaryText(payload)
	if strings.TrimSpace(text) == "" {
		return "", errors.BadRequest("DOC_PARSE_FAILED", "doc content parse failed")
	}
	return text, nil
}

func parseDocxBytes(payload []byte) (string, error) {
	if len(payload) == 0 {
		return "", errors.BadRequest("DOCX_CONTENT_EMPTY", "docx content missing")
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
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" {
				var text string
				if err := decoder.DecodeElement(&text, &t); err != nil {
					return "", err
				}
				if text != "" {
					builder.WriteString(text)
					builder.WriteString(" ")
				}
			}
		case xml.EndElement:
			if t.Name.Local == "p" {
				builder.WriteString("\n")
			}
		}
	}
	return builder.String(), nil
}

func parsePDFBlocks(payload []byte) ([]DocumentBlock, error) {
	if len(payload) == 0 {
		return nil, errors.BadRequest("PDF_CONTENT_EMPTY", "pdf content missing")
	}
	tmpFile, err := os.CreateTemp("", "ragodesk-*.pdf")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	if _, err := tmpFile.Write(payload); err != nil {
		_ = tmpFile.Close()
		return nil, err
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}
	reader, err := pdf.Open(filepath.Clean(tmpFile.Name()))
	if err != nil {
		return nil, err
	}
	blocks := make([]DocumentBlock, 0, reader.NumPage())
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		content := page.Content()
		var builder strings.Builder
		for _, text := range content.Text {
			if text.S == "" {
				continue
			}
			builder.WriteString(text.S)
			builder.WriteString(" ")
		}
		blocks = append(blocks, DocumentBlock{
			Text:   builder.String(),
			PageNo: int32(i),
		})
	}
	return blocks, nil
}

func looksLikeZip(payload []byte) bool {
	return len(payload) >= 4 && payload[0] == 'P' && payload[1] == 'K' && payload[2] == 0x03 && payload[3] == 0x04
}

func extractDocBinaryText(payload []byte) string {
	utf16Text := extractUTF16Text(payload)
	asciiText := extractASCIIText(payload)
	if len(utf16Text) >= len(asciiText) {
		return utf16Text
	}
	return asciiText
}

func extractUTF16Text(payload []byte) string {
	if len(payload) < 4 {
		return ""
	}
	count := len(payload) / 2
	u16 := make([]uint16, 0, count)
	for i := 0; i+1 < len(payload); i += 2 {
		u16 = append(u16, binary.LittleEndian.Uint16(payload[i:]))
	}
	runes := utf16.Decode(u16)
	return filterPrintable(runes)
}

func extractASCIIText(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}
	var out []string
	buf := make([]byte, 0, 128)
	flush := func() {
		if len(buf) >= 3 {
			out = append(out, string(buf))
		}
		buf = buf[:0]
	}
	for _, b := range payload {
		if b >= 32 && b < 127 {
			buf = append(buf, b)
			continue
		}
		flush()
	}
	flush()
	return strings.Join(out, " ")
}

func filterPrintable(runes []rune) string {
	var builder strings.Builder
	lastSpace := false
	for _, r := range runes {
		if r == 0 || r == '\uFFFD' {
			continue
		}
		if unicode.IsSpace(r) {
			if !lastSpace {
				builder.WriteRune(' ')
				lastSpace = true
			}
			continue
		}
		if unicode.IsPrint(r) {
			builder.WriteRune(r)
			lastSpace = false
		}
	}
	return strings.TrimSpace(builder.String())
}

func parseMarkdownBlocks(raw string) ([]DocumentBlock, string) {
	lines := strings.Split(raw, "\n")
	blocks := make([]DocumentBlock, 0)
	section := ""
	title := ""
	buf := make([]string, 0, 16)
	flush := func() {
		if len(buf) == 0 {
			return
		}
		blocks = append(blocks, DocumentBlock{
			Text:    strings.Join(buf, "\n"),
			Section: section,
		})
		buf = buf[:0]
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			heading := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			if heading != "" {
				if title == "" {
					title = heading
				}
				flush()
				section = heading
				continue
			}
		}
		buf = append(buf, line)
	}
	flush()
	if len(blocks) == 0 {
		blocks = append(blocks, DocumentBlock{Text: raw})
	}
	return blocks, title
}

func splitTextBlocks(raw string) ([]DocumentBlock, string) {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	blocks := make([]DocumentBlock, 0)
	section := ""
	title := ""
	buf := make([]string, 0, 8)
	flush := func() {
		if len(buf) == 0 {
			return
		}
		blocks = append(blocks, DocumentBlock{
			Text:    strings.Join(buf, "\n"),
			Section: section,
		})
		buf = buf[:0]
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			flush()
			continue
		}
		if heading, ok := detectHeading(trimmed); ok {
			if title == "" {
				title = heading
			}
			flush()
			section = heading
			continue
		}
		buf = append(buf, line)
	}
	flush()
	if len(blocks) == 0 && strings.TrimSpace(raw) != "" {
		blocks = append(blocks, DocumentBlock{Text: raw})
	}
	return blocks, title
}

func detectHeading(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", false
	}
	if strings.HasPrefix(trimmed, "#") {
		heading := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		if heading != "" {
			return heading, true
		}
	}
	if len([]rune(trimmed)) > 120 {
		return "", false
	}
	if match := numberedHeadingRe.FindStringSubmatch(trimmed); match != nil {
		heading := strings.TrimSpace(match[1])
		if heading != "" {
			return heading, true
		}
	}
	if strings.HasSuffix(trimmed, ":") && len([]rune(trimmed)) <= 80 {
		heading := strings.TrimSpace(strings.TrimSuffix(trimmed, ":"))
		if heading != "" {
			return heading, true
		}
	}
	if isAllCapsHeading(trimmed) {
		return trimmed, true
	}
	return "", false
}

func isAllCapsHeading(line string) bool {
	if len([]rune(line)) > 80 {
		return false
	}
	letters := 0
	for _, r := range line {
		if unicode.IsLetter(r) {
			letters++
			if unicode.IsLower(r) {
				return false
			}
		}
	}
	return letters >= 3
}

func inferTitleFromBlocks(blocks []DocumentBlock) string {
	for _, block := range blocks {
		lines := strings.Split(block.Text, "\n")
		for _, line := range lines {
			candidate := strings.TrimSpace(line)
			if candidate == "" {
				continue
			}
			if !hasLetterOrDigit(candidate) {
				continue
			}
			return shortenTitle(candidate, 120)
		}
	}
	return ""
}

func hasLetterOrDigit(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func shortenTitle(s string, max int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= max {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(string(runes[:max]))
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
