// !!!
// THIS PACKAGE SHOULD BE CALLED ONLY WITH TEST DATABASES URL (or all data will be lost);
// LIKE
//
// POSTGRES_TEST_DB_URL=postgres://root:password@localhost:5432/shop?sslmode=disable REDIS_TEST_DB_URL=localhost:6379 STATUS=testing go test tests/e2e/send_coins_test.go
package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

type authReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Should be runned with TEST_DB_URL env var!
func TestAuthorization(t *testing.T) {
	postgresURL := os.Getenv("POSTGRES_TEST_DB_URL")
	if postgresURL == "" {
		// postgresURL = "postgres://root:password@localhost:5432/shop?sslmode=disable"
		t.Fatal("env TEST_DB_URL cannot be nil")
	}

	db, err := sql.Open("postgres", postgresURL)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// delete all inventory
	_, err = db.ExecContext(ctx, `DELETE FROM Inventory`)
	require.NoError(t, err)

	// delete all transfers
	_, err = db.ExecContext(ctx, `DELETE FROM Transfers`)
	require.NoError(t, err)

	// delete all users
	_, err = db.ExecContext(ctx, `DELETE FROM Users`)
	require.NoError(t, err)

	payload := authReq{
		Username: "testuser",
		Password: "testuser",
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/auth", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	var resp *http.Response
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode != 0 {
			break
		}
		time.Sleep(2 * time.Second)
	}
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("Waited for 200_OK, recieved %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}

	// get response token
	var ans map[string]string
	err = json.NewDecoder(resp.Body).Decode(&ans)
	require.NoError(t, err)
	respToken := ans["token"]

	// get redis token
	redisURL := os.Getenv("REDIS_TEST_DB_URL")
	if redisURL == "" {
		// redisURL = "localhost:6379"
		t.Fatal("env REDIS_TEST_DB_URL cannot be nil")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	dbToken, err := rdb.Get(context.Background(), "testuser").Result()
	require.NoError(t, err)

	require.Equal(t, respToken, dbToken)

	usrC := 0
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Users WHERE username=$1", "testuser").Scan(&usrC)
	require.NoError(t, err)
	if usrC != 1 {
		t.Fatal("no usr in Users table")
	}
}
