import React from "react";
import { Routes, Route } from "react-router-dom";
import Layout from "./components/Layout";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Register from "./pages/Register";
import RecipeList from "./pages/Recipes/RecipeList";
import RecipeView from "./pages/Recipes/RecipeView";
import RecipeForm from "./pages/Recipes/RecipeForm";
import Inventory from "./pages/Inventory";
import ShoppingListView from "./pages/ShoppingLists/ShoppingListView";
import GenerateShoppingList from "./pages/ShoppingLists/Generate";
import ProtectedRoute from "./components/ProtectedRoute";

export default function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />

        <Route path="/recipes" element={<RecipeList />} />
        <Route path="/recipes/new" element={<ProtectedRoute><RecipeForm /></ProtectedRoute>} />
        <Route path="/recipes/:id" element={<RecipeView />} />

        <Route path="/inventory" element={<ProtectedRoute><Inventory /></ProtectedRoute>} />

        <Route path="/shoppinglists" element={<ProtectedRoute><GenerateShoppingList /></ProtectedRoute>} />
        <Route path="/shoppinglists/:id" element={<ProtectedRoute><ShoppingListView /></ProtectedRoute>} />
      </Routes>
    </Layout>
  );
}

