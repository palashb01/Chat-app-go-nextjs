"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { fetchChannels, createChannel } from "@/lib/api";
import { useWebSocket } from "@/context/webSocketProvider";
import { Channel } from "@/types/websocket";

export default function ChannelListPage() {
  const [channels, setChannels] = useState<Channel[]>([]);
  const [newUserIds, setNewUserIds] = useState("");
  const [newChannelName, setNewChannelName] = useState("");
  const [channelType, setChannelType] = useState("GROUP");
  const router = useRouter();
  const { messages } = useWebSocket();

  useEffect(() => {
    const userId = localStorage.getItem("user_id");
    if (!userId) return;

    fetchChannels(Number(userId)).then(data => {
      setChannels(data || []);
    });
  }, []);

  const handleCreateChannel = async () => {
    const userId = Number(localStorage.getItem("user_id"));
    const ids = newUserIds.split(",").map(id => parseInt(id.trim())).filter(id => !isNaN(id));
    if (!ids.includes(userId)) ids.push(userId);

    const data = await createChannel({
      channel_type: channelType as "DIRECT" | "GROUP",
      channel_name: channelType === "GROUP" ? newChannelName : "",
      user_ids: ids
    });

    router.push(`/chat/${data.channel_id}`);
  };

  return (
    <div className="p-4">
      <h2 className="text-xl font-bold mb-4">Your Channels</h2>
      <ul>
        {channels.map((ch) => (
          <li
            key={ch.id}
            className="cursor-pointer hover:underline mb-2"
            onClick={() => router.push(`/chat/${ch.id}`)}
          >
            {ch.channel_name || `Channel ${ch.id}`} ({ch.channel_type})
          </li>
        ))}
      </ul>

      <div className="mt-6 border-t pt-4">
        <h3 className="font-semibold mb-2">Create New Channel</h3>
        <div className="mb-2">
          <select value={channelType} onChange={(e) => setChannelType(e.target.value)} className="border rounded px-2 py-1">
            <option value="GROUP">GROUP</option>
            <option value="DIRECT">DIRECT</option>
          </select>
        </div>
        {channelType === "GROUP" && (
          <input
            type="text"
            className="border px-2 py-1 rounded mb-2 w-full"
            placeholder="Group name"
            value={newChannelName}
            onChange={(e) => setNewChannelName(e.target.value)}
          />
        )}
        <input
          type="text"
          className="border px-2 py-1 rounded w-full"
          placeholder="User IDs comma-separated (e.g. 2,3)"
          value={newUserIds}
          onChange={(e) => setNewUserIds(e.target.value)}
        />
        <button
          onClick={handleCreateChannel}
          className="mt-2 bg-green-600 text-white px-4 py-2 rounded"
        >
          Create Channel
        </button>
      </div>

      <div className="mt-6">
        <h3 className="font-bold">Live Messages</h3>
        {messages.map((msg, i) => (
          <div key={i} className="text-sm">
            📨 {msg.channelID} - User {msg.senderID}: {msg.content}
          </div>
        ))}
      </div>
    </div>
  );
}