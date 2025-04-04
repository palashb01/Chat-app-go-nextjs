"use client";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { fetchMessages } from "@/lib/api";
import { useWebSocket } from "@/context/webSocketProvider";
import { Message } from "@/types/websocket";

export default function ChatPage() {
  const { channelId } = useParams();
  const [messages, setMessages] = useState<Message[]>([]);
  const [text, setText] = useState("");
  const { messages: allMessages, sendMessage } = useWebSocket();

  useEffect(() => {
    fetchMessages(Number(channelId))
      .then(data => setMessages(data || []))
      .catch(error => {
        console.error("Error fetching messages:", error);
        setMessages([]);
      });
  }, [channelId]);

  const send = () => {
    sendMessage(Number(channelId), text);
    setText("");
  };

  return (
    <div className="p-4">
      <h2 className="text-lg font-semibold mb-2">Channel {channelId}</h2>
      <div className="h-[60vh] overflow-y-auto border p-2 rounded mb-2">
        {(!messages || messages.length === 0) ? (
          <div>No messages yet</div>
        ) : (
          [...messages, ...(allMessages || []).filter((m: Message) => m.channelID === Number(channelId))].map((msg, idx) => (
            <div key={idx} className="mb-1">
              <span className="font-semibold">User {msg.senderID}:</span> {msg.content}
            </div>
          ))
        )}
      </div>
      <div className="flex">
        <input
          value={text}
          onChange={(e) => setText(e.target.value)}
          className="flex-1 border rounded px-2 py-1"
          placeholder="Type a message"
        />
        <button
          onClick={send}
          className="ml-2 bg-blue-500 text-white px-4 py-1 rounded"
        >
          Send
        </button>
      </div>
    </div>
  );
}