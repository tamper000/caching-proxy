package utils

import (
	"log/slog"
	"os"
	"regexp"
)

func GenerateRegexp(list []string) []*regexp.Regexp {
	var result []*regexp.Regexp
	for _, r := range list {
		re, err := regexp.Compile(r)
		if err != nil {
			slog.Error("Blacklist failed", "error", err)
			os.Exit(1)
		}

		result = append(result, re)
	}

	return result
}
