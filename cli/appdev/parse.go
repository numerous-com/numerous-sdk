package appdev

import (
	"encoding/json"
	"log/slog"
)

type AppCodeCoordinate struct {
	Line   uint `json:"line"`
	Offset uint `json:"offset"`
}

type AppSyntaxError struct {
	Msg     string            `json:"msg"`
	Context string            `json:"context"`
	Pos     AppCodeCoordinate `json:"pos"`
}

type AppModuleNotFoundError struct {
	Module string `json:"module"`
}

type AppUnknownError struct {
	Typename  string `json:"typename"`
	Traceback string `json:"traceback"`
}

type AppNotFoundError struct {
	App       string   `json:"app"`
	FoundApps []string `json:"found_apps"`
}

type ParseAppDefinitionError struct {
	AppNotFound    *AppNotFoundError       `json:"appnotfound"`
	Syntax         *AppSyntaxError         `json:"appsyntax"`
	ModuleNotFound *AppModuleNotFoundError `json:"modulenotfound"`
	Unknown        *AppUnknownError        `json:"unknown"`
}

type ParseAppDefinitionResult struct {
	App   *AppDefinition           `json:"app"`
	Error *ParseAppDefinitionError `json:"error"`
}

func ParseAppDefinition(definition []byte) (ParseAppDefinitionResult, error) {
	var result ParseAppDefinitionResult

	if err := json.Unmarshal(definition, &result); err != nil {
		slog.Warn(
			"Failed to unmarshal app definition",
			slog.Any("definition", definition),
		)

		return result, err
	} else if result.App != nil {
		result.App.SetElementParents()
	}

	return result, nil
}
