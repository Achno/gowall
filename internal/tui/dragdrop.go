package tui

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var supportedDropExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".gif":  true,
	".bmp":  true,
	".tif":  true,
	".tiff": true,
	".ico":  true,
	".svg":  true,
}

func isSupportedImageExt(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return supportedDropExts[ext]
}

// ParseDroppedPaths extracts valid image file paths (or directories) from a
// terminal drag-and-drop payload. It supports:
//   - Bracketed paste payloads from Alacritty, Kitty, WezTerm, Ghostty, Foot, GNOME Terminal, Konsole, iTerm2
//   - Single-quoted paths: '/path/to/my image.png' (including '\'' escaped quotes)
//   - Double-quoted paths: "/path/to/my image.png"
//   - Backslash-escaped paths: /path/to/my\ image\ \(1\).png
//   - RFC 2483 text/uri-list and file:// URIs: file:///path/to/my%20image.png
//   - Home-relative paths: ~/Pictures/wallpaper.png
func ParseDroppedPaths(raw string) []string {
	raw = strings.Trim(raw, " \t\r\n\x00")
	if raw == "" {
		return nil
	}

	var results []string
	seen := make(map[string]bool)
	addIfValid := func(candidate string) bool {
		cleaned := cleanDroppedToken(candidate)
		if cleaned == "" {
			return false
		}
		for _, tryPath := range []string{cleaned, "/" + strings.TrimPrefix(cleaned, "/")} {
			info, err := os.Stat(tryPath)
			if err != nil {
				continue
			}
			abs, err := filepath.Abs(tryPath)
			if err != nil {
				abs = tryPath
			}
			if info.IsDir() || isSupportedImageExt(abs) {
				if !seen[abs] {
					seen[abs] = true
					results = append(results, abs)
				}
				return true
			}
		}
		return false
	}

	// 1. Try the entire payload as a single path first
	if addIfValid(raw) {
		return results
	}

	// 2. Process line by line (handles text/uri-list and multi-line drops)
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if addIfValid(line) {
			continue
		}
		// 3. Tokenize shell-quoted or space-separated multiple file drops on a single line
		for _, tok := range splitShellDroppedTokens(line) {
			addIfValid(tok)
		}
	}

	return results
}

func cleanDroppedToken(tok string) string {
	s := strings.TrimSpace(tok)
	if s == "" {
		return ""
	}

	// Unquote single-quoted shell string ('/path/to/file.png' or '/path/john'\''s.png')
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		s = s[1 : len(s)-1]
		s = strings.ReplaceAll(s, `'\''`, `'`)
	} else if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	} else {
		// Strip unbalanced leading/trailing quotes from partial captures
		s = strings.Trim(s, `'"`)
	}

	s = strings.TrimSpace(s)

	// Handle file:// URI
	if strings.HasPrefix(strings.ToLower(s), "file://") {
		if u, err := url.Parse(s); err == nil && u.Path != "" {
			if unescaped, err := url.PathUnescape(u.Path); err == nil {
				s = unescaped
			} else {
				s = u.Path
			}
		} else {
			rest := s[len("file://"):]
			rest = strings.TrimPrefix(rest, "localhost")
			if unescaped, err := url.PathUnescape(rest); err == nil {
				s = unescaped
			} else {
				s = rest
			}
		}
	} else if strings.Contains(s, `\`) {
		// Unescape shell backslash escapes (only on Unix paths where '/' is present)
		if strings.Contains(s, "/") {
			var b strings.Builder
			b.Grow(len(s))
			for i := 0; i < len(s); i++ {
				if s[i] == '\\' && i+1 < len(s) {
					i++
					b.WriteByte(s[i])
				} else {
					b.WriteByte(s[i])
				}
			}
			s = b.String()
		}
	}

	// Expand ~/ home prefix
	if strings.HasPrefix(s, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			s = filepath.Join(home, s[2:])
		}
	}

	return strings.TrimSpace(s)
}

func splitShellDroppedTokens(line string) []string {
	var tokens []string
	var cur strings.Builder
	inSingle := false
	inDouble := false

	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case ch == '\'' && !inDouble:
			if inSingle && i+3 < len(line) && line[i:i+4] == `'\''` {
				cur.WriteByte('\'')
				i += 3
				continue
			}
			inSingle = !inSingle
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case ch == '\\' && !inSingle && i+1 < len(line):
			i++
			cur.WriteByte(line[i])
		case (ch == ' ' || ch == '\t') && !inSingle && !inDouble:
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(ch)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}
