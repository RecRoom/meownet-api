package utils

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const blocklistDir = "db/words"
const sepEvasionMinLen = 5

var blocklistFiles = []string{"en", "es", "fr"}

var (
	blocklistOnce sync.Once
	blocklistRe   *regexp.Regexp
)

var leetClasses = map[rune]string{
	'a': "4@",
	'b': "8",
	'e': "3",
	'g': "9",
	'i': "1!",
	'l': "1",
	'o': "0",
	's': "5$",
	't': "7+",
	'z': "2",
}

func loadBlocklist() {
	var patterns []string
	seen := map[string]struct{}{}
	for _, name := range blocklistFiles {
		path := blocklistDir + "/" + name
		f, err := os.Open(path)
		if err != nil {
			log.Printf("[MODERATION] blocklist %s: %v", path, err)
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			entry := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if entry == "" {
				continue
			}
			if _, dup := seen[entry]; dup {
				continue
			}
			seen[entry] = struct{}{}
			if p := entryPattern(entry); p != "" {
				patterns = append(patterns, p)
			}
		}
		f.Close()
	}
	if len(patterns) == 0 {
		log.Printf("[MODERATION] blocklist empty; deterministic text filter disabled")
		return
	}
	full := `(?:^|[^a-z])(?:` + strings.Join(patterns, "|") + `)(?:$|[^a-z])`
	re, err := regexp.Compile(full)
	if err != nil {
		log.Printf("[MODERATION] blocklist compile error: %v", err)
		return
	}
	blocklistRe = re
	log.Printf("[MODERATION] blocklist loaded: %d entries", len(patterns))
}

func entryPattern(entry string) string {
	var letters []rune
	for _, r := range entry {
		if r >= 'a' && r <= 'z' {
			letters = append(letters, r)
		}
	}
	if len(letters) < 2 {
		return ""
	}
	sep := ""
	if len(letters) >= sepEvasionMinLen {
		sep = `[^a-z0-9]*`
	}
	var sb strings.Builder
	for i, r := range letters {
		if i > 0 {
			sb.WriteString(sep)
		}
		sb.WriteByte('[')
		sb.WriteByte(byte(r))
		sb.WriteString(escapeClass(leetClasses[r]))
		sb.WriteByte(']')
	}
	return sb.String()
}

func escapeClass(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '\\', ']', '^', '-':
			sb.WriteByte('\\')
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

func foldToASCII(s string) string {
	var sb strings.Builder
	for _, r := range norm.NFKD.String(strings.ToLower(s)) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

func IsBlocklisted(text string) bool {
	blocklistOnce.Do(loadBlocklist)
	if blocklistRe == nil {
		return false
	}
	return blocklistRe.MatchString(foldToASCII(text))
}
