import React, { useState } from "react";
import api from "../api/client";
import { useNavigate } from "react-router-dom";

export default function Register() {
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const navigate = useNavigate();

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await api.post("/auth/register", { email, password, display_name: displayName });
      alert("Registered. Please login.");
      navigate("/login");
    } catch (err: any) {
      alert(err?.response?.data?.error || "Registration failed");
    }
  }

  return (
    <div className="max-w-md mx-auto bg-white p-6 rounded shadow">
      <h1 className="text-2xl font-semibold mb-4">Register</h1>
      <form onSubmit={submit} className="space-y-3">
        <input className="w-full border px-3 py-2 rounded" placeholder="Email" value={email} onChange={e => setEmail(e.target.value)} />
        <input className="w-full border px-3 py-2 rounded" placeholder="Display name" value={displayName} onChange={e => setDisplayName(e.target.value)} />
        <input type="password" className="w-full border px-3 py-2 rounded" placeholder="Password" value={password} onChange={e => setPassword(e.target.value)} />
        <button className="w-full bg-green-600 text-white py-2 rounded">Create account</button>
      </form>
    </div>
  );
}
