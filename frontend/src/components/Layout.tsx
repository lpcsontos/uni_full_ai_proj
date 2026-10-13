import React from "react";
import Nav from "./Nav";

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col">
      <Nav />
      <main className="container mx-auto px-4 py-6 flex-1">{children}</main>
      <footer className="bg-white border-t py-4 text-center text-sm text-gray-500">
        © RecipeMgr
      </footer>
    </div>
  );
}
