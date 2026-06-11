"use client";

import type { TaskQuery } from "@/lib/types";

interface TaskToolbarProps {
  query: TaskQuery;
  isAdmin: boolean;
  onChange: (patch: Partial<TaskQuery>) => void;
}

const selectClass =
  "rounded-lg border border-zinc-300 bg-white px-2.5 py-2 text-sm outline-none focus:border-indigo-500 dark:border-zinc-700 dark:bg-zinc-900";

export function TaskToolbar({ query, isAdmin, onChange }: TaskToolbarProps) {
  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center">
      <input
        type="search"
        value={query.q}
        onChange={(e) => onChange({ q: e.target.value, page: 1 })}
        placeholder="Search by title…"
        aria-label="Search tasks by title"
        className="w-full flex-1 rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 sm:min-w-48 dark:border-zinc-700 dark:bg-zinc-900"
      />

      <div className="flex flex-wrap items-center gap-2">
        <select
          value={query.status}
          onChange={(e) =>
            onChange({ status: e.target.value as TaskQuery["status"], page: 1 })
          }
          aria-label="Filter by status"
          className={selectClass}
        >
          <option value="">All statuses</option>
          <option value="todo">To do</option>
          <option value="in_progress">In progress</option>
          <option value="done">Done</option>
        </select>

        <select
          value={query.sort}
          onChange={(e) =>
            onChange({ sort: e.target.value as TaskQuery["sort"], page: 1 })
          }
          aria-label="Sort by"
          className={selectClass}
        >
          <option value="created_at">Created date</option>
          <option value="due_date">Due date</option>
          <option value="priority">Priority</option>
        </select>

        <button
          type="button"
          onClick={() =>
            onChange({ order: query.order === "asc" ? "desc" : "asc", page: 1 })
          }
          aria-label={`Sort ${query.order === "asc" ? "ascending" : "descending"}`}
          title={query.order === "asc" ? "Ascending" : "Descending"}
          className="rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm transition hover:bg-zinc-100 dark:border-zinc-700 dark:bg-zinc-900 dark:hover:bg-zinc-800"
        >
          {query.order === "asc" ? "↑" : "↓"}
        </button>

        {isAdmin && (
          <label className="flex cursor-pointer items-center gap-1.5 rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm dark:border-zinc-700 dark:bg-zinc-900">
            <input
              type="checkbox"
              checked={query.all}
              onChange={(e) => onChange({ all: e.target.checked, page: 1 })}
              className="accent-indigo-600"
            />
            All users
          </label>
        )}
      </div>
    </div>
  );
}
