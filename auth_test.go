package gogohandlers_yt_batteries

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	ggh "github.com/dmi-feo/gogohandlers"
	"github.com/stretchr/testify/require"
	yttc "github.com/tractoai/testcontainers-ytsaurus"
	"go.ytsaurus.tech/yt/go/yterrors"
)

func TestAuthMiddleware(t *testing.T) {
	ctx := context.Background()
	container, err := yttc.RunContainer(ctx, yttc.WithAuth())
	require.NoError(t, err)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var handlerFunc = func(ggreq *ggh.GGRequest[struct{}, struct{}, struct{}]) (*ggh.GGResponse[struct{}, YtLikeErrorData], error) {
		return &ggh.GGResponse[struct{}, YtLikeErrorData]{}, nil
	}

	ytProxy, err := container.GetProxy(ctx)
	require.NoError(t, err)
	ytSettings := YtClientSettings{YtProxy: ytProxy}

	handler := ggh.Uitzicht[struct{}, struct{}, struct{}, struct{}, YtLikeErrorData]{
		ServiceProvider: new(struct{}),
		HandlerFunc:     handlerFunc,
		Middlewares: []func(hFunc func(*ggh.GGRequest[struct{}, struct{}, struct{}]) (*ggh.GGResponse[struct{}, YtLikeErrorData], error)) func(*ggh.GGRequest[struct{}, struct{}, struct{}]) (*ggh.GGResponse[struct{}, YtLikeErrorData], error){
			GetAuthMiddleware[struct{}, struct{}, struct{}, struct{}, YtLikeErrorData](ytSettings),
			ggh.GetErrorHandlingMiddleware[struct{}, struct{}, struct{}, struct{}, YtLikeErrorData](HandleGGHYTErrors),
			ggh.GetDataProcessingMiddleware[struct{}, struct{}, struct{}, struct{}, YtLikeErrorData](nil),
		},
		Logger: logger,
	}

	t.Run("request without auth", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusUnauthorized, response.Code)

		var respBody YtLikeErrorData
		err = json.Unmarshal(response.Body.Bytes(), &respBody)
		require.NoError(t, err)
		require.Equal(t, int(yterrors.CodeAuthenticationError), respBody.Code)
	})

	t.Run("request with bad token", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		request.Header.Add("Authorization", "OAuth wrong-token")
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusUnauthorized, response.Code)
		var respBody YtLikeErrorData
		err = json.Unmarshal(response.Body.Bytes(), &respBody)
		require.NoError(t, err)
		require.Equal(t, int(yterrors.CodeAuthenticationError), respBody.Code)
	})

	t.Run("request with correct token", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		request.Header.Add("Authorization", fmt.Sprintf("OAuth %s", yttc.DefaultToken))
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		require.Equal(t, http.StatusOK, response.Code)
	})
}