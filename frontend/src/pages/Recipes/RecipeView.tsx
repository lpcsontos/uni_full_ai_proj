
import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import api from "../../api/client";

/**
 * RecipeView
 * - uses useParams() to get the route param `id`
 * - fetches GET /api/recipes/{id} and normalizes PascalCase fields returned by the backend
 * - renders title, description, ingredients and steps safely
 */
export default function RecipeView() {
  // ensure we actually get the id param from the route
  const params = useParams();
  const id = params.id;

  const [recipe, setRecipe] = useState<any | null>(null);
  const [ingredients, setIngredients] = useState<any[]>([]);
  const [steps, setSteps] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!id) return; // guard: nothing to fetch without id
    setLoading(true);

    api.get(`/recipes/${id}`)
      .then((r) => {
        console.log("GET /recipes/:id raw:", r.data);
        const apiObj = r.data;

        // normalize backend PascalCase -> frontend-friendly shape
        const normalized = {
          id: apiObj?.ID ?? apiObj?.id,
          title: apiObj?.Title ?? apiObj?.title ?? "Untitled",
          description: apiObj?.Description ?? apiObj?.description ?? "",
          servings: apiObj?.Servings ?? apiObj?.servings ?? null,
          prepTimeMinutes: apiObj?.PrepTimeMinutes ?? apiObj?.prepTimeMinutes ?? null,
          ingredients: Array.isArray(apiObj?.Ingredients) ? apiObj.Ingredients.map((ing: any) => ({
            id: ing?.IngredientID ?? ing?.id,
            name: ing?.Name ?? ing?.name ?? "Unnamed",
            quantity: ing?.Quantity ?? ing?.quantity ?? 0,
            unit: ing?.Unit ?? ing?.unit ?? "",
          })) : [],
          steps: Array.isArray(apiObj?.Steps) ? apiObj.Steps : (apiObj?.steps ?? []),
        };

        setRecipe(normalized);
        setIngredients(normalized.ingredients);
        setSteps(Array.isArray(normalized.steps) ? normalized.steps : []);
      })
      .catch((err) => {
        console.error("Failed to fetch recipe:", err);
        setRecipe(null);
        setIngredients([]);
        setSteps([]);
      })
      .finally(() => setLoading(false));
  }, [id]);

  if (!id) return <div className="p-4">No recipe id provided in URL.</div>;
  if (loading) return <div className="p-4">Loading...</div>;
  if (!recipe) return <div className="p-4">Recipe not found.</div>;

  return (
    <div className="bg-white p-4 rounded shadow max-w-3xl mx-auto">
      <h2 className="text-2xl font-semibold">{recipe.title}</h2>
      <p className="text-gray-600">{recipe.description}</p>

      <div className="mt-4">
        <h3 className="font-semibold">Ingredients</h3>
        {ingredients.length === 0 ? (
          <div className="text-sm text-gray-500">No ingredients listed for this recipe.</div>
        ) : (
          <ul className="list-disc pl-6">
            {ingredients.map((ing: any, i: number) => (
              <li key={ing.id ?? i}>
                {ing.name ?? "Unnamed"} — {ing.quantity ?? 0} {ing.unit ?? ""}
              </li>
            ))}
          </ul>
        )}
      </div>

      <div className="mt-4">
        <h3 className="font-semibold">Steps</h3>
        {steps.length === 0 ? (
          <div className="text-sm text-gray-500">No steps provided.</div>
        ) : (
          <ol className="list-decimal pl-6">
            {steps.map((s: string, i: number) => <li key={i}>{s}</li>)}
          </ol>
        )}
      </div>
    </div>
  );
}
