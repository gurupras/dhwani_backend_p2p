package utils

import "regexp"

func GetRegexGroups(regex *regexp.Regexp, str string) (result map[string]string) {
	match := regex.FindStringSubmatch(str)
	result = make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i > 0 && i <= len(match) {
			result[name] = match[i]
		}
	}
	return result
}
