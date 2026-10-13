
import React, { useEffect, useState, useRef } from "react";
// src/pages/Recipes/RecipeForm.tsx
import api from "../../api/client";
import { useNavigate } from "react-router-dom";

export default function RecipeForm() {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [servings, setServings] = useState<number>(2);
  const [ingredients, setIngredients] = useState([{ name: "", quantity: 0, unit: "g" }]);
  const navigate = useNavigate();

  function updateIngredient(idx: number, key: string, value: any) {
    const copy = [...ingredients];
    // @ts-ignore
    copy[idx][key] = value;
    setIngredients(copy);
  }

  function addIngredient() {
    setIngredients([...ingredients, { name: "", quantity: 0, unit: "g" }]);
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    const payload = { title, description, servings, ingredients };
    try {
      // Backend documented: POST /api/recipes/ returns 201 { "id": <number> }
      const res = await api.post("/recipes/", payload);
      const id = res.data?.id;
      if (id) {
        navigate(`/recipes/${id}`);
        return;
      }
      // fallback to list if no id
      navigate("/recipes");
    } catch (err: any) {
      console.error("Failed to create recipe:", err);
      alert(err?.response?.data?.error || err?.message || "Failed to create recipe");
    }
  }

  return (
    <div className="max-w-2xl mx-auto bg-white p-6 rounded shadow">
      <h2 className="text-xl font-semibold mb-4">New Recipe</h2>
      <form onSubmit={submit} className="space-y-3">
        <input className="w-full border px-3 py-2 rounded" placeholder="Title" value={title} onChange={e => setTitle(e.target.value)} required />
        <textarea className="w-full border px-3 py-2 rounded" placeholder="Description" value={description} onChange={e => setDescription(e.target.value)} />
        <input type="number" className="w-32 border px-3 py-2 rounded" value={servings} onChange={e => setServings(Number(e.target.value))} min={1} />
        <div>
          <h3 className="font-semibold">Ingredients</h3>
          {ingredients.map((ing, i) => (
            <div key={i} className="grid grid-cols-12 gap-2 items-center mb-2">
              <input className="col-span-6 border px-2 py-1 rounded" placeholder="Name" value={ing.name} onChange={e => updateIngredient(i, "name", e.target.value)} required />
              <input type="number" className="col-span-3 border px-2 py-1 rounded" value={ing.quantity} onChange={e => updateIngredient(i, "quantity", Number(e.target.value))} min={0} />
              <input className="col-span-2 border px-2 py-1 rounded" value={ing.unit} onChange={e => updateIngredient(i, "unit", e.target.value)} />
            </div>
          ))}
          <button type="button" onClick={addIngredient} className="text-sm text-blue-600">Add ingredient</button>
        </div>
        <div>
          <button className="bg-green-600 text-white px-4 py-2 rounded">Create</button>
        </div>
      </form>
    </div>
  );
}

