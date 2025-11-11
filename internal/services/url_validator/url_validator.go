package urlvalidator

import (
	"net/url"
)

type URLValidator struct{}

func New() *URLValidator {
	return &URLValidator{}
}

func (uv URLValidator) Validate(str string) bool {
	u, err := url.Parse(str)

	return err == nil && u.Scheme != "" && u.Host != ""
}
