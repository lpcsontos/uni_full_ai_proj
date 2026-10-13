package model

import "time"

type User struct {
    ID           int64
    Email        string
    PasswordHash string
    DisplayName  string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

type RecipeIngredient struct {
    IngredientID int64
    Name         string
    Quantity     float64
    Unit         string
}

type Recipe struct {
    ID             int64
    UserID         int64
    Title          string
    Description    string
    Servings       int
    PrepTimeMinutes int
    Ingredients    []RecipeIngredient
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type InventoryItem struct {
    ID             int64
    IngredientID   int64
    Name           string
    Quantity       float64
    Unit           string
    ExpirationDate *time.Time
}

type ShoppingList struct {
    ID        int64
    UserID    int64
    Name      string
    CreatedAt time.Time
}

type ShoppingListItem struct {
    ID           int64   `json:"id"`
    ListID       int64   `json:"list_id"`
    IngredientID int64   `json:"ingredient_id"`
    Name         string  `json:"name"`
    Quantity     float64 `json:"quantity"`
    Unit         string  `json:"unit"`
    Purchased    bool    `json:"purchased"`
}
