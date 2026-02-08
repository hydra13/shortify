// Модуль валидации URL
package urlvalidator

import (
	"net/url"
	"regexp"
	"strings"
)

// URLValidator представляет собой сервис для валидации URL.
//
// Пример использования:
//
//	validator := urlvalidator.New()
//	isValid := validator.Validate("https://example.com")
type URLValidator struct {
	re *regexp.Regexp
}

func New() *URLValidator {
	v := &URLValidator{}

	re, err := regexp.Compile(`http[s]?://[\w\.]+`)
	if err == nil {
		v.re = re
	}

	return v
}

func (uv URLValidator) Validate(str string) bool {
	return uv.validateByRegexp(str)
}

// validateByURL валидация с помощью встроенной библиотеки url.
func (uv URLValidator) validateByURL(str string) bool {
	u, err := url.Parse(str)

	return err == nil && u.Scheme != "" && u.Host != ""
}

// validateByRegexp валидация с помощью regexp.
func (uv URLValidator) validateByRegexp(str string) bool {
	if uv.re == nil {
		return false
	}

	return uv.re.MatchString(str)
}

// validateByStringsAndURL оптимизированная версия валидации с комбинированием strings и url.
func (uv URLValidator) validateByStringsAndURL(str string) bool {
	if !strings.Contains(str, "://") {
		return false
	}

	u, err := url.Parse(str)

	return err == nil && u.Scheme != "" && u.Host != ""
}
