package version

import (
	"context"

	"numerous.com/cli/internal/output"
	"numerous.com/cli/internal/version"
)

type VersionChecker interface {
	Check(ctx context.Context) (version.CheckVersionOutput, error)
}

func Check(checker VersionChecker) bool {
	out, err := checker.Check(context.Background())

	switch {
	case err != nil:
		output.PrintWarning("Failed to check CLI version", "An error occurred")
	case out.Result == version.VersionCheckResultOK:
	case out.Result == version.VersionCheckResultWarning:
		output.PrintWarning("CLI is outdated", out.Message)
	case out.Result == version.VersionCheckResultCritical:
		output.PrintError("CLI is version is not supported. Please update it", out.Message)
		return false
	case out.Result == version.VersionCheckResultUnknown:
		output.PrintWarning("Failed to check CLI version", out.Message)
	}

	return true
}
