package appident

import "regexp"

func IsValidIdentifier(i string) bool {
	m, err := regexp.Match(`^[a-z0-9-]+$`, []byte(i))
	return m && err == nil
}
