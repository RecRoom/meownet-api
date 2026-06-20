package utils

import "strings"

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func EscapeLike(s string) string {
	return likeEscaper.Replace(s)
}
