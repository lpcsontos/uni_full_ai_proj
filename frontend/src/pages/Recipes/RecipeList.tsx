
import React, { useEffect, useState } from "react";
import api from "../../api/client";
import { Link } from "react-router-dom";

export default function RecipeList() {
  const [recipes, setRecipes] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  async function fetchRecipes() {
    setLoading(true);
    try {
      // Backend returns an array of PascalCase objects
      const r = await api.get("/recipes");
      console.log("GET /recipes raw:", r.data);
      const raw = Array.isArray(r.data) ? r.data : (r.data?.recipes ?? []);
      const normalized = raw.map((apiObj: any) => ({
        id: apiObj?.ID ?? apiObj?.id,
        title: apiObj?.Title ?? apiObj?.title ?? "Untitled",
        description: apiObj?.Description ?? apiObj?.description ?? "",
        servings: apiObj?.Servings ?? apiObj?.servings ?? null,
      }));
      setRecipes(normalized);
    } catch (err) {
      console.error("Failed to fetch recipes:", err);
      setRecipes([]);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchRecipes();
  }, []);

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">Recipes</h2>
        <Link to="/recipes/new" className="bg-green-600 text-white px-3 py-1 rounded">New</Link>
      </div>

      {loading ? <div>Loading...</div> : null}

      <div className="grid md:grid-cols-2 gap-4">
        {recipes.map((r: any) => (
          <div key={r.id ?? r.title} className="bg-white p-4 rounded shadow">
            <h3 className="font-semibold">{r.title}</h3>
            <p className="text-sm text-gray-600">{r.description}</p>
            <div className="mt-3 flex justify-between items-center">
              <Link to={`/recipes/${r.id}`} className="text-blue-600">View</Link>
              <span className="text-sm text-gray-500">{r.servings ?? "-" } servings</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

