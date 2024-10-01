package version

import "context"

type VersionCheckResult string

const (
	VersionCheckResultOK       VersionCheckResult = "OK"
	VersionCheckResultWarning  VersionCheckResult = "Warning"
	VersionCheckResultCritical VersionCheckResult = "Critical"
	VersionCheckResultUnknown  VersionCheckResult = "Unknown"
)

type CheckVersionOutput struct {
	Result  VersionCheckResult
	Message string
}

type versionCheckResponse struct {
	VersionCheck struct {
		Typename string `graphql:"__typename"`
		OK       struct {
			Version string
		} `graphql:"... on VersionCheckOK"`
		Warning struct {
			Message string
		} `graphql:"... on VersionCheckWarning"`
		Critical struct {
			Message string
		} `graphql:"... on VersionCheckCritical"`
		Unknown struct {
			Version string
		} `graphql:"... on VersionUnknown"`
	} `graphql:"checkVersion(version: $version)"`
}

func (s *Service) Check(ctx context.Context) (CheckVersionOutput, error) {
	var resp versionCheckResponse

	err := s.client.Query(ctx, &resp, map[string]interface{}{"version": Version})
	if err != nil {
		return CheckVersionOutput{}, err
	}

	result := resp.VersionCheck
	switch result.Typename {
	case "VersionCheckOK":
		return CheckVersionOutput{Result: VersionCheckResultOK, Message: "Version is actual"}, nil
	case "VersionCheckWarning":
		return CheckVersionOutput{Result: VersionCheckResultWarning, Message: result.Warning.Message}, nil
	case "VersionCheckCritical":
		return CheckVersionOutput{Result: VersionCheckResultCritical, Message: result.Critical.Message}, nil
	case "VersionUnknown":
		return CheckVersionOutput{Result: VersionCheckResultUnknown, Message: "Unknown version '" + result.Unknown.Version + "'"}, nil
	default:
		panic("unexpected response from server")
	}
}
