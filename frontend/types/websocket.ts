export interface Message {
  type: string;
  channelID: number;
  senderID: number;
  content: string;
  created_at: string;
}

export interface Channel {
  id: number;
  channel_name: string;
  channel_type: string;
  created_at: string;
} 