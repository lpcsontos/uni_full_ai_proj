package api

import (
    "context"
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/go-chi/chi/v5"

    "github.com/yourorg/recipemgr/internal/model"
    "github.com/yourorg/recipemgr/internal/usecase"
    "github.com/yourorg/recipemgr/pkg/logger"
)

// Handler holds usecases and logger.
type Handler struct {
    ucs *usecase.Usecases
    l   *logger.Logger
}

// NewHandler constructs a new API handler.
func NewHandler(ucs *usecase.Usecases, l *logger.Logger) *Handler {
    return &Handler{ucs: ucs, l: l}
}

// RegisterRoutes registers all API routes under the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
    r.Post("/auth/register", h.Register)
    r.Post("/auth/login", h.Login)
    r.Post("/auth/refresh", h.Refresh)

    r.Route("/recipes", func(r chi.Router) {
        r.Get("/", h.ListRecipes)
        r.Post("/", h.AuthMiddleware(h.CreateRecipe))
        r.Get("/{id}", h.GetRecipe)
        // PUT/DELETE omitted for brevity; add owner checks in production
    })

    r.Route("/inventory", func(r chi.Router) {
        r.Put("/", h.AuthMiddleware(h.UpsertInventory))
        r.Get("/", h.AuthMiddleware(h.GetInventory))
    })

    r.Post("/shoppinglists/generate", h.AuthMiddleware(h.GenerateShoppingList))

	
    r.Route("/shoppinglists", func(r chi.Router) {
        r.Post("/generate", h.AuthMiddleware(h.GenerateShoppingList))
        r.Get("/", h.AuthMiddleware(h.ListShoppingLists))       // optional: list all for user
        r.Get("/{id}", h.AuthMiddleware(h.GetShoppingList))     // get single list with items
        r.Delete("/{id}", h.AuthMiddleware(h.DeleteShoppingList)) // optional
    })
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, code int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}

// parseUserIDFromAuth extracts the user id stored in the request context by AuthMiddleware.
func parseUserIDFromAuth(r *http.Request) (int64, bool) {
    val := r.Context().Value("uid")
    if val == nil {
        return 0, false
    }
    switch v := val.(type) {
    case int64:
        return v, true
    case float64:
        return int64(v), true
    case string:
        i, err := strconv.ParseInt(v, 10, 64)
        if err != nil {
            return 0, false
        }
        return i, true
    default:
        return 0, false
    }
}

//
// Auth handlers
//

type registerReq struct {
    Email       string `json:"email"`
    Password    string `json:"password"`
    DisplayName string `json:"display_name"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    var req registerReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
        return
    }
    if req.Email == "" || req.Password == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email and password required"})
        return
    }
    id, err := h.ucs.Register(r.Context(), req.Email, req.Password, req.DisplayName)
    if err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

type loginReq struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var req loginReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
        return
    }
    access, refreshRaw, err := h.ucs.Login(r.Context(), req.Email, req.Password)
    if err != nil {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
        return
    }
    // set refresh token as HttpOnly cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "refresh_token",
        Value:    refreshRaw,
        HttpOnly: true,
        Path:     "/",
        Expires:  time.Now().Add(h.ucs.JM().RefreshExp()),
    })
    writeJSON(w, http.StatusOK, map[string]string{"access_token": access})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
    // For the demo we expect refresh token in cookie. Production should validate and rotate tokens.
    c, err := r.Cookie("refresh_token")
    if err != nil {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "no refresh token"})
        return
    }
    raw := c.Value
    // Hash token for DB lookup
    hashed := h.ucs.JM().HashToken(raw)

    // We need to find which user this token belongs to and whether it's valid.
    // The repository/usecase should expose a method to validate and rotate refresh tokens.
    // For brevity, return Not Implemented here and let the usecase implement the full flow.
    _ = hashed
    writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "refresh flow not implemented"})
}

//
// Middleware
//

// AuthMiddleware validates the Authorization header Bearer token and injects uid into context.
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        auth := r.Header.Get("Authorization")
        if auth == "" {
            writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing auth"})
            return
        }
        parts := strings.SplitN(auth, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid auth header"})
            return
        }
        uid, err := h.ucs.JM().VerifyAccessToken(parts[1])
        if err != nil {
            writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
            return
        }
        ctx := context.WithValue(r.Context(), "uid", uid)
        next(w, r.WithContext(ctx))
    }
}

//
// Recipes handlers
//

type createRecipeReq struct {
    Title           string `json:"title"`
    Description     string `json:"description"`
    Servings        int    `json:"servings"`
    PrepTimeMinutes int    `json:"prep_time_minutes"`
    Ingredients     []struct {
        Name     string  `json:"name"`
        Quantity float64 `json:"quantity"`
        Unit     string  `json:"unit"`
    } `json:"ingredients"`
}

func (h *Handler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
    uid, ok := parseUserIDFromAuth(r)
    if !ok {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
        return
    }
    var req createRecipeReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
        return
    }
    rec := &model.Recipe{
        UserID:          uid,
        Title:           req.Title,
        Description:     req.Description,
        Servings:        req.Servings,
        PrepTimeMinutes: req.PrepTimeMinutes,
    }
    for _, ing := range req.Ingredients {
        rec.Ingredients = append(rec.Ingredients, model.RecipeIngredient{
            Name:     ing.Name,
            Quantity: ing.Quantity,
            Unit:     ing.Unit,
        })
    }
    id, err := h.ucs.Repos().CreateRecipe(r.Context(), rec)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id})
}

func (h *Handler) GetRecipe(w http.ResponseWriter, r *http.Request) {
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
        return
    }
    rec, err := h.ucs.Repos().GetRecipe(r.Context(), id)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    if rec == nil {
        writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
        return
    }
    writeJSON(w, http.StatusOK, rec)
}

func (h *Handler) ListRecipes(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("q")
    limit := 20
    offset := 0
    recs, err := h.ucs.Repos().ListRecipes(r.Context(), limit, offset, q)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusOK, recs)
}

//
// Inventory handlers
//

type upsertInventoryReq struct {
    Name          string  `json:"name"`
    Quantity      float64 `json:"quantity"`
    Unit          string  `json:"unit"`
    ExpirationISO string  `json:"expiration_date,omitempty"`
}

func (h *Handler) UpsertInventory(w http.ResponseWriter, r *http.Request) {
    uid, ok := parseUserIDFromAuth(r)
    if !ok {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
        return
    }
    var req upsertInventoryReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
        return
    }
    var exp *time.Time
    if req.ExpirationISO != "" {
        t, err := time.Parse("2006-01-02", req.ExpirationISO)
        if err == nil {
            exp = &t
        }
    }
    if err := h.ucs.Repos().UpsertInventoryItem(r.Context(), uid, req.Name, req.Quantity, req.Unit, exp); err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetInventory(w http.ResponseWriter, r *http.Request) {
    uid, ok := parseUserIDFromAuth(r)
    if !ok {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
        return
    }
    items, err := h.ucs.Repos().GetInventoryForUser(r.Context(), uid)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusOK, items)
}

//
// Shopping list generation
//

type genListReq struct {
    RecipeIDs []int64 `json:"recipe_ids"`
    Name      string  `json:"name"`
}

func (h *Handler) GenerateShoppingList(w http.ResponseWriter, r *http.Request) {
    uid, ok := parseUserIDFromAuth(r)
    if !ok {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
        return
    }
    var req genListReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
        return
    }
    if len(req.RecipeIDs) == 0 {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no recipe ids provided"})
        return
    }
    listID, err := h.ucs.Repos().GenerateShoppingListFromRecipes(r.Context(), uid, req.RecipeIDs, req.Name)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusCreated, map[string]interface{}{"id": listID})
}


// ListShoppingLists returns all shopping lists for the authenticated user.
func (h *Handler) ListShoppingLists(w http.ResponseWriter, r *http.Request) {
    uid, ok := parseUserIDFromAuth(r)
    if !ok {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
        return
    }
    lists, err := h.ucs.Repos().ListShoppingLists(r.Context(), uid)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    writeJSON(w, http.StatusOK, lists)
}

// GetShoppingList returns a shopping list and its items.
func (h *Handler) GetShoppingList(w http.ResponseWriter, r *http.Request) {
    uid, ok := parseUserIDFromAuth(r)
    if !ok {
        writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthenticated"})
        return
    }
    idStr := chi.URLParam(r, "id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
        return
    }
    sl, err := h.ucs.Repos().GetShoppingList(r.Context(), id)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    if sl == nil {
        writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
        return
    }
    // ensure owner
    if sl.UserID != uid {
        writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
        return
    }
    items, err := h.ucs.Repos().GetShoppingListItems(r.Context(), id)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
        return
    }
    resp := map[string]interface{}{
        "list":  sl,
        "items": items,
    }
    writeJSON(w, http.StatusOK, resp)
}

// DeleteShoppingList is a placeholder for deletion (optional).
func (h *Handler) DeleteShoppingList(w http.ResponseWriter, r *http.Request) {
    // implement if you need delete behavior; for now return 501
    writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "delete not implemented"})
}
