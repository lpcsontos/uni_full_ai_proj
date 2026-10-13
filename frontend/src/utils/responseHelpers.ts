
// src/utils/responseHelpers.ts
export function mapRecipeListItem(apiObj: any) {
  // map backend PascalCase to frontend-friendly shape
  return {
    id: apiObj?.ID ?? apiObj?.id,
    userId: apiObj?.UserID ?? apiObj?.userId,
    title: apiObj?.Title ?? apiObj?.title ?? "Untitled",
    description: apiObj?.Description ?? apiObj?.description ?? "",
    servings: apiObj?.Servings ?? apiObj?.servings ?? null,
    prepTimeMinutes: apiObj?.PrepTimeMinutes ?? apiObj?.prepTimeMinutes ?? null,
    // Ingredients in list responses may be null; keep as-is
    ingredients: apiObj?.Ingredients ?? apiObj?.ingredients ?? null,
    createdAt: apiObj?.CreatedAt ?? apiObj?.createdAt ?? null,
    updatedAt: apiObj?.UpdatedAt ?? apiObj?.updatedAt ?? null,
  };
}

export function mapRecipeDetail(apiObj: any) {
  // apiObj is the full recipe object returned by GET /api/recipes/{id}
  const ingredientsRaw = apiObj?.Ingredients ?? apiObj?.ingredients ?? [];
  const ingredients = Array.isArray(ingredientsRaw)
    ? ingredientsRaw.map((ing: any) => ({
        id: ing?.IngredientID ?? ing?.id,
        name: ing?.Name ?? ing?.name ?? "Unnamed",
        quantity: ing?.Quantity ?? ing?.quantity ?? 0,
        unit: ing?.Unit ?? ing?.unit ?? "",
      }))
    : [];

  return {
    id: apiObj?.ID ?? apiObj?.id,
    userId: apiObj?.UserID ?? apiObj?.userId,
    title: apiObj?.Title ?? apiObj?.title ?? "Untitled",
    description: apiObj?.Description ?? apiObj?.description ?? "",
    servings: apiObj?.Servings ?? apiObj?.servings ?? null,
    prepTimeMinutes: apiObj?.PrepTimeMinutes ?? apiObj?.prepTimeMinutes ?? null,
    ingredients,
    createdAt: apiObj?.CreatedAt ?? apiObj?.createdAt ?? null,
    updatedAt: apiObj?.UpdatedAt ?? apiObj?.updatedAt ?? null,
  };
}

export function mapInventoryItem(apiObj: any) {
  return {
    id: apiObj?.ID ?? apiObj?.id,
    userId: apiObj?.UserID ?? apiObj?.userId,
    name: apiObj?.Name ?? apiObj?.name ?? apiObj?.item_name ?? "Unnamed",
    quantity: apiObj?.Quantity ?? apiObj?.quantity ?? apiObj?.qty ?? 0,
    unit: apiObj?.Unit ?? apiObj?.unit ?? "",
    expirationDate: apiObj?.ExpirationDate ?? apiObj?.expiration_date ?? null,
  };
}
