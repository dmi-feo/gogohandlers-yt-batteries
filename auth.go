package gogohandlers_yt_batteries

import (
	"context"
	"log/slog"

	ggh "github.com/dmi-feo/gogohandlers"
)

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
