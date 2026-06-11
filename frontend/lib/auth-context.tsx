"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { apiFetch, getToken, setToken } from "./api";
import type { User } from "./types";

interface AuthState {
  user: User | null;
  /** True until the persisted session has been validated on mount. */
  initializing: boolean;
  login: (email: string, password: string) => Promise<void>;
  signup: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthState | null>(null);

interface AuthResponse {
  token: string;
  user: User;
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [initializing, setInitializing] = useState(true);

  // Restore the session from localStorage on first load so a page
  // refresh keeps the user logged in.
  useEffect(() => {
    let cancelled = false;
    async function restore() {
      if (!getToken()) {
        setInitializing(false);
        return;
      }
      try {
        const res = await apiFetch<{ user: User }>("/api/auth/me");
        if (!cancelled) setUser(res.user);
      } catch {
        setToken(null);
      } finally {
        if (!cancelled) setInitializing(false);
      }
    }
    restore();
    return () => {
      cancelled = true;
    };
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await apiFetch<AuthResponse>("/api/auth/login", {
      method: "POST",
      body: { email, password },
    });
    setToken(res.token);
    setUser(res.user);
  }, []);

  const signup = useCallback(async (email: string, password: string) => {
    const res = await apiFetch<AuthResponse>("/api/auth/signup", {
      method: "POST",
      body: { email, password },
    });
    setToken(res.token);
    setUser(res.user);
  }, []);

  const logout = useCallback(() => {
    setToken(null);
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({ user, initializing, login, signup, logout }),
    [user, initializing, login, signup, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
