import React from "react";
import { Link, useNavigate } from "react-router-dom";

export default function Nav() {
  const navigate = useNavigate();
  const token = localStorage.getItem("access_token");
  return (
    <header className="bg-white shadow">
      <div className="container mx-auto px-4 py-3 flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <Link to="/" className="text-xl font-semibold">RecipeMgr</Link>
          <nav className="hidden md:flex space-x-2 text-sm text-gray-600">
            <Link to="/recipes" className="hover:text-gray-900">Recipes</Link>
            <Link to="/inventory" className="hover:text-gray-900">Inventory</Link>
            <Link to="/shoppinglists" className="hover:text-gray-900">Shopping Lists</Link>
          </nav>
        </div>
        <div>
          {token ? (
            <button
              onClick={() => { localStorage.removeItem("access_token"); navigate("/login"); }}
              className="text-sm bg-red-500 text-white px-3 py-1 rounded"
            >
              Logout
            </button>
          ) : (
            <div className="space-x-2">
              <Link to="/login" className="text-sm px-3 py-1">Login</Link>
              <Link to="/register" className="text-sm bg-blue-600 text-white px-3 py-1 rounded">Register</Link>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
