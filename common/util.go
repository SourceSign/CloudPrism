package common

import (
	"regexp"
	"strings"
)

func SanitizeForID(unclean string) string {
	compiledRegExp := regexp.MustCompile("[^a-zA-Z0-9-.]+")

	lowercase := strings.ToLower(strings.ReplaceAll(unclean, " ", "-"))
	sanitized := compiledRegExp.ReplaceAllString(lowercase, "")

	return sanitized
}
