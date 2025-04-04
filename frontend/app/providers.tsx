"use client";

import { WebSocketProvider } from "@/context/webSocketProvider";

export function Providers({ children }: { children: React.ReactNode }) {
  return <WebSocketProvider>{children}</WebSocketProvider>;
} 