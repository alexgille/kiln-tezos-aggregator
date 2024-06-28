package api_test

import (
	"context"
	"kiln-tezos-delegation/api"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerRouting(t *testing.T) {
	t.Run("ok on supported endpoint route", func(t *testing.T) {
		port, err := getFreePort()
		require.NoError(t, err)

		hdlMock := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		})

		srv := api.NewServer(":"+strconv.Itoa(port), "/xtz/delegations", hdlMock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// warning: running server in a go routine might introduce a race condition
		go func() {
			err := srv.Start(ctx)
			if err != nil {
				t.Errorf("Could not start HTTP server: %s", err)
			}
		}()

		if ok := waitConnect(port, 5*time.Second); !ok {
			t.Fatal("Connection timeout to localhost server")
		}

		resp, err := http.Get("http://localhost:" + strconv.Itoa(port) + "/xtz/delegations")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ko on unsupported endpoint route", func(t *testing.T) {
		port, err := getFreePort()
		require.NoError(t, err)

		hdlMock := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		})

		srv := api.NewServer(":"+strconv.Itoa(port), "/xtz/delegations", hdlMock)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// warning: running server in a go routine might introduce a race condition
		go func() {
			err = srv.Start(ctx)
			if err != nil {
				t.Errorf("Could not start HTTP server: %s", err)
			}
		}()

		if ok := waitConnect(port, 5*time.Second); !ok {
			t.Fatal("Connection timeout to localhost server")
		}

		resp, err := http.Get("http://localhost:" + strconv.Itoa(port) + "/unsupported")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// getFreePort returns a random available port.
func getFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

// waitConnect tries to connect to the localhost on the given port and returns true if it succeeded.
func waitConnect(port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", "localhost:"+strconv.Itoa(port), timeout)
	if err == nil && conn != nil {
		defer conn.Close()
		return true
	}
	return false
}
