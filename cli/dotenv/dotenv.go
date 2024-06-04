package dotenv

import (
	"os"
	"strings"
)

// Loads the given path as an `.env` file, and parses it into a map.
//  1. Ignores everything on a line after a `#` comment symbol
//  2. Splits each line into key-value pairs by the first `=` symbol.
//  3. Trims whitespace before and after both key and value
//  4. Trims matching quotation marks (single and double quotes).
func Load(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	env := parse(content)

	return env, nil
}

func parse(content []byte) map[string]string {
	env := make(map[string]string)
	envLines := strings.Split(string(content), "\n")
	for _, envLine := range envLines {
		// remove everything after #
		commentIdx := strings.Index(envLine, "#")
		if commentIdx != -1 {
			envLine = envLine[:commentIdx]
		}

		keyvalue := strings.SplitN(envLine, "=", 2) // nolint: mnd
		if len(keyvalue) != 2 {                     // nolint: mnd
			continue
		}

		name := strings.TrimSpace(keyvalue[0])
		value := strings.TrimSpace(keyvalue[1])

		env[name] = trimQuotes(value)
	}

	return env
}

func trimQuotes(value string) string {
	for len(value) >= 2 && (value[0] == '\'' || value[0] == '"') {
		if value[len(value)-1] != value[0] {
			break
		}

		value = value[1 : len(value)-1]
	}

	return value
}
