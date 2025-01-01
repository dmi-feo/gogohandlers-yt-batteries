package gogohandlers_yt_batteries

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	ggh "github.com/dmi-feo/gogohandlers"
	"go.ytsaurus.tech/yt/go/yt"
	"go.ytsaurus.tech/yt/go/yt/ythttp"
)

type YtClientSettings struct {
	YtProxy string
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
		if err != http.ErrNoCookie {
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

type YtServiceProvider interface {
	GetYtClientFactory() YtClientFactory
}

func GetAuthMiddleware[TServiceProvider YtServiceProvider, TReqBody, TGetParams, TRespBody, TErrorData any]() func(func(*ggh.GGRequest[TServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, TErrorData], error)) func(*ggh.GGRequest[TServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, TErrorData], error) {
	return func(hFunc func(*ggh.GGRequest[TServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, TErrorData], error)) func(*ggh.GGRequest[TServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, TErrorData], error) {
		return func(ggreq *ggh.GGRequest[TServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, TErrorData], error) {
			sp := *ggreq.ServiceProvider
			ytClientFactory := sp.GetYtClientFactory()
			ytClient, err := ytClientFactory.FromRequest(ggreq.Request)
			if err != nil {
				return &ggh.GGResponse[TRespBody, TErrorData]{}, AuthError{}
			}

			whoAmI, err := ytClient.WhoAmI(ggreq.Request.Context(), nil)
			if err != nil {
				ggreq.Logger.Error("Error getting whoami", slog.String("error", err.Error()))
				return &ggh.GGResponse[TRespBody, TErrorData]{}, AuthError{}
			}
			ggreq.Request = ggreq.Request.WithContext(context.WithValue(ggreq.Request.Context(), "user", whoAmI.Login))

			ggresp, err := hFunc(ggreq)
			return ggresp, err
		}
	}
}
