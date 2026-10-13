import React from "react";
import { Link } from "react-router-dom";

export default function Home() {
  return (
    <div className="max-w-3xl mx-auto text-center">
      <h1 className="text-3xl font-bold mb-4">RecipeMgr</h1>
      <p className="text-gray-600 mb-6">Manage recipes, track your pantry, and generate shopping lists based on what you already have.</p>
      <div className="flex justify-center gap-3">
        <Link to="/recipes" className="bg-green-600 text-white px-4 py-2 rounded">Browse Recipes</Link>
        <Link to="/shoppinglists" className="bg-blue-600 text-white px-4 py-2 rounded">Generate Shopping List</Link>
      </div>
    </div>
  );
}
