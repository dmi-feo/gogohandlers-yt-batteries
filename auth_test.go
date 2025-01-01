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

type TestAppServiceProvider struct {
	logger          *slog.Logger
	YtClientFactory YtClientFactory
}

func (tsp TestAppServiceProvider) GetYtClientFactory() YtClientFactory {
	return tsp.YtClientFactory
}

func NewTestAppServiceProvider(logger *slog.Logger, ytSettings YtClientSettings) *TestAppServiceProvider {
	return &TestAppServiceProvider{
		logger: logger,
		YtClientFactory: YtClientFactory{
			logger:   logger,
			settings: ytSettings,
		},
	}
}

func TestAuthMiddleware(t *testing.T) {
	ctx := context.Background()
	container, err := yttc.RunContainer(ctx, yttc.WithAuth())
	require.NoError(t, err)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ytProxy, err := container.GetProxy(ctx)
	require.NoError(t, err)
	ytSettings := YtClientSettings{YtProxy: ytProxy}
	sp := NewTestAppServiceProvider(logger, ytSettings)

	var handlerFunc = func(ggreq *ggh.GGRequest[TestAppServiceProvider, struct{}, struct{}]) (*ggh.GGResponse[struct{}, YtLikeErrorData], error) {
		return &ggh.GGResponse[struct{}, YtLikeErrorData]{}, nil
	}

	handler := ggh.Uitzicht[TestAppServiceProvider, struct{}, struct{}, struct{}, YtLikeErrorData]{
		ServiceProvider: sp,
		HandlerFunc:     handlerFunc,
		Middlewares: []func(hFunc func(*ggh.GGRequest[TestAppServiceProvider, struct{}, struct{}]) (*ggh.GGResponse[struct{}, YtLikeErrorData], error)) func(*ggh.GGRequest[TestAppServiceProvider, struct{}, struct{}]) (*ggh.GGResponse[struct{}, YtLikeErrorData], error){
			GetAuthMiddleware[TestAppServiceProvider, struct{}, struct{}, struct{}, YtLikeErrorData](),
			ggh.GetErrorHandlingMiddleware[TestAppServiceProvider, struct{}, struct{}, struct{}, YtLikeErrorData](HandleGGHYTErrors),
			ggh.GetDataProcessingMiddleware[TestAppServiceProvider, struct{}, struct{}, struct{}, YtLikeErrorData](nil),
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
