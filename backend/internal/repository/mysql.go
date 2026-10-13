package repository

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "math"
    "strings"
    "time"

    "github.com/yourorg/recipemgr/internal/model"
    "github.com/yourorg/recipemgr/pkg/logger"
    "github.com/yourorg/recipemgr/pkg/unitconv"
)



type MySQLRepository struct {
    db *sql.DB
    l  *logger.Logger
}

func NewMySQLRepository(db *sql.DB, l *logger.Logger) *MySQLRepository {
    return &MySQLRepository{db: db, l: l}
}

// --- Users ---
func (r *MySQLRepository) CreateUser(ctx context.Context, u *model.User) (int64, error) {
    res, err := r.db.ExecContext(ctx, `INSERT INTO users (email, password_hash, display_name) VALUES (?, ?, ?)`, u.Email, u.PasswordHash, u.DisplayName)
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}

// ListShoppingLists returns all shopping lists for the given user ordered by created_at desc.
func (r *MySQLRepository) ListShoppingLists(ctx context.Context, userID int64) ([]model.ShoppingList, error) {
    rows, err := r.db.QueryContext(ctx, `
        SELECT id, user_id, name, created_at
        FROM shopping_lists
        WHERE user_id = ?
        ORDER BY created_at DESC
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var lists []model.ShoppingList
    for rows.Next() {
        var sl model.ShoppingList
        if err := rows.Scan(&sl.ID, &sl.UserID, &sl.Name, &sl.CreatedAt); err != nil {
            return nil, err
        }
        lists = append(lists, sl)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return lists, nil
}

func (r *MySQLRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id, email, password_hash, display_name, created_at, updated_at FROM users WHERE email = ?`, email)
    var u model.User
    if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &u, nil
}

func (r *MySQLRepository) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id, email, password_hash, display_name, created_at, updated_at FROM users WHERE id = ?`, id)
    var u model.User
    if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &u, nil
}

// --- Refresh tokens ---
func (r *MySQLRepository) StoreRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
    _, err := r.db.ExecContext(ctx, `INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)`, userID, tokenHash, expiresAt)
    return err
}

func (r *MySQLRepository) DeleteRefreshToken(ctx context.Context, userID int64, tokenHash string) error {
    _, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE user_id = ? AND token_hash = ?`, userID, tokenHash)
    return err
}

func (r *MySQLRepository) FindRefreshToken(ctx context.Context, userID int64, tokenHash string) (bool, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id FROM refresh_tokens WHERE user_id = ? AND token_hash = ? AND expires_at > NOW()`, userID, tokenHash)
    var id int64
    if err := row.Scan(&id); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

// --- Ingredients helper ---
func (r *MySQLRepository) EnsureIngredient(ctx context.Context, name string) (int64, error) {
    // try select
    var id int64
    err := r.db.QueryRowContext(ctx, `SELECT id FROM ingredients WHERE name = ?`, name).Scan(&id)
    if err == nil {
        return id, nil
    }
    if !errors.Is(err, sql.ErrNoRows) {
        return 0, err
    }
    res, err := r.db.ExecContext(ctx, `INSERT INTO ingredients (name) VALUES (?)`, name)
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}

// --- Recipes ---
func (r *MySQLRepository) CreateRecipe(ctx context.Context, rec *model.Recipe) (int64, error) {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return 0, err
    }
    res, err := tx.ExecContext(ctx, `INSERT INTO recipes (user_id, title, description, servings, prep_time_minutes) VALUES (?, ?, ?, ?, ?)`,
        rec.UserID, rec.Title, rec.Description, rec.Servings, rec.PrepTimeMinutes)
    if err != nil {
        tx.Rollback()
        return 0, err
    }
    rid, _ := res.LastInsertId()
    for _, ri := range rec.Ingredients {
        ingID, err := r.EnsureIngredient(ctx, ri.Name)
        if err != nil {
            tx.Rollback()
            return 0, err
        }
        _, err = tx.ExecContext(ctx, `INSERT INTO recipe_ingredients (recipe_id, ingredient_id, quantity, unit) VALUES (?, ?, ?, ?)`,
            rid, ingID, ri.Quantity, ri.Unit)
        if err != nil {
            tx.Rollback()
            return 0, err
        }
    }
    if err := tx.Commit(); err != nil {
        return 0, err
    }
    return rid, nil
}

func (r *MySQLRepository) GetRecipe(ctx context.Context, id int64) (*model.Recipe, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id, user_id, title, description, servings, prep_time_minutes, created_at, updated_at FROM recipes WHERE id = ?`, id)
    var rec model.Recipe
    if err := row.Scan(&rec.ID, &rec.UserID, &rec.Title, &rec.Description, &rec.Servings, &rec.PrepTimeMinutes, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    rows, err := r.db.QueryContext(ctx, `SELECT ri.quantity, ri.unit, i.id, i.name FROM recipe_ingredients ri JOIN ingredients i ON ri.ingredient_id = i.id WHERE ri.recipe_id = ?`, id)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    for rows.Next() {
        var q float64
        var unit string
        var iid int64
        var name string
        if err := rows.Scan(&q, &unit, &iid, &name); err != nil {
            return nil, err
        }
        rec.Ingredients = append(rec.Ingredients, model.RecipeIngredient{
            IngredientID: iid,
            Name:         name,
            Quantity:     q,
            Unit:         unit,
        })
    }
    return &rec, nil
}

func (r *MySQLRepository) ListRecipes(ctx context.Context, limit, offset int, q string) ([]model.Recipe, error) {
    // simple title search
    rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, title, description, servings, prep_time_minutes, created_at, updated_at FROM recipes WHERE title LIKE ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, "%"+q+"%", limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []model.Recipe
    for rows.Next() {
        var rec model.Recipe
        if err := rows.Scan(&rec.ID, &rec.UserID, &rec.Title, &rec.Description, &rec.Servings, &rec.PrepTimeMinutes, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
            return nil, err
        }
        out = append(out, rec)
    }
    return out, nil
}

// --- Inventory upsert ---
func (r *MySQLRepository) UpsertInventoryItem(ctx context.Context, userID int64, ingName string, qty float64, unit string, exp *time.Time) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    ingID, err := r.EnsureIngredient(ctx, ingName)
    if err != nil {
        tx.Rollback()
        return err
    }
    // try update
    res, err := tx.ExecContext(ctx, `UPDATE inventory_items SET quantity = ?, unit = ?, expiration_date = ? WHERE user_id = ? AND ingredient_id = ?`, qty, unit, exp, userID, ingID)
    if err != nil {
        tx.Rollback()
        return err
    }
    ra, _ := res.RowsAffected()
    if ra == 0 {
        _, err = tx.ExecContext(ctx, `INSERT INTO inventory_items (user_id, ingredient_id, quantity, unit, expiration_date) VALUES (?, ?, ?, ?, ?)`, userID, ingID, qty, unit, exp)
        if err != nil {
            tx.Rollback()
            return err
        }
    }
    return tx.Commit()
}

func (r *MySQLRepository) GetInventoryForUser(ctx context.Context, userID int64) ([]model.InventoryItem, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT ii.id, ii.ingredient_id, i.name, ii.quantity, ii.unit, ii.expiration_date FROM inventory_items ii JOIN ingredients i ON ii.ingredient_id = i.id WHERE ii.user_id = ?`, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []model.InventoryItem
    for rows.Next() {
        var it model.InventoryItem
        var exp sql.NullTime
        if err := rows.Scan(&it.ID, &it.IngredientID, &it.Name, &it.Quantity, &it.Unit, &exp); err != nil {
            return nil, err
        }
        if exp.Valid {
            t := exp.Time
            it.ExpirationDate = &t
        }
        out = append(out, it)
    }
    return out, nil
}

// --- Shopping list generation ---
func (r *MySQLRepository) GenerateShoppingListFromRecipes(ctx context.Context, userID int64, recipeIDs []int64, listName string) (int64, error) {
    if len(recipeIDs) == 0 {
        return 0, fmt.Errorf("no recipes provided")
    }

    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return 0, err
    }
    defer func() {
        // ensure rollback on panic; commit/rollback handled below
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p)
        }
    }()

    // Step 1: aggregate raw requirements per ingredient per unit (unit normalized)
    type entry struct {
        ingredientID int64
        name         string
        qty          float64
        unit         string // normalized unit
    }
    // map ingredientID -> slice of entries (different units possible)
    reqMap := map[int64][]entry{}

    for _, rid := range recipeIDs {
        rows, err := tx.QueryContext(ctx, `
            SELECT ri.quantity, ri.unit, i.id, i.name
            FROM recipe_ingredients ri
            JOIN ingredients i ON ri.ingredient_id = i.id
            WHERE ri.recipe_id = ?`, rid)
        if err != nil {
            tx.Rollback()
            return 0, err
        }
        for rows.Next() {
            var q float64
            var unit sql.NullString
            var iid int64
            var name string
            if err := rows.Scan(&q, &unit, &iid, &name); err != nil {
                rows.Close()
                tx.Rollback()
                return 0, err
            }
            u := ""
            if unit.Valid {
                u = unitconv.NormalizeUnit(unit.String)
            }
            // append entry
            reqMap[iid] = append(reqMap[iid], entry{
                ingredientID: iid,
                name:         name,
                qty:          q,
                unit:         u,
            })
        }
        rows.Close()
    }

    // Step 2: for each ingredient, try to merge entries into as few items as possible using unit conversions.
    // We'll produce a flat list of final items keyed by "ingredientID|unit".
    finalItems := map[string]entry{} // key: fmt.Sprintf("%d|%s", iid, unit)

    for iid, entries := range reqMap {
        if len(entries) == 0 {
            continue
        }
        // start with the first entry as the preferred unit
        preferred := entries[0]
        key := fmt.Sprintf("%d|%s", iid, preferred.unit)
        finalItems[key] = preferred

        // merge remaining entries
        for i := 1; i < len(entries); i++ {
            e := entries[i]
            // try to convert e.qty from e.unit to preferred.unit
            if preferred.unit == "" && e.unit != "" {
                // if preferred has no unit but e has, adopt e's unit as preferred
                // move existing preferred qty into new unit if possible (preferred.qty -> e.unit)
                // attempt conversion of preferred.qty -> e.unit
                if preferred.qty != 0 {
                    if conv, convErr := unitconv.Convert(preferred.qty, preferred.unit, e.unit); convErr == nil {
                        // replace preferred with converted value in e.unit
                        preferred.unit = e.unit
                        preferred.qty = conv
                        // update map key
                        oldKey := fmt.Sprintf("%d|%s", iid, "")
                        delete(finalItems, oldKey)
                        key = fmt.Sprintf("%d|%s", iid, preferred.unit)
                        finalItems[key] = preferred
                    } else {
                        // cannot convert preferred to e.unit; keep both as separate entries
                        newKey := fmt.Sprintf("%d|%s", iid, e.unit)
                        if ex, ok := finalItems[newKey]; ok {
                            ex.qty += e.qty
                            finalItems[newKey] = ex
                        } else {
                            finalItems[newKey] = e
                        }
                        continue
                    }
                } else {
                    // preferred qty zero, just adopt e.unit
                    preferred.unit = e.unit
                    key = fmt.Sprintf("%d|%s", iid, preferred.unit)
                    finalItems[key] = preferred
                }
                continue
            }

            // normal case: both have units (possibly empty)
            convQty, convErr := unitconv.Convert(e.qty, e.unit, preferred.unit)
            if convErr == nil {
                // merge into preferred
                ex := finalItems[key]
                ex.qty += convQty
                finalItems[key] = ex
            } else {
                // incompatible units: keep as separate item keyed by its unit
                newKey := fmt.Sprintf("%d|%s", iid, e.unit)
                if ex, ok := finalItems[newKey]; ok {
                    ex.qty += e.qty
                    finalItems[newKey] = ex
                } else {
                    finalItems[newKey] = e
                }
            }
        }
    }

    // Step 3: subtract inventory where possible (attempt conversion of inventory -> item unit)
    for k, it := range finalItems {
        // fetch inventory for this user & ingredient
        var invQty sql.NullFloat64
        var invUnit sql.NullString
        err := tx.QueryRowContext(ctx, `SELECT quantity, unit FROM inventory_items WHERE user_id = ? AND ingredient_id = ?`, userID, it.ingredientID).Scan(&invQty, &invUnit)
        if err != nil && err != sql.ErrNoRows {
            tx.Rollback()
            return 0, err
        }
        if err == nil && invQty.Valid {
            invUnitNorm := unitconv.NormalizeUnit(invUnit.String)
            itemUnit := it.unit
            // if both units empty, subtract directly
            if invUnitNorm == "" && itemUnit == "" {
                it.qty = it.qty - invQty.Float64
                if it.qty < 0 {
                    it.qty = 0
                }
                finalItems[k] = it
                continue
            }
            // try convert inventory to item unit
            convInv, convErr := unitconv.Convert(invQty.Float64, invUnitNorm, itemUnit)
            if convErr == nil {
                it.qty = it.qty - convInv
                if it.qty < 0 {
                    it.qty = 0
                }
                finalItems[k] = it
                continue
            }
            // conversion failed: try converting item qty to inventory unit and subtract there,
            // then convert remainder back to item unit if possible.
            if invUnitNorm != "" && itemUnit != "" {
                convItemToInv, convErr2 := unitconv.Convert(it.qty, itemUnit, invUnitNorm)
                if convErr2 == nil {
                    remainingInInvUnit := convItemToInv - invQty.Float64
                    if remainingInInvUnit <= 0 {
                        // inventory covers the item fully
                        it.qty = 0
                        finalItems[k] = it
                        continue
                    }
                    // convert remaining back to item unit
                    convBack, convErr3 := unitconv.Convert(remainingInInvUnit, invUnitNorm, itemUnit)
                    if convErr3 == nil {
                        it.qty = convBack
                        finalItems[k] = it
                        continue
                    }
                }
            }
            // if we reach here, we couldn't subtract due to incompatible units; leave qty as-is
        }
    }

    // Step 4: create shopping list and insert items with qty > 0
    res, err := tx.ExecContext(ctx, `INSERT INTO shopping_lists (user_id, name) VALUES (?, ?)`, userID, listName)
    if err != nil {
        tx.Rollback()
        return 0, err
    }
    listID, err := res.LastInsertId()
    if err != nil {
        tx.Rollback()
        return 0, err
    }

    // insert items
    insertStmt, err := tx.PrepareContext(ctx, `INSERT INTO shopping_list_items (list_id, ingredient_id, quantity, unit) VALUES (?, ?, ?, ?)`)
    if err != nil {
        tx.Rollback()
        return 0, err
    }
    defer insertStmt.Close()

    for _, it := range finalItems {
        // skip zero or negative quantities
        if it.qty <= 0 {
            continue
        }
        // round to 2 decimals
        q := math.Round(it.qty*100) / 100
        // ensure unit is not empty string -> store as NULL or empty; we'll store empty string
        unitToStore := strings.TrimSpace(it.unit)
        if unitToStore == "" {
            unitToStore = ""
        }
        if _, err := insertStmt.ExecContext(ctx, listID, it.ingredientID, q, unitToStore); err != nil {
            tx.Rollback()
            return 0, err
        }
    }

    if err := tx.Commit(); err != nil {
        return 0, err
    }
    return listID, nil
}

// GetShoppingList returns the shopping list metadata or nil if not found.
func (r *MySQLRepository) GetShoppingList(ctx context.Context, id int64) (*model.ShoppingList, error) {
    row := r.db.QueryRowContext(ctx, `SELECT id, user_id, name, created_at FROM shopping_lists WHERE id = ?`, id)
    var sl model.ShoppingList
    if err := row.Scan(&sl.ID, &sl.UserID, &sl.Name, &sl.CreatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &sl, nil
}

// GetShoppingListItems returns items for a shopping list with ingredient names joined.
func (r *MySQLRepository) GetShoppingListItems(ctx context.Context, listID int64) ([]model.ShoppingListItem, error) {
    rows, err := r.db.QueryContext(ctx, `
        SELECT s.id, s.list_id, s.ingredient_id, i.name, s.quantity, s.unit, s.purchased
        FROM shopping_list_items s
        JOIN ingredients i ON s.ingredient_id = i.id
        WHERE s.list_id = ?`, listID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []model.ShoppingListItem
    for rows.Next() {
        var it model.ShoppingListItem
        if err := rows.Scan(&it.ID, &it.ListID, &it.IngredientID, &it.Name, &it.Quantity, &it.Unit, &it.Purchased); err != nil {
            return nil, err
        }
        out = append(out, it)
    }
    return out, nil
}
