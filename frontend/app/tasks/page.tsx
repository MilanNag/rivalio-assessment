"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Pagination } from "@/components/pagination";
import { Spinner } from "@/components/spinner";
import { TaskCard } from "@/components/task-card";
import { TaskDetailModal } from "@/components/task-detail-modal";
import { TaskFormModal } from "@/components/task-form-modal";
import { TaskToolbar } from "@/components/task-toolbar";
import { ThemeToggle } from "@/components/theme-toggle";
import { useAuth } from "@/lib/auth-context";
import { DEFAULT_QUERY, useTasks } from "@/lib/use-tasks";
import type { Task, TaskInput, TaskQuery } from "@/lib/types";

export default function TasksPage() {
  const { user, initializing, logout } = useAuth();
  const router = useRouter();

  const [query, setQuery] = useState<TaskQuery>(DEFAULT_QUERY);
  const [searchInput, setSearchInput] = useState("");
  const [formTask, setFormTask] = useState<Task | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [detailTask, setDetailTask] = useState<Task | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  const {
    tasks,
    meta,
    loading,
    error,
    refetch,
    createTask,
    updateTask,
    toggleComplete,
    deleteTask,
  } = useTasks(query);

  // Redirect unauthenticated visitors to the login page.
  useEffect(() => {
    if (!initializing && !user) router.replace("/login");
  }, [initializing, user, router]);

  // Debounce search input into the query.
  useEffect(() => {
    const handle = setTimeout(() => {
      setQuery((prev) =>
        prev.q === searchInput ? prev : { ...prev, q: searchInput, page: 1 },
      );
    }, 300);
    return () => clearTimeout(handle);
  }, [searchInput]);

  // Auto-dismiss toasts.
  useEffect(() => {
    if (!toast) return;
    const handle = setTimeout(() => setToast(null), 4000);
    return () => clearTimeout(handle);
  }, [toast]);

  if (initializing || !user) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <Spinner label="Loading…" />
      </main>
    );
  }

  function updateQuery(patch: Partial<TaskQuery>) {
    if (patch.q !== undefined) {
      setSearchInput(patch.q);
      return;
    }
    setQuery((prev) => ({ ...prev, ...patch }));
  }

  async function handleFormSubmit(input: TaskInput) {
    if (formTask) {
      await updateTask(formTask.id, input);
    } else {
      await createTask(input);
    }
  }

  async function handleToggle(task: Task) {
    try {
      await toggleComplete(task);
    } catch {
      setToast("Could not update the task. Change reverted.");
    }
  }

  async function handleDelete(task: Task) {
    try {
      await deleteTask(task);
    } catch {
      setToast("Could not delete the task. Change reverted.");
    }
  }

  const hasFilters =
    query.status !== "" || query.q !== "" || query.all !== false;

  return (
    <div className="mx-auto min-h-screen w-full max-w-3xl px-4 py-6 sm:py-10">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">TaskFlow</h1>
          <p className="text-sm text-zinc-500 dark:text-zinc-400">
            {user.email}
            {user.role === "admin" && (
              <span className="ml-2 rounded-full bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-700 dark:bg-indigo-950/60 dark:text-indigo-300">
                admin
              </span>
            )}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <button
            type="button"
            onClick={() => {
              logout();
              router.replace("/login");
            }}
            className="rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm font-medium shadow-sm transition hover:bg-zinc-100 dark:border-zinc-700 dark:bg-zinc-900 dark:hover:bg-zinc-800"
          >
            Log out
          </button>
        </div>
      </header>

      <div className="mt-6 flex flex-col gap-3">
        <div className="flex items-center justify-between gap-3">
          <h2 className="text-lg font-semibold">Your tasks</h2>
          <button
            type="button"
            onClick={() => {
              setFormTask(null);
              setFormOpen(true);
            }}
            className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm transition hover:bg-indigo-500"
          >
            + New task
          </button>
        </div>

        <TaskToolbar
          query={{ ...query, q: searchInput }}
          isAdmin={user.role === "admin"}
          onChange={updateQuery}
        />
      </div>

      <main className="mt-4">
        {error ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-center dark:border-red-900 dark:bg-red-950/40">
            <p className="font-medium text-red-700 dark:text-red-300">
              Failed to load tasks
            </p>
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">
              {error}
            </p>
            <button
              type="button"
              onClick={refetch}
              className="mt-3 rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-red-500"
            >
              Retry
            </button>
          </div>
        ) : loading && tasks.length === 0 ? (
          <div className="flex justify-center py-16">
            <Spinner label="Loading tasks…" />
          </div>
        ) : tasks.length === 0 ? (
          <div className="rounded-xl border border-dashed border-zinc-300 p-10 text-center dark:border-zinc-700">
            <p className="text-lg font-medium">
              {hasFilters ? "No tasks match" : "No tasks yet"}
            </p>
            <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
              {hasFilters
                ? "Try adjusting your search or filters."
                : "Create your first task to get started."}
            </p>
          </div>
        ) : (
          <>
            <ul className="space-y-3" aria-busy={loading}>
              {tasks.map((task) => (
                <TaskCard
                  key={task.id}
                  task={task}
                  showOwner={query.all}
                  onToggleComplete={handleToggle}
                  onEdit={(t) => {
                    setFormTask(t);
                    setFormOpen(true);
                  }}
                  onDelete={handleDelete}
                  onOpen={setDetailTask}
                />
              ))}
            </ul>
            {meta && (
              <div className="mt-4">
                <Pagination
                  meta={meta}
                  onPageChange={(page) => updateQuery({ page })}
                />
              </div>
            )}
          </>
        )}
      </main>

      {formOpen && (
        <TaskFormModal
          task={formTask}
          onClose={() => setFormOpen(false)}
          onSubmit={handleFormSubmit}
        />
      )}

      {detailTask && (
        <TaskDetailModal
          task={detailTask}
          onClose={() => setDetailTask(null)}
        />
      )}

      {toast && (
        <div
          role="alert"
          className="fixed bottom-4 left-1/2 z-50 -translate-x-1/2 rounded-lg bg-zinc-900 px-4 py-2 text-sm text-white shadow-lg dark:bg-zinc-100 dark:text-zinc-900"
        >
          {toast}
        </div>
      )}
    </div>
  );
}
