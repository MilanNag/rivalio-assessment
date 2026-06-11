"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth-context";
import { Spinner } from "@/components/spinner";

export default function HomePage() {
  const { user, initializing } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (initializing) return;
    router.replace(user ? "/tasks" : "/login");
  }, [user, initializing, router]);

  return (
    <main className="flex min-h-screen items-center justify-center">
      <Spinner label="Loading…" />
    </main>
  );
}
