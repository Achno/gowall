package providers

import (
	"fmt"
	"strings"
)

func BuildPrompt(base, filename, format string) string {

	prompt := base
	if format == "txt" {
		prompt += "Format the output in plain text"
	} else {
		prompt += " Format the output in Markdown."
		prompt = AddPageContextToPrompt(filename, prompt)
	}

	return prompt
}

// AddPageContextToPrompt enhances the prompt with page-specific context for multi-page documents.
// It extracts page information from filenames with format "document.pdf-page-2-of-5" or "document.pdf-page-2"
// and adds appropriate context to help the OCR provider understand the document structure and place headings.
func AddPageContextToPrompt(filename, originalPrompt string) string {
	prompt := originalPrompt

	if !strings.Contains(filename, "-page-") {
		return prompt
	}

	// Extract page info from filename like "document.pdf-page-2-of-5"
	parts := strings.Split(filename, "-page-")
	if len(parts) != 2 {
		return prompt
	}

	pageInfo := parts[1] // "2-of-5" or just "2"

	var pageNum, totalPages string
	if strings.Contains(pageInfo, "-of-") {
		pageParts := strings.Split(pageInfo, "-of-")
		pageNum = pageParts[0]
		totalPages = pageParts[1]
	} else {
		pageNum = pageInfo
	}

	if pageNum == "1" {
		if totalPages != "" {
			prompt += fmt.Sprintf(" This is the FIRST PAGE of a %s-page document. Use top-level headings (# and ##) as appropriate for a document beginning.", totalPages)
		} else {
			prompt += " This is the FIRST PAGE of a multi-page document. Use top-level headings (# and ##) as appropriate for a document beginning."
		}
	} else {
		if totalPages != "" {
			prompt += fmt.Sprintf(" This is PAGE %s of %s total pages (NOT the first page). Assume this document has already started with main headings on previous pages. Use continuation-level headings (##) unless you see clear evidence this page starts a major new section.", pageNum, totalPages)
		} else {
			prompt += fmt.Sprintf(" This is PAGE %s of a multi-page document (NOT the first page). Assume this document has already started with main headings on previous pages. Use continuation-level headings (##) unless you see clear evidence this page starts a major new section.", pageNum)
		}
	}

	return prompt
}
