package gogohandlers_yt_batteries

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	yttc "github.com/tractoai/testcontainers-ytsaurus"
)

func TestYtClientFactory(t *testing.T) {
	ctx := context.Background()
	container, err := yttc.RunContainer(ctx, yttc.WithAuth())
	require.NoError(t, err)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ytProxy, err := container.GetProxy(ctx)
	require.NoError(t, err)
	ytSettings := YtClientSettings{YtProxy: ytProxy, Token: yttc.DefaultToken}
	ytClientFactory := YtClientFactory{logger: logger, settings: ytSettings}

	t.Run("use request yt client", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Authorization", fmt.Sprintf("OAuth %s", yttc.DefaultToken))
		ytCli, err := ytClientFactory.FromRequest(request)
		require.NoError(t, err)

		whoAmIResp, err := ytCli.WhoAmI(ctx, nil)
		require.NoError(t, err)
		require.Equal(t, "admin", whoAmIResp.Login)
	})

	t.Run("use yt client from settings", func(t *testing.T) {
		ytCli, err := ytClientFactory.FromSettings()
		require.NoError(t, err)

		whoAmIResp, err := ytCli.WhoAmI(ctx, nil)
		require.NoError(t, err)
		require.Equal(t, "admin", whoAmIResp.Login)
	})
}
