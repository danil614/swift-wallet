package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"swiftwallet/internal/config"
	"swiftwallet/internal/db"
	"swiftwallet/internal/repository"
	"swiftwallet/internal/router"
	"swiftwallet/internal/service"
)

func TestHTTP_No5xx_OnThousandRPS(t *testing.T) {
	_ = os.Setenv("DB_HOST", "localhost")
	_ = os.Setenv("DB_PORT", "5432")
	_ = os.Setenv("DB_USER", "wallet")
	_ = os.Setenv("DB_PASS", "wallet")
	_ = os.Setenv("DB_NAME", "wallet")

	cfg, _ := config.New()
	pool, err := db.NewPool(cfg)
	require.NoError(t, err)
	defer pool.Close()

	// резетим баланс кошелька
	wID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	_, err = pool.Exec(context.Background(), `UPDATE wallets SET balance=0 WHERE id=$1`, wID)
	require.NoError(t, err)

	svc := service.New(repository.New(pool))
	ts := httptest.NewServer(router.New(svc))
	defer ts.Close()

	const (
		totalOps   = 1000
		concurrent = 80 // семафор << max_connections
	)

	sem := make(chan struct{}, concurrent)
	wg := sync.WaitGroup{}
	wg.Add(totalOps)

	errCh := make(chan error, totalOps)

	bodyTmpl := `{"walletId":"%s","operationType":"DEPOSIT","amount":1}`
	url := ts.URL + "/api/v1/wallet"

	for i := 0; i < totalOps; i++ {
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			body := bytes.NewBufferString(fmt.Sprintf(bodyTmpl, wID))
			resp, err := http.Post(url, "application/json", body)
			if err != nil {
				errCh <- fmt.Errorf("request error: %w", err)
				return
			}
			if resp.StatusCode >= 500 {
				errCh <- fmt.Errorf("unexpected 5xx: %s", resp.Status)
			}
			resp.Body.Close()
		}()
	}
	wg.Wait()
	close(errCh)

	for e := range errCh {
		require.NoError(t, e) // тест упадёт на первом же 5xx / ошибке
	}

	getURL := ts.URL + "/api/v1/wallets/" + wID.String()
	resp, err := http.Get(getURL)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var dto struct {
		Balance int64 `json:"balance"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&dto))
	require.Equal(t, int64(totalOps), dto.Balance)
}
