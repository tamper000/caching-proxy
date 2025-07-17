package utils

import (
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/tamper000/caching-proxy/internal/errors"
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

func CheckBearer(header string) (string, error) {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.ErrBearerAuth
	}
	secret := parts[1]

	return secret, nil
}
