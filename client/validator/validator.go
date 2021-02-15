package validator

import "regexp"

const (
	disconnect = `^\s*disconnect\s*$`
	listFiles  = `^\s*list-files\s*$`
	register   = `^\s*register\s+[a-z]+(\s+(?:\"[^"]+\")\s*)+$`
	unregister = `^\s*unregister\s+[a-z]+(\s+(?:\"[^"]+\")\s*)+$`
	download   = `^\s*download\s+[a-z]+(\s+(?:\"[^"]+\")\s*){2}$`
)

// Validator is a struct that contains:
//    - regexes - a slice of regular expressions, used for validating user input
type Validator struct {
	regexes []*regexp.Regexp
}

// CreateValidator is a factory method that:
//    - creates and returns a pointer to Validator struct with predefined regexes
func CreateValidator() *Validator {
	regexes := make([]*regexp.Regexp, 0, 5)
	regexes = append(regexes, regexp.MustCompile(disconnect))
	regexes = append(regexes, regexp.MustCompile(listFiles))
	regexes = append(regexes, regexp.MustCompile(register))
	regexes = append(regexes, regexp.MustCompile(unregister))
	regexes = append(regexes, regexp.MustCompile(download))
	return &Validator{
		regexes: regexes,
	}
}

// Validate is a function that:
//    - accepts:
//        - command - user input(string)
//    - returns:
//        - true - if the command is valid
//        - false - otherwise
func (v *Validator) Validate(command string) bool {
	for _, regex := range v.regexes {
		if regex.MatchString(command) {
			return true
		}
	}

	return false
}
