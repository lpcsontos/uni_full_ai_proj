
import React, { useEffect, useState } from "react";
// src/pages/ShoppingLists/ShoppingListView.tsx
import { useParams } from "react-router-dom";
import api from "../../api/client";

export default function ShoppingListView() {
  const { id } = useParams();
  const [list, setList] = useState<any>(null);
  const [items, setItems] = useState<any[]>([]);

  useEffect(() => {
    if (!id) return;
    // Backend documented: GET /api/shoppinglists/{id} returns { list: {...}, items: [...] }
    api.get(`/shoppinglists/${id}`).then(r => {
      console.log("GET /shoppinglists/:id response:", r.data);
      setList(r.data?.list ?? null);
      setItems(Array.isArray(r.data?.items) ? r.data.items : []);
    }).catch(err => {
      console.error("Failed to fetch shopping list:", err);
    });
  }, [id]);

  if (!list) return <div>Loading...</div>;
  return (
    <div className="bg-white p-4 rounded shadow max-w-3xl mx-auto">
      <h2 className="text-xl font-semibold mb-2">{list.name}</h2>
      <ul className="space-y-2">
        {items.map(it => (
          <li key={it.id} className="flex justify-between items-center">
            <div>
              <div className="font-medium">{it.name}</div>
              <div className="text-sm text-gray-500">{it.quantity} {it.unit}</div>
            </div>
            <div>
              <input type="checkbox" checked={!!it.purchased} readOnly />
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}

