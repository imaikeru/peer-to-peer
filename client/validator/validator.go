package validation

import "regexp"

const (
	disconnect = `^\sdisconnect\s$`
	listFiles  = `^\slist-files\s$`
	register   = `^\s*register\s+[a-z]+(\s+(?:\"[^"]+\")\s*)+$`
	unregister = `^\s*unregister\s+[a-z]+(\s+(?:\"[^"]+\")\s*)+$`
	download   = `^\s*register\s+[a-z]+(\s+(?:\"[^"]+\")\s*){2}$`
)

type Validator struct {
	regexes []*regexp.Regexp
	// disconnectRegex *regexp.Regexp
	// listFilesRegex  *regexp.Regexp
	// registerRegex   *regexp.Regexp
	// unregisterRegex *regexp.Regexp
	// downloadRegex   *regexp.Regexp
}

func createValidator() *Validator {
	regexes := make([]*regexp.Regexp, 0, 5)
	regexes = append(regexes, regexp.MustCompile(disconnect))
	regexes = append(regexes, regexp.MustCompile(listFiles))
	regexes = append(regexes, regexp.MustCompile(register))
	regexes = append(regexes, regexp.MustCompile(unregister))
	regexes = append(regexes, regexp.MustCompile(download))
	return &Validator{
		regexes: regexes,
	}
	// return &Validator{
	// 	disconnectRegex: regexp.MustCompile(`^\sdisconnect\s$`),
	// 	listFilesRegex:  regexp.MustCompile(`^\slist-files\s$`),
	// 	registerRegex:   regexp.MustCompile(`^\s*register\s+[a-z]+(\s+(?:\"[^"]+\")\s*)+$`),
	// 	unregisterRegex: regexp.MustCompile(`^\s*unregister\s+[a-z]+(\s+(?:\"[^"]+\")\s*)+$`),
	// 	downloadRegex:   regexp.MustCompile(`^\s*register\s+[a-z]+(\s+(?:\"[^"]+\")\s*){2}$`),
	// }
}

func (v *Validator) validate(command string) bool {
	// disconnectRegex := regexp.MustCompile(`^\sdisconnect\s$`)
	// listFilesRegex := regexp.MustCompile(`^\slist-files\s$`)
	// registerRegex := regexp.MustCompile(`^\s*register\s+[a-z]+(\s+(?:\"[^"]+\")\s*)+$`)
	// unregisterRegex := regexp.MustCompile(`^\s*unregister\s+[a-z]+(\s+(?:\"[^"]+\")\s*)+$`)
	// downloadRegex := regexp.MustCompile(`^\s*register\s+[a-z]+(\s+(?:\"[^"]+\")\s*){2}$`)

	for _, regex := range v.regexes {
		if regex.MatchString(command) {
			return true
		}
	}

	return false
}
