
// src/pages/Inventory.tsx
import React, { useEffect, useState } from "react";
import api from "../api/client";

export default function Inventory() {
  const [items, setItems] = useState<any[]>([]);
  const [name, setName] = useState("");
  const [quantity, setQuantity] = useState<number>(0);
  const [unit, setUnit] = useState("g");
  const [loading, setLoading] = useState(false);

  async function fetchInventory() {
    setLoading(true);
    try {
      const r = await api.get("/inventory");
      console.log("GET /inventory raw:", r.data);
      const raw = Array.isArray(r.data) ? r.data : (r.data?.items ?? []);
      // map backend PascalCase -> frontend-friendly fields if needed
      const normalized = raw.map((it: any) => ({
        id: it?.ID ?? it?.id,
        name: it?.Name ?? it?.name ?? "Unnamed",
        quantity: it?.Quantity ?? it?.quantity ?? 0,
        unit: it?.Unit ?? it?.unit ?? "",
      }));
      setItems(normalized);
    } catch (err) {
      console.error("Failed to fetch inventory:", err);
      setItems([]);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchInventory();
  }, []);

  async function upsert(e: React.FormEvent) {
    e.preventDefault();
    try {
      await api.put("/inventory", { name, quantity, unit });
      await fetchInventory();
      setName(""); setQuantity(0); setUnit("g");
    } catch (err) {
      console.error("Failed to update inventory:", err);
      alert("Failed to update inventory");
    }
  }

  return (
    <div className="max-w-2xl mx-auto p-4">
      <div className="bg-white p-4 rounded shadow mb-4">
        <h3 className="font-semibold mb-2">Add or update item</h3>
        <form onSubmit={upsert} className="grid grid-cols-4 gap-2">
          <input className="col-span-2 border px-2 py-1 rounded" placeholder="Name" value={name} onChange={e => setName(e.target.value)} required />
          <input type="number" className="col-span-1 border px-2 py-1 rounded" value={quantity} onChange={e => setQuantity(Number(e.target.value))} min={0} />
          <input className="col-span-1 border px-2 py-1 rounded" value={unit} onChange={e => setUnit(e.target.value)} />
          <div className="col-span-4">
            <button className="bg-blue-600 text-white px-3 py-1 rounded mt-2">Save</button>
          </div>
        </form>
      </div>

      <div className="bg-white p-4 rounded shadow">
        <h3 className="font-semibold mb-2">Inventory</h3>
        {loading ? <div>Loading...</div> : null}
        <ul className="space-y-2">
          {items.map(it => (
            <li key={it.id ?? it.name} className="flex justify-between">
              <div>
                <div className="font-medium">{it.name}</div>
                <div className="text-sm text-gray-500">{it.quantity} {it.unit}</div>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

