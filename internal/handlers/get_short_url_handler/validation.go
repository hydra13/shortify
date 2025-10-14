package getshorturlhandler

import (
	"net/url"
)

func validation(str string) bool {
	u, err := url.Parse(str)

	return err == nil && u.Scheme != "" && u.Host != ""
}
