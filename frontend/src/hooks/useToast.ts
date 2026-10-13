import { useState } from "react";

export function useToast() {
  const [msg, setMsg] = useState<string | null>(null);
  function show(message: string, ms = 3000) {
    setMsg(message);
    setTimeout(() => setMsg(null), ms);
  }
  return { msg, show };
}
