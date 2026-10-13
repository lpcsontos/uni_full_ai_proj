
import React, { useEffect, useState } from "react";
// src/pages/ShoppingLists/Generate.tsx
import api from "../../api/client";
import { Link } from "react-router-dom";

export default function GenerateShoppingList() {
  const [recipes, setRecipes] = useState<any[]>([]);
  const [selected, setSelected] = useState<number[]>([]);
  const [name, setName] = useState("");

  useEffect(() => {
    // GET /api/recipes returns array
    api.get("/recipes").then(r => {
      const data = Array.isArray(r.data) ? r.data : (r.data?.recipes ?? []);
      setRecipes(Array.isArray(data) ? data : []);
    }).catch(() => setRecipes([]));
  }, []);

  async function generate(e: React.FormEvent) {
    e.preventDefault();
    if (selected.length === 0) return alert("Select at least one recipe");
    try {
      // POST /api/shoppinglists/generate returns 201 { "id": <number> }
      const res = await api.post("/shoppinglists/generate", { recipe_ids: selected, name });
      const id = res.data?.id;
      if (id) {
        window.location.href = `/shoppinglists/${id}`;
        return;
      }
      window.location.href = "/shoppinglists";
    } catch {
      alert("Failed to generate list");
    }
  }

  function toggle(id: number) {
    setSelected(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id]);
  }

  return (
    <div className="max-w-3xl mx-auto">
      <h2 className="text-xl font-semibold mb-4">Generate Shopping List</h2>
      <form onSubmit={generate} className="bg-white p-4 rounded shadow space-y-3">
        <input className="w-full border px-3 py-2 rounded" placeholder="List name" value={name} onChange={e => setName(e.target.value)} />
        <div>
          <h3 className="font-semibold mb-2">Select recipes</h3>
          <div className="grid md:grid-cols-2 gap-2">
            {recipes.map(r => (
              <label key={r.id} className="bg-gray-50 p-2 rounded flex items-center">
                <input type="checkbox" checked={selected.includes(r.id)} onChange={() => toggle(r.id)} className="mr-2" />
                <div>
                  <div className="font-medium">{r.title}</div>
                  <div className="text-sm text-gray-500">{r.description}</div>
                </div>
              </label>
            ))}
          </div>
        </div>
        <div>
          <button className="bg-blue-600 text-white px-4 py-2 rounded">Generate</button>
          <Link to="/shoppinglists" className="ml-3 text-sm text-gray-600">My lists</Link>
        </div>
      </form>
    </div>
  );
}

