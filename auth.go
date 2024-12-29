package gogohandlers_yt_batteries

import (
	"context"
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

func GetAuthMiddleware[TServiceProvider ggh.ServiceProvider, TReqBody, TGetParams, TRespBody, TErrorData any](settings YtClientSettings) func(func(*ggh.GGRequest[TServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, TErrorData], error)) func(*ggh.GGRequest[TServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, TErrorData], error) {
	return func(hFunc func(*ggh.GGRequest[TServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, TErrorData], error)) func(*ggh.GGRequest[TServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, TErrorData], error) {
		return func(ggreq *ggh.GGRequest[TServiceProvider, TReqBody, TGetParams]) (*ggh.GGResponse[TRespBody, TErrorData], error) {
			var ytClientConfig yt.Config

			ytCookie, err := ggreq.Request.Cookie(yt.YTCypressCookie)
			if err != nil {
				if err != http.ErrNoCookie {
					ggreq.Logger.Error("Error getting cookie", slog.String("error", err.Error()))
					return &ggh.GGResponse[TRespBody, TErrorData]{}, AuthError{}
				}
			}
			if ytCookie != nil {
				csrfToken := ggreq.Request.Header.Get(yt.XCSRFToken)
				ytClientConfig = yt.Config{
					Proxy: settings.YtProxy,
					Credentials: &yt.CookieCredentials{
						Cookie:    &http.Cookie{Name: yt.YTCypressCookie, Value: ytCookie.Value},
						CSRFToken: csrfToken,
					},
				}
			} else {
				tokenHeader := ggreq.Request.Header.Get("Authorization")
				if tokenHeader == "" {
					return &ggh.GGResponse[TRespBody, TErrorData]{}, AuthError{}
				} else {
					token := strings.Split(tokenHeader, " ")[1]
					ytClientConfig = yt.Config{
						Proxy: settings.YtProxy,
						Credentials: &yt.TokenCredentials{
							Token: token,
						},
					}
				}
			}

			ytClient, err := ythttp.NewClient(&ytClientConfig)
			if err != nil {
				ggreq.Logger.Error("Error getting yt client", slog.String("error", err.Error()))
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
