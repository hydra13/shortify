// Package urlvalidator - валидатор URL.
package urlvalidator

import (
	"net/url"
	"regexp"
	"strings"
)

type URLValidator struct{}

func New() *URLValidator {
	return &URLValidator{}
}

func (uv URLValidator) Validate(str string) bool {
	return uv.validateByStringsAndURL(str)
}

// validateByURL валидация с помощью встроенной библиотеки url.
func (uv URLValidator) validateByURL(str string) bool {
	u, err := url.Parse(str)

	return err == nil && u.Scheme != "" && u.Host != ""
}

// validateByRegexp валидация с помощью regexp.
func (uv URLValidator) validateByRegexp(str string) bool {
	matched, err := regexp.MatchString(`http[s]?://[\w\.]+`, str)
	if err != nil {
		return false
	}

	return matched
}

// validateByStringsAndURL оптимизированная версия валидации с комбинированием strings и url.
func (uv URLValidator) validateByStringsAndURL(str string) bool {
	if !strings.Contains(str, "://") {
		return false
	}

	u, err := url.Parse(str)

	return err == nil && u.Scheme != "" && u.Host != ""
}
