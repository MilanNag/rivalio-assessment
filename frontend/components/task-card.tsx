"use client";

import type { Task } from "@/lib/types";

const priorityStyles: Record<Task["priority"], string> = {
  high: "bg-red-100 text-red-700 dark:bg-red-950/60 dark:text-red-300",
  medium:
    "bg-amber-100 text-amber-700 dark:bg-amber-950/60 dark:text-amber-300",
  low: "bg-emerald-100 text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-300",
};

const statusLabels: Record<Task["status"], string> = {
  todo: "To do",
  in_progress: "In progress",
  done: "Done",
};

function formatDate(iso: string | null): string | null {
  if (!iso) return null;
  return new Date(iso).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

interface TaskCardProps {
  task: Task;
  showOwner?: boolean;
  onToggleComplete: (task: Task) => void;
  onEdit: (task: Task) => void;
  onDelete: (task: Task) => void;
  onOpen: (task: Task) => void;
}

export function TaskCard({
  task,
  showOwner,
  onToggleComplete,
  onEdit,
  onDelete,
  onOpen,
}: TaskCardProps) {
  const done = task.status === "done";
  const due = formatDate(task.dueDate);
  const overdue =
    !done && task.dueDate !== null && new Date(task.dueDate) < new Date();

  return (
    <li className="group flex items-start gap-3 rounded-xl border border-zinc-200 bg-white p-4 shadow-sm transition hover:shadow-md dark:border-zinc-800 dark:bg-zinc-900">
      <input
        type="checkbox"
        checked={done}
        onChange={() => onToggleComplete(task)}
        aria-label={done ? "Mark as not done" : "Mark as done"}
        className="mt-1 size-4 shrink-0 accent-indigo-600"
      />

      <button
        type="button"
        onClick={() => onOpen(task)}
        className="min-w-0 flex-1 text-left"
      >
        <p
          className={`truncate font-medium ${
            done ? "text-zinc-400 line-through dark:text-zinc-500" : ""
          }`}
        >
          {task.title}
        </p>
        {task.description && (
          <p className="mt-0.5 line-clamp-2 text-sm text-zinc-500 dark:text-zinc-400">
            {task.description}
          </p>
        )}
        <div className="mt-2 flex flex-wrap items-center gap-2 text-xs">
          <span
            className={`rounded-full px-2 py-0.5 font-medium ${priorityStyles[task.priority]}`}
          >
            {task.priority}
          </span>
          <span className="rounded-full bg-zinc-100 px-2 py-0.5 font-medium text-zinc-600 dark:bg-zinc-800 dark:text-zinc-300">
            {statusLabels[task.status]}
          </span>
          {due && (
            <span
              className={
                overdue
                  ? "font-medium text-red-600 dark:text-red-400"
                  : "text-zinc-500 dark:text-zinc-400"
              }
            >
              Due {due}
              {overdue ? " (overdue)" : ""}
            </span>
          )}
          {showOwner && task.userEmail && (
            <span className="text-zinc-400 dark:text-zinc-500">
              {task.userEmail}
            </span>
          )}
        </div>
      </button>

      <div className="flex shrink-0 gap-1 opacity-100 transition sm:opacity-0 sm:group-hover:opacity-100">
        <button
          type="button"
          onClick={() => onEdit(task)}
          aria-label={`Edit ${task.title}`}
          className="rounded-lg p-1.5 text-zinc-400 transition hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-800 dark:hover:text-zinc-200"
        >
          ✏️
        </button>
        <button
          type="button"
          onClick={() => onDelete(task)}
          aria-label={`Delete ${task.title}`}
          className="rounded-lg p-1.5 text-zinc-400 transition hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/40 dark:hover:text-red-400"
        >
          🗑️
        </button>
      </div>
    </li>
  );
}
