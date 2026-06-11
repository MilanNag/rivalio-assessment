"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { API_URL, apiFetch, getToken } from "./api";
import type {
  ListMeta,
  Task,
  TaskInput,
  TaskListResponse,
  TaskQuery,
} from "./types";

export const DEFAULT_QUERY: TaskQuery = {
  status: "",
  q: "",
  sort: "created_at",
  order: "desc",
  page: 1,
  limit: 10,
  all: false,
};

function buildQueryString(query: TaskQuery): string {
  const params = new URLSearchParams();
  if (query.status) params.set("status", query.status);
  if (query.q) params.set("q", query.q);
  params.set("sort", query.sort);
  params.set("order", query.order);
  params.set("page", String(query.page));
  params.set("limit", String(query.limit));
  if (query.all) params.set("all", "true");
  return params.toString();
}

interface UseTasksResult {
  tasks: Task[];
  meta: ListMeta | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
  createTask: (input: TaskInput) => Promise<Task>;
  updateTask: (id: string, input: TaskInput) => Promise<Task>;
  /** Optimistically toggles done status; rolls back on failure. */
  toggleComplete: (task: Task) => Promise<void>;
  /** Optimistically removes the task; rolls back on failure. */
  deleteTask: (task: Task) => Promise<void>;
}

export function useTasks(query: TaskQuery): UseTasksResult {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [meta, setMeta] = useState<ListMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const queryString = buildQueryString(query);

  // Latest tasks snapshot for optimistic rollbacks.
  const tasksRef = useRef<Task[]>([]);
  useEffect(() => {
    tasksRef.current = tasks;
  }, [tasks]);

  const refetch = useCallback(() => setRefreshKey((k) => k + 1), []);

  useEffect(() => {
    const controller = new AbortController();
    let cancelled = false;

    async function load() {
      setLoading(true);
      try {
        const res = await apiFetch<TaskListResponse>(
          `/api/tasks?${queryString}`,
          { signal: controller.signal },
        );
        if (cancelled) return;
        setTasks(res.data);
        setMeta(res.meta);
        setError(null);
      } catch (err) {
        if (cancelled || controller.signal.aborted) return;
        setError(err instanceof Error ? err.message : "Failed to load tasks.");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [queryString, refreshKey]);

  // Live updates: re-fetch the current page whenever the server reports a
  // task change (covers changes made in other tabs / by admins / etc).
  useEffect(() => {
    const token = getToken();
    if (!token) return;

    const source = new EventSource(
      `${API_URL}/api/events?access_token=${encodeURIComponent(token)}`,
    );
    source.onmessage = () => setRefreshKey((k) => k + 1);
    return () => source.close();
  }, []);

  const createTask = useCallback(
    async (input: TaskInput) => {
      const res = await apiFetch<{ data: Task }>("/api/tasks", {
        method: "POST",
        body: input,
      });
      refetch();
      return res.data;
    },
    [refetch],
  );

  const updateTask = useCallback(
    async (id: string, input: TaskInput) => {
      const res = await apiFetch<{ data: Task }>(`/api/tasks/${id}`, {
        method: "PATCH",
        body: input,
      });
      refetch();
      return res.data;
    },
    [refetch],
  );

  const toggleComplete = useCallback(async (task: Task) => {
    const nextStatus = task.status === "done" ? "todo" : "done";
    setTasks((prev) =>
      prev.map((t) => (t.id === task.id ? { ...t, status: nextStatus } : t)),
    );
    try {
      await apiFetch<{ data: Task }>(`/api/tasks/${task.id}`, {
        method: "PATCH",
        body: { status: nextStatus },
      });
    } catch (err) {
      // Roll back the optimistic change.
      setTasks((prev) =>
        prev.map((t) =>
          t.id === task.id ? { ...t, status: task.status } : t,
        ),
      );
      throw err;
    }
  }, []);

  const deleteTask = useCallback(async (task: Task) => {
    const snapshot = tasksRef.current;
    setTasks((prev) => prev.filter((t) => t.id !== task.id));
    try {
      await apiFetch<void>(`/api/tasks/${task.id}`, { method: "DELETE" });
      setRefreshKey((k) => k + 1);
    } catch (err) {
      setTasks(snapshot);
      throw err;
    }
  }, []);

  return {
    tasks,
    meta,
    loading,
    error,
    refetch,
    createTask,
    updateTask,
    toggleComplete,
    deleteTask,
  };
}
