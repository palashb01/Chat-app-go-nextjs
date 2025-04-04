# 💬 Realtime Chat App (Go + Next.js + WebSockets)

This is a full-stack real-time chat application built as a learning project to dive deep into **WebSockets**, **Golang**, and **Next.js**.

It includes:

- 🔧 A **Go (Golang)** backend using Gorilla Mux + WebSocket
- 🌐 A modern **Next.js 15** frontend with Tailwind CSS
- 🧠 **PostgreSQL** via NeonDB for persistent data
- 📡 **Realtime messaging** using WebSockets
- 🧪 Manual API testing with Postman or CLI tools

---

## 🗂️ Project Structure

/backend → Go backend server (Gorilla Mux + WebSocket) 
/frontend → Next.js 15 app (App Router + TailwindCSS)

---

## 🔥 Features

- User login (dummy username for now)
- Create channels (Group or Direct)
- Auto-subscribe to all user channels on WebSocket connect
- Real-time messaging across channels
- Message persistence in Postgres
- JWT-ready backend (can enable later)

---

## 🧪 API Endpoints (Backend)

| Method | Endpoint                         | Description                       |
|--------|----------------------------------|-----------------------------------|
| POST   | `/users`                         | Create or fetch user by username |
| GET    | `/users?username=alice`          | Fetch user ID by username        |
| GET    | `/my_channels?user_id=1`         | Get user's channels               |
| POST   | `/create_channel`                | Create a group/direct channel     |
| GET    | `/fetch_messages?channel_id=1`   | Get messages in a channel         |
| POST   | `/channels/:id/members`          | Add a user to an existing channel |
| GET    | `/`                              | Health check                      |
| GET    | `/ws?user_id=1`                  | WebSocket connection              |

---

## 📦 Technologies Used

### 🔙 Backend (Go)
- Go 1.21+
- Gorilla Mux
- Gorilla WebSocket
- PostgreSQL (via NeonDB)
- JWT-ready (optional)

### 🌐 Frontend (Next.js)
- Next.js 15 (App Router)
- Tailwind CSS
- Axios for API requests
- Global WebSocket Context for real-time sync

---


