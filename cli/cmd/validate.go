package cmd

import "regexp"

func validIdent(i string) bool {
	m, err := regexp.Match(`^[a-z0-9-]+$`, []byte(i))
	return m && err == nil
}
