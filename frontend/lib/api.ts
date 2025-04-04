"use client";
import axios from "axios";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function createUserIfNotExists(username: string) {
  const res = await axios.get(`${API_URL}/check_user?username=${username}`);
  if (res.data?.exists) return res.data;

  const createRes = await axios.post(`${API_URL}/users`, { username });
  return createRes.data;
}

export async function fetchChannels(userId: number) {
  const res = await axios.get(`${API_URL}/my_channels?user_id=${userId}`);
  return res.data;
}
export async function fetchMessages(channelId: number) {
  const res = await axios.get(`${API_URL}/fetch_messages?channel_id=${channelId}`);
  return res.data;
}

export async function createChannel(payload: {
  channel_type: "DIRECT" | "GROUP";
  channel_name: string;
  user_ids: number[];
}) {
  const res = await axios.post(`${API_URL}/create_channel`, payload);
  return res.data;
}