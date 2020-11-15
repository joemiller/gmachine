package indentor

import "strings"

func Indent(indentation, s string) string {
	indentedLines := []string{}
	for _, line := range strings.Split(s, "\n") {
		// line = strings.TrimSpace(line)
		line = indentation + line
		indentedLines = append(indentedLines, line)
	}
	return strings.Join(indentedLines, "\n")
}
