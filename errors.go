package gogohandlers_yt_batteries

import (
	"log/slog"

	"go.ytsaurus.tech/yt/go/yterrors"
)

type YtLikeErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func HandleGGHYTErrors(err error, l *slog.Logger) (statusCode int, errorData *YtLikeErrorData) {
	l.Warn("Handling error", slog.String("error", err.Error()))
	switch err.(type) {
	case AuthError:
		statusCode, errorData = 401, &YtLikeErrorData{Code: int(yterrors.CodeAuthenticationError), Message: err.Error()}
	}
	if statusCode != 0 {
		l.Warn("Handled error", slog.Int("status_code", statusCode))
	}
	return
}
