// !!!
// THIS PACKAGE SHOULD BE CALLED ONLY WITH TEST_DB (or all data will be lost)
// LIKE
//
// POSTGRES_TEST_DB_URL=postgres://root:password@localhost:5432/shop?sslmode=disable STATUS=testing go test tests/e2e/send_coins_test.go
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

	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

type sendConReq struct {
	ToUser string `json:"toUser"`
	Amount int32  `json:"amount"`
}

// Should be runned with POSTGRES_TEST_DB_URL env var!
func TestSendCoinE2E(t *testing.T) {
	dbURL := os.Getenv("POSTGRES_TEST_DB_URL")
	if dbURL == "" {
		// dbURL = "postgres://root:password@localhost:5432/shop?sslmode=disable"
		t.Fatal("env POSTGRES_TEST_DB_URL cannot be nil")
	}

	db, err := sql.Open("postgres", dbURL)
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

	// Create sendUser
	_, err = db.ExecContext(ctx, `
		INSERT INTO Users (username, password, coins)
		VALUES ($1, $2, $3)
	`, "testuser", "mockpassword", 100) // using username=testuser to easly skip token auth middleware
	require.NoError(t, err)

	// Create recieveUser
	_, err = db.ExecContext(ctx, `
		INSERT INTO Users (username, password, coins)
		VALUES ($1, $2, $3)
	`, "recieveuser", "mockpassword", 100)
	require.NoError(t, err)

	payload := sendConReq{
		ToUser: "recieveuser",
		Amount: 20,
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/api/sendCoin", bytes.NewBuffer(payloadBytes))
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

	time.Sleep(1 * time.Second)

	senderBalance := 0
	err = db.QueryRowContext(ctx, "SELECT coins FROM Users WHERE username=$1 LIMIT 1;", "testuser").Scan(&senderBalance)
	require.NoError(t, err)
	if senderBalance != 80 {
		t.Fatalf("invalid sender balance: %d", senderBalance)
	}

	recieverBalance := 0
	err = db.QueryRowContext(ctx, "SELECT coins FROM Users WHERE username=$1 LIMIT 1;", "recieveuser").Scan(&recieverBalance)
	require.NoError(t, err)
	if recieverBalance != 120 {
		t.Fatalf("invalid reciever balance: %d", recieverBalance)
	}

	txCount := 0
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM Transfers WHERE from_username=$1 AND to_username=$2", "testuser", "recieveuser").Scan(&txCount)
	require.NoError(t, err)
	if txCount != 1 {
		t.Fatalf("invalid transfer count: %d", txCount)
	}
}
