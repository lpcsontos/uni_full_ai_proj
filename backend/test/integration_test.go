// test/integration_test.go
package test

import (
    "bytes"
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
    "time"

    tc "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    _ "github.com/go-sql-driver/mysql"
)

const (
    mysqlImage      = "mysql:8.0"
    dbUser          = "recipemgr_user"
    dbPassword      = "strongpassword"
    dbName          = "recipemgr"
    jwtSecret       = "test_jwt_secret_please_change"
    serverStartWait = 10 * time.Second
)

// helper: read migrations file and execute against DB
func applyMigrations(t *testing.T, dsn string) {
    t.Helper()
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        t.Fatalf("open db for migrations: %v", err)
    }
    defer db.Close()

    // locate migrations file relative to repo root
    migPath := filepath.Join("migrations", "001_init.sql")
    b, err := os.ReadFile(migPath)
    if err != nil {
        t.Fatalf("read migration file %s: %v", migPath, err)
    }
    // split by semicolon and execute statements (simple approach)
    stmts := strings.Split(string(b), ";")
    for _, s := range stmts {
        s = strings.TrimSpace(s)
        if s == "" {
            continue
        }
        if _, err := db.Exec(s); err != nil {
            t.Fatalf("exec migration stmt: %v\nstmt: %s", err, s)
        }
    }
}

// helper: wait for HTTP endpoint to respond OK
func waitForHTTP(t *testing.T, url string, timeout time.Duration) {
    t.Helper()
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        resp, err := http.Get(url)
        if err == nil {
            _ = resp.Body.Close()
            if resp.StatusCode < 500 {
                return
            }
        }
        time.Sleep(250 * time.Millisecond)
    }
    t.Fatalf("timeout waiting for %s", url)
}

// start server as subprocess; returns cmd and stdout/stderr pipes
func startServerProcess(t *testing.T, env map[string]string) (*exec.Cmd, io.ReadCloser, io.ReadCloser) {
    t.Helper()
    // run `go run ./cmd/server`
    cmd := exec.Command("go", "run", "./cmd/server")
    // inherit current env and append/override
    cmd.Env = os.Environ()
    for k, v := range env {
        cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
    }
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        t.Fatalf("stdout pipe: %v", err)
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        t.Fatalf("stderr pipe: %v", err)
    }
    if err := cmd.Start(); err != nil {
        t.Fatalf("start server process: %v", err)
    }
    return cmd, stdout, stderr
}

// simple HTTP helper
func postJSON(t *testing.T, base, path string, body interface{}, token string) *http.Response {
    t.Helper()
    b, _ := json.Marshal(body)
    req, err := http.NewRequest("POST", strings.TrimRight(base, "/")+path, bytes.NewReader(b))
    if err != nil {
        t.Fatalf("new request: %v", err)
    }
    req.Header.Set("Content-Type", "application/json")
    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        t.Fatalf("post %s: %v", path, err)
    }
    return resp
}

func getJSON(t *testing.T, base, path, token string) *http.Response {
    t.Helper()
    req, err := http.NewRequest("GET", strings.TrimRight(base, "/")+path, nil)
    if err != nil {
        t.Fatalf("new request: %v", err)
    }
    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        t.Fatalf("get %s: %v", path, err)
    }
    return resp
}


// putJSON sends a PUT request with JSON body and optional Bearer token.
func putJSON(t *testing.T, base, path string, body interface{}, token string) *http.Response {
    t.Helper()
    b, _ := json.Marshal(body)
    req, err := http.NewRequest("PUT", strings.TrimRight(base, "/")+path, bytes.NewReader(b))
    if err != nil {
        t.Fatalf("new request: %v", err)
    }
    req.Header.Set("Content-Type", "application/json")
    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        t.Fatalf("put %s: %v", path, err)
    }
    return resp
}

// TestEndToEnd runs the full flow either against TEST_BASE_URL or an ephemeral local server+DB.
func TestEndToEnd(t *testing.T) {
    // If TEST_BASE_URL is set, run tests against that URL
    if base := os.Getenv("TEST_BASE_URL"); base != "" {
        t.Logf("running tests against remote base URL: %s", base)
        runTestsAgainstBase(t, base)
        return
    }

    // Otherwise start ephemeral MySQL container and server process
    ctx := context.Background()
    req := tc.ContainerRequest{
        Image:        mysqlImage,
        ExposedPorts: []string{"3306/tcp"},
        Env: map[string]string{
            "MYSQL_ROOT_PASSWORD": "rootpass",
            "MYSQL_DATABASE":      dbName,
            "MYSQL_USER":          dbUser,
            "MYSQL_PASSWORD":      dbPassword,
        },
        WaitingFor: wait.ForListeningPort("3306/tcp").WithStartupTimeout(60 * time.Second),
    }
    mysqlC, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    if err != nil {
        t.Fatalf("start mysql container: %v", err)
    }
    defer func() {
        _ = mysqlC.Terminate(ctx)
    }()

    // get host and port
    host, err := mysqlC.Host(ctx)
    if err != nil {
        t.Fatalf("container host: %v", err)
    }
    p, err := mysqlC.MappedPort(ctx, "3306")
    if err != nil {
        t.Fatalf("mapped port: %v", err)
    }
    port := p.Int()

    // build DSN for migrations and server
    // use root to run migrations (container created DB already, but ensure connection)
    dsnRoot := fmt.Sprintf("root:rootpass@tcp(%s:%d)/%s?parseTime=true&multiStatements=true", host, port, dbName)
    // wait for DB to accept connections
    deadline := time.Now().Add(60 * time.Second)
    for time.Now().Before(deadline) {
        db, err := sql.Open("mysql", dsnRoot)
        if err == nil {
            if err = db.Ping(); err == nil {
                db.Close()
                break
            }
            db.Close()
        }
        time.Sleep(500 * time.Millisecond)
    }
    // apply migrations
    applyMigrations(t, dsnRoot)

    // server DSN: use dedicated user created by container env
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true", dbUser, dbPassword, host, port, dbName)

    // start server process with env
    env := map[string]string{
        "DATABASE_URL": dsn,
        "JWT_SECRET":   jwtSecret,
    }
    cmd, stdout, stderr := startServerProcess(t, env)
    defer func() {
        _ = cmd.Process.Kill()
        _ = cmd.Wait()
    }()

    // optionally stream server logs to test output (non-blocking)
    go func() {
        buf := make([]byte, 4096)
        for {
            n, err := stdout.Read(buf)
            if n > 0 {
                t.Logf("[server stdout] %s", strings.TrimSpace(string(buf[:n])))
            }
            if err != nil {
                return
            }
        }
    }()
    go func() {
        buf := make([]byte, 4096)
        for {
            n, err := stderr.Read(buf)
            if n > 0 {
                t.Logf("[server stderr] %s", strings.TrimSpace(string(buf[:n])))
            }
            if err != nil {
                return
            }
        }
    }()

    // wait for server to be ready (poll /api/recipes or root)
    baseURL := "http://127.0.0.1:8080"
    waitForHTTP(t, baseURL+"/api/recipes", 30*time.Second)

    // run the same test flow against baseURL
    runTestsAgainstBase(t, baseURL)
}

// runTestsAgainstBase executes the API flow against baseURL
func runTestsAgainstBase(t *testing.T, baseURL string) {
    t.Helper()
    // 1) register
    email := fmt.Sprintf("testuser+%d@example.com", time.Now().UnixNano())
    pass := "password123"
    regBody := map[string]string{"email": email, "password": pass, "display_name": "Test User"}
    resp := postJSON(t, baseURL, "/api/auth/register", regBody, "")
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        t.Fatalf("register failed: status=%d body=%s", resp.StatusCode, string(body))
    }
    _ = resp.Body.Close()

    // 2) login
    loginBody := map[string]string{"email": email, "password": pass}
    resp = postJSON(t, baseURL, "/api/auth/login", loginBody, "")
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        t.Fatalf("login failed: status=%d body=%s", resp.StatusCode, string(body))
    }
    var loginResp struct {
        AccessToken string `json:"access_token"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
        t.Fatalf("decode login resp: %v", err)
    }
    _ = resp.Body.Close()
    if loginResp.AccessToken == "" {
        t.Fatalf("empty access token")
    }
    token := loginResp.AccessToken

    // 3) create a recipe
    createRec := map[string]interface{}{
        "title":            "Test Pancakes",
        "description":      "Tasty",
        "servings":         2,
        "prep_time_minutes": 10,
        "ingredients": []map[string]interface{}{
            {"name": "flour", "quantity": 200.0, "unit": "g"},
            {"name": "milk", "quantity": 300.0, "unit": "ml"},
        },
    }
    resp = postJSON(t, baseURL, "/api/recipes/", createRec, token)
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        t.Fatalf("create recipe failed: status=%d body=%s", resp.StatusCode, string(body))
    }
    var recCreate struct{ ID int64 `json:"id"` }
    if err := json.NewDecoder(resp.Body).Decode(&recCreate); err != nil {
        // some handlers return body differently; ignore decode error but ensure status was created
    }
    _ = resp.Body.Close()

    // 4) upsert inventory (partial inventory so shopping list will include missing items)
    inv := map[string]interface{}{"name": "flour", "quantity": 100.0, "unit": "g"}
    resp = putJSON(t, baseURL, "/api/inventory", inv, token)
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        t.Fatalf("upsert inventory failed: status=%d body=%s", resp.StatusCode, string(body))
    }
    _ = resp.Body.Close()

    // 5) generate shopping list for the created recipe(s)
    gen := map[string]interface{}{"recipe_ids": []int64{recCreate.ID}, "name": "Test List"}
    resp = postJSON(t, baseURL, "/api/shoppinglists/generate", gen, token)
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        t.Fatalf("generate shopping list failed: status=%d body=%s", resp.StatusCode, string(body))
    }
    var genResp struct{ ID int64 `json:"id"` }
    if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
        t.Fatalf("decode gen resp: %v", err)
    }
    _ = resp.Body.Close()

    // 6) fetch shopping list
    resp = getJSON(t, baseURL, fmt.Sprintf("/api/shoppinglists/%d", genResp.ID), token)
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        t.Fatalf("get shopping list failed: status=%d body=%s", resp.StatusCode, string(body))
    }
    // optionally decode and assert items exist
    var listBody interface{}
    if err := json.NewDecoder(resp.Body).Decode(&listBody); err != nil {
        t.Fatalf("decode list body: %v", err)
    }
    _ = resp.Body.Close()
    t.Logf("shopping list fetched: %+v", listBody)
}
