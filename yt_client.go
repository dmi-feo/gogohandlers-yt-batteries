package gogohandlers_yt_batteries

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"go.ytsaurus.tech/yt/go/yt"
	"go.ytsaurus.tech/yt/go/yt/ythttp"
)

type YtClientSettings struct {
	YtProxy string
	Token   string
}

type AuthError struct{}

func (e AuthError) Error() string {
	return "Authentication error"
}

type YtClientFactory struct {
	logger   *slog.Logger
	settings YtClientSettings
}

func (ytcf *YtClientFactory) FromRequest(req *http.Request) (yt.Client, error) {
	var ytClientConfig yt.Config

	ytCookie, err := req.Cookie(yt.YTCypressCookie)
	if err != nil {
		if !errors.Is(err, http.ErrNoCookie) {
			ytcf.logger.Error("Error getting cookie", slog.String("error", err.Error()))
			return nil, err
		}
	}
	if ytCookie != nil {
		csrfToken := req.Header.Get(yt.XCSRFToken)
		ytClientConfig = yt.Config{
			Proxy: ytcf.settings.YtProxy,
			Credentials: &yt.CookieCredentials{
				Cookie:    &http.Cookie{Name: yt.YTCypressCookie, Value: ytCookie.Value},
				CSRFToken: csrfToken,
			},
		}
	} else {
		tokenHeader := req.Header.Get("Authorization")
		if tokenHeader == "" {
			ytcf.logger.Warn("Neither cookie not header were provided for auth")
			return nil, fmt.Errorf("no auth data provided")
		} else {
			token := strings.Split(tokenHeader, " ")[1]
			ytClientConfig = yt.Config{
				Proxy: ytcf.settings.YtProxy,
				Credentials: &yt.TokenCredentials{
					Token: token,
				},
			}
		}
	}

	ytClient, err := ythttp.NewClient(&ytClientConfig)
	if err != nil {
		ytcf.logger.Error("Error getting yt client", slog.String("error", err.Error()))
		return nil, err
	}

	return ytClient, nil
}

func (ytcf *YtClientFactory) FromSettings() (yt.Client, error) {
	if ytcf.settings.Token == "" {
		return nil, fmt.Errorf("yt token not set")
	}
	token := ytcf.settings.Token

	ytClientConfig := yt.Config{
		Proxy: ytcf.settings.YtProxy,
		Credentials: &yt.TokenCredentials{
			Token: token,
		},
	}

	ytClient, err := ythttp.NewClient(&ytClientConfig)
	if err != nil {
		ytcf.logger.Error("Error getting yt client", slog.String("error", err.Error()))
		return nil, err
	}

	return ytClient, nil
}

type YtServiceProvider interface {
	GetYtClientFactory() YtClientFactory
}
