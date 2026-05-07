package envreader

import (
	"bufio"
	"io"
	"os"
	"strings"
	"unicode"
)

// LoadKeys reads a .env file and returns all key names (no values)
func LoadKeys(path string) ([]string, error) {
	m, err := LoadMap(path)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys, nil
}

// LoadMap reads a .env file and returns a map of key -> value
func LoadMap(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseEnvFile(f)
}

func parseEnvFile(r io.Reader) (map[string]string, error) {
	scanner := bufio.NewScanner(r)
	result := map[string]string{}
	var pending string

	for scanner.Scan() {
		line := scanner.Text()
		if pending != "" {
			line = pending + line
			pending = ""
		}

		if endsWithUnescapedBackslash(line) {
			pending = strings.TrimSuffix(line, "\\")
			continue
		}

		if key, val, ok := parseEnvLine(line); ok {
			result[key] = val
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func endsWithUnescapedBackslash(line string) bool {
	count := 0
	for i := len(line) - 1; i >= 0; i-- {
		if line[i] == '\\' {
			count++
			continue
		}
		break
	}
	return count%2 == 1
}

func parseEnvLine(line string) (string, string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", false
	}

	if strings.HasPrefix(trimmed, "export ") {
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "export "))
	}

	idx := strings.Index(trimmed, "=")
	if idx < 0 {
		return "", "", false
	}

	key := strings.TrimSpace(trimmed[:idx])
	if key == "" {
		return "", "", false
	}

	value := strings.TrimSpace(trimmed[idx+1:])
	if value == "" {
		return key, "", true
	}

	if value[0] == '\'' || value[0] == '"' {
		parsed, ok := parseQuotedValue(value)
		if !ok {
			return "", "", false
		}
		return key, parsed, true
	}

	value = stripInlineComment(value)
	return key, strings.TrimSpace(value), true
}

func parseQuotedValue(value string) (string, bool) {
	if len(value) < 2 {
		return "", false
	}

	quote := value[0]
	var out strings.Builder
	escaped := false

	for i := 1; i < len(value); i++ {
		ch := value[i]
		if escaped {
			escaped = false
			if quote == '"' {
				switch ch {
				case 'n':
					out.WriteByte('\n')
				case 'r':
					out.WriteByte('\r')
				case 't':
					out.WriteByte('\t')
				default:
					out.WriteByte(ch)
				}
			} else {
				out.WriteByte(ch)
			}
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		if ch == quote {
			return out.String(), true
		}

		out.WriteByte(ch)
	}

	return "", false
}

func stripInlineComment(value string) string {
	var out strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for i, ch := range value {
		if escaped {
			out.WriteRune(ch)
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			out.WriteRune(ch)
			continue
		}

		if ch == '"' && !inSingle {
			inDouble = !inDouble
			out.WriteRune(ch)
			continue
		}

		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			out.WriteRune(ch)
			continue
		}

		if !inSingle && !inDouble {
			if ch == '#' {
				if i == 0 || unicode.IsSpace(rune(value[i-1])) {
					break
				}
			}
		}

		out.WriteRune(ch)
	}

	return out.String()
}
