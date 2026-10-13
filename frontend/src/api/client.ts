
// src/api/client.ts
import axios, { AxiosError } from "axios";

const baseURL = (import.meta.env.VITE_API_BASE as string) || "http://127.0.0.1:8080/api";

const api = axios.create({
  baseURL,
  withCredentials: true, // backend sets refresh cookie
  headers: { "Content-Type": "application/json" },
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem("access_token");
  if (token && config.headers) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (err: AxiosError) => {
    console.error("API error:", err?.response?.status, err?.response?.data || err.message);
    return Promise.reject(err);
  }
);

export default api;

