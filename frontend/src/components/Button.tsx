import React from "react";

export default function Button({ children, className, ...rest }: React.ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <button {...rest} className={`bg-blue-600 text-white px-4 py-2 rounded ${className || ""}`}>
      {children}
    </button>
  );
}
