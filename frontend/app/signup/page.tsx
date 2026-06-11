"use client";

import { AuthForm } from "@/components/auth-form";
import { useAuth } from "@/lib/auth-context";

export default function SignupPage() {
  const { signup } = useAuth();
  return <AuthForm mode="signup" onSubmit={signup} />;
}
