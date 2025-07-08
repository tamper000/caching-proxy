package utils

import "regexp"

func GenerateRegexp(list []string) []*regexp.Regexp {
	var result []*regexp.Regexp
	for _, r := range list {
		re, err := regexp.Compile(r)
		if err != nil {
			panic(err)
		}

		result = append(result, re)
	}

	return result
}
