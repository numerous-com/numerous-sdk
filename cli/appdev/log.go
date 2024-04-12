package appdev

import (
	"encoding/json"
	"log/slog"
)

func SlogJSON(key string, value any) slog.Attr {
	j, err := json.Marshal(value)

	if err != nil {
		return slog.Group(key, slog.Any("error", err), slog.Any("value", value))
	} else {
		return slog.String(key, string(j))
	}
}
