package fileutils

import (
	g "github.com/lassi-koykka/fin-dev-api/src/utils"
	"os"
	"strings"
)

func ParseFileLines(path string) []string {
	data, err := os.ReadFile(path)
	g.Check(err)
	content := string(data)
	keywords := strings.Split(content, "\n")
	result := []string{}
	for _, kw := range keywords {
		trimmed := strings.TrimSpace(kw)
		if len(trimmed) < 1 {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
