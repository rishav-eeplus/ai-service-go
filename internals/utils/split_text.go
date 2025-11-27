package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// Section represents a section of text with its title and summary
type Section struct {
	Title    string
	Summary  string
	Content  string // Full content including title and summary
	Metadata map[string]string
}

// SplitText splits text into chunks with overlap
func SplitText(text string, chunkSize, overlap int) ([]string, error) {
	if chunkSize <= 0 || overlap >= chunkSize {
		return nil, errors.New("invalid chunk size or overlap")
	}

	var chunks []string

	if len(text) <= chunkSize {
		return []string{text}, nil
	}

	start := 0
	for start < len(text) {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}

		// Try to break at sentence boundary
		if end < len(text) {
			searchStart := end - min(100, end-start)
			snippet := text[searchStart:end]

			// Look for sentence ending punctuation
			re := regexp.MustCompile(`[.!?]\s`)
			matches := re.FindAllStringIndex(snippet, -1)

			if len(matches) > 0 {
				lastMatch := matches[len(matches)-1]
				end = searchStart + lastMatch[1]
			} else {
				// Try to break at word boundary
				for end > start && !unicode.IsSpace(rune(text[end-1])) {
					end--
				}
				if end == start {
					end = start + chunkSize
				}
			}
		}

		chunk := strings.TrimSpace(text[start:end])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		// Move start position
		if end-overlap > start {
			start = end - overlap
		} else {
			start = end
		}
	}

	// Merge small last chunk if necessary
	if len(chunks) > 1 {
		lastChunk := chunks[len(chunks)-1]
		if len(lastChunk) < overlap {
			chunks[len(chunks)-2] += " " + lastChunk
			chunks = chunks[:len(chunks)-1]
		}
	}

	return chunks, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SplitTextBySections splits text into sections based on ## titles and $$ summaries
// Returns sections with metadata containing the summary
func SplitTextBySections(text string) ([]Section, error) {
	if text == "" {
		return nil, errors.New("text cannot be empty")
	}

	var sections []Section
	lines := strings.Split(text, "\n")

	var currentSection *Section
	var contentLines []string
	inSection := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		// Check if this line starts a new section (begins with ##)
		if strings.HasPrefix(trimmedLine, "##") {
			// Save previous section if exists
			if currentSection != nil {
				currentSection.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
				sections = append(sections, *currentSection)
			}
			// Start new section
			title := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "##"))
			currentSection = &Section{
				Title:    title,
				Metadata: make(map[string]string),
			}
			contentLines = []string{line} // Include title in content
			inSection = true

			// Check if next line has summary ($$...$$)
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if strings.HasPrefix(nextLine, "$$") && strings.HasSuffix(nextLine, "$$") {
					// Extract summary
					summary := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(nextLine, "$$"), "$$"))
					currentSection.Summary = summary
					currentSection.Metadata["summary"] = summary

					// Add summary line to content as well
					i++ // Skip the summary line in next iteration
					contentLines = append(contentLines, lines[i])
				}
			}
		} else if inSection {
			// Add line to current section's content
			contentLines = append(contentLines, line)
		}
	}

	// Save the last section if exists
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
		sections = append(sections, *currentSection)
	}

	if len(sections) == 0 {
		return nil, errors.New("no sections found in text")
	}

	return sections, nil
}
