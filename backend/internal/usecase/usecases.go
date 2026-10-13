package usecase

import (
    "context"
    "errors"
    "time"

    "golang.org/x/crypto/bcrypt"

    "github.com/yourorg/recipemgr/internal/model"
    "github.com/yourorg/recipemgr/pkg/jwt"
    "github.com/yourorg/recipemgr/pkg/logger"
)

type Repositories interface {
    CreateUser(ctx context.Context, u *model.User) (int64, error)
    GetUserByEmail(ctx context.Context, email string) (*model.User, error)
    GetUserByID(ctx context.Context, id int64) (*model.User, error)

    StoreRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
    DeleteRefreshToken(ctx context.Context, userID int64, tokenHash string) error
    FindRefreshToken(ctx context.Context, userID int64, tokenHash string) (bool, error)

    CreateRecipe(ctx context.Context, rec *model.Recipe) (int64, error)
    GetRecipe(ctx context.Context, id int64) (*model.Recipe, error)
    ListRecipes(ctx context.Context, limit, offset int, q string) ([]model.Recipe, error)

    UpsertInventoryItem(ctx context.Context, userID int64, ingName string, qty float64, unit string, exp *time.Time) error
    GetInventoryForUser(ctx context.Context, userID int64) ([]model.InventoryItem, error)

    GenerateShoppingListFromRecipes(ctx context.Context, userID int64, recipeIDs []int64, listName string) (int64, error)

	 
	 ListShoppingLists(ctx context.Context, userID int64) ([]model.ShoppingList, error)
    GetShoppingList(ctx context.Context, id int64) (*model.ShoppingList, error)
    GetShoppingListItems(ctx context.Context, listID int64) ([]model.ShoppingListItem, error)
}

type Usecases struct {
    repos Repositories
    jm    *jwt.Manager
    l     *logger.Logger
}

func NewUsecases(repos Repositories, jm *jwt.Manager, l *logger.Logger) *Usecases {
    return &Usecases{repos: repos, jm: jm, l: l}
}

// Register user
func (u *Usecases) Register(ctx context.Context, email, password, displayName string) (int64, error) {
    existing, err := u.repos.GetUserByEmail(ctx, email)
    if err != nil {
        return 0, err
    }
    if existing != nil {
        return 0, errors.New("email already registered")
    }
    pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return 0, err
    }
    user := &model.User{
        Email:        email,
        PasswordHash: string(pw),
        DisplayName:  displayName,
    }
    return u.repos.CreateUser(ctx, user)
}

// Login returns access token and refresh token (raw)
func (u *Usecases) Login(ctx context.Context, email, password string) (access string, refresh string, err error) {
    user, err := u.repos.GetUserByEmail(ctx, email)
    if err != nil {
        return "", "", err
    }
    if user == nil {
        return "", "", errors.New("invalid credentials")
    }
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return "", "", errors.New("invalid credentials")
    }
    access, err = u.jm.CreateAccessToken(user.ID)
    if err != nil {
        return "", "", err
    }
    raw, hashed, err := u.jm.CreateRefreshToken(user.ID)
    if err != nil {
        return "", "", err
    }
    expires := time.Now().Add(u.jm.RefreshExp())
    if err := u.repos.StoreRefreshToken(ctx, user.ID, hashed, expires); err != nil {
        return "", "", err
    }
    return access, raw, nil
}

// Repos returns the underlying repositories interface.
func (u *Usecases) Repos() Repositories {
    return u.repos
}

// JM returns the jwt manager.
func (u *Usecases) JM() *jwt.Manager {
    return u.jm
}
