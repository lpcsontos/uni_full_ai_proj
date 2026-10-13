import { useState } from "react";
import api from "../api/client";

export function useAuth() {
  const [user, setUser] = useState<{ id: number; email: string } | null>(null);

  async function register(email: string, password: string, displayName?: string) {
    const res = await api.post("/auth/register", { email, password, display_name: displayName });
    return res.data;
  }

  async function login(email: string, password: string) {
    const res = await api.post("/auth/login", { email, password });
    const token = res.data?.access_token;
    if (token) {
      localStorage.setItem("access_token", token);
      // optionally fetch profile here
      setUser({ id: -1, email });
    }
    return res.data;
  }

  function logout() {
    localStorage.removeItem("access_token");
    setUser(null);
  }

  return { user, register, login, logout, setUser };
}
