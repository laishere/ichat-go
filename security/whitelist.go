package security

import (
	"ichat-go/config"
	"ichat-go/errs"
	"strings"
)

func IsWhiteListed(url string) bool {
	i := strings.Index(url, config.App.ApiPrefix)
	if i != 0 {
		panic(errs.Forbidden)
	}
	url = url[len(config.App.ApiPrefix):]
	if strings.HasPrefix(url, "/ws") {
		return true
	}
	if strings.HasPrefix(url, "/file/") {
		return true
	}
	whitelist := []string{"/login", "/logout", "/register"}
	for _, path := range whitelist {
		if path == url {
			return true
		}
	}
	return false
}
