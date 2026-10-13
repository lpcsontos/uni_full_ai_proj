export interface User {
  id: number;
  email: string;
  display_name?: string;
}

export interface Recipe {
  id?: number;
  user_id?: number;
  title: string;
  description?: string;
  servings?: number;
  prep_time_minutes?: number;
  ingredients?: RecipeIngredient[];
}

export interface RecipeIngredient {
  id?: number;
  recipe_id?: number;
  ingredient_id?: number;
  name: string;
  quantity: number;
  unit: string;
}

export interface InventoryItem {
  id?: number;
  user_id?: number;
  ingredient_id?: number;
  name: string;
  quantity: number;
  unit: string;
  expiration_date?: string;
}

export interface ShoppingList {
  id?: number;
  user_id?: number;
  name: string;
  created_at?: string;
}

export interface ShoppingListItem {
  id?: number;
  list_id?: number;
  ingredient_id?: number;
  name: string;
  quantity: number;
  unit: string;
  purchased?: boolean;
}
