"use client";
import { createContext, useContext, useEffect, useRef, useState } from "react";
import { fetchChannels } from "@/lib/api";
import { Message, Channel } from "@/types/websocket";

interface WebSocketContextType {
  messages: Message[];
  sendMessage: (channelID: number, text: string) => void;
  connected: boolean;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }
  return context;
};

export function WebSocketProvider({ children }: { children: React.ReactNode }) {
  const wsRef = useRef<WebSocket | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    const userId = Number(localStorage.getItem("user_id"));
    if (!userId) return;

    const ws = new WebSocket(`ws://localhost:8080/ws?user_id=${userId}`);
    wsRef.current = ws;

    ws.onopen = async () => {
      setConnected(true);
      try {
        const channels = await fetchChannels(userId) || [];
        channels.forEach((ch: Channel) => {
          ws.send(JSON.stringify({ type: "subscribe", channelID: ch.id }));
        });
      } catch (error) {
        console.error("Error fetching channels:", error);
      }
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      setMessages((prev) => [...prev, msg]);
    };

    return () => {
      ws.close();
    };
  }, []);

  const sendMessage = (channelID: number, text: string) => {
    wsRef.current?.send(
      JSON.stringify({ type: "message", channelID, text })
    );
  };

  return (
    <WebSocketContext.Provider value={{ messages, sendMessage, connected }}>
      {children}
    </WebSocketContext.Provider>
  );
} 