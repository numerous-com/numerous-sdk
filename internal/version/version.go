package version

import _ "embed"

// Embeds a version number through version.txt which is created from
// pyproject.toml by Make
//
//go:embed version.txt
var Version string
