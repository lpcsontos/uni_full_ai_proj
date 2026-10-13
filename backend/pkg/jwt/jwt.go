
package jwt

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "errors"
    "io"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type Manager struct {
    secret     []byte
    accessExp  time.Duration
    refreshExp time.Duration
}

func NewManager(secret string, accessExp, refreshExp time.Duration) *Manager {
    return &Manager{
        secret:     []byte(secret),
        accessExp:  accessExp,
        refreshExp: refreshExp,
    }
}

func (m *Manager) CreateAccessToken(userID int64) (string, error) {
    claims := jwt.MapClaims{
        "sub": userID,
        "exp": time.Now().Add(m.accessExp).Unix(),
        "iat": time.Now().Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secret)
}

// CreateRefreshToken returns (rawToken, hashedToken, error).
// rawToken is returned to the client; hashedToken should be stored in DB.
func (m *Manager) CreateRefreshToken(userID int64) (string, string, error) {
    // generate secure random bytes
    b := make([]byte, 32)
    if _, err := io.ReadFull(rand.Reader, b); err != nil {
        return "", "", err
    }
    // include user id and timestamp in raw token for traceability if desired
    raw := hex.EncodeToString(b)
    hashed := m.HashToken(raw)
    return raw, hashed, nil
}

// HashToken returns a SHA256 hex of the token for DB storage/comparison.
func (m *Manager) HashToken(token string) string {
    h := sha256.Sum256([]byte(token))
    return hex.EncodeToString(h[:])
}

// VerifyAccessToken returns userID if token valid.
func (m *Manager) VerifyAccessToken(tokenStr string) (int64, error) {
    token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
        if t.Method != jwt.SigningMethodHS256 {
            return nil, errors.New("unexpected signing method")
        }
        return m.secret, nil
    })
    if err != nil {
        return 0, err
    }
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        sub, ok := claims["sub"].(float64)
        if !ok {
            return 0, errors.New("invalid sub")
        }
        return int64(sub), nil
    }
    return 0, errors.New("invalid token")
}

// RefreshExp returns the configured refresh token duration.
func (m *Manager) RefreshExp() time.Duration {
    return m.refreshExp
}

