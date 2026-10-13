import React from "react";

export default function Toast({ message }: { message: string | null }) {
  if (!message) return null;
  return (
    <div className="fixed bottom-6 right-6 bg-black text-white px-4 py-2 rounded shadow">
      {message}
    </div>
  );
}
