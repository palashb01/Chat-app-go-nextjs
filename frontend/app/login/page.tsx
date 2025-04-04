"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { createUserIfNotExists } from "@/lib/api";

export default function LoginPage() {
  const [username, setUsername] = useState("");
  const router = useRouter();

  const handleLogin = async () => {
    const res = await createUserIfNotExists(username);
    localStorage.setItem("user_id", res.id);
    router.push("/channels");
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center">
      <h1 className="text-2xl font-bold mb-4">Login</h1>
      <input
        type="text"
        className="border px-4 py-2 rounded mb-2"
        placeholder="Enter username"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
      />
      <button onClick={handleLogin} className="bg-blue-500 text-white px-4 py-2 rounded">
        Login
      </button>
    </div>
  );
}