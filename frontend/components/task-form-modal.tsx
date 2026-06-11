"use client";

import { useEffect, useState, type FormEvent } from "react";
import { ApiError } from "@/lib/api";
import type { Task, TaskInput } from "@/lib/types";
import { validateTaskForm, type TaskFormValues } from "@/lib/validation";

interface TaskFormModalProps {
  /** Task being edited; null means "create new". */
  task: Task | null;
  onClose: () => void;
  onSubmit: (input: TaskInput) => Promise<void>;
}

function toDateInputValue(iso: string | null): string {
  if (!iso) return "";
  return new Date(iso).toISOString().slice(0, 10);
}

export function TaskFormModal({ task, onClose, onSubmit }: TaskFormModalProps) {
  const [values, setValues] = useState<TaskFormValues>({
    title: task?.title ?? "",
    description: task?.description ?? "",
    status: task?.status ?? "todo",
    priority: task?.priority ?? "medium",
    dueDate: toDateInputValue(task?.dueDate ?? null),
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onClose]);

  function set<K extends keyof TaskFormValues>(key: K, value: string) {
    setValues((prev) => ({ ...prev, [key]: value }));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setFormError(null);

    const validationErrors = validateTaskForm(values);
    setErrors(validationErrors);
    if (Object.keys(validationErrors).length > 0) return;

    setSubmitting(true);
    try {
      await onSubmit({
        title: values.title.trim(),
        description: values.description,
        status: values.status as TaskInput["status"],
        priority: values.priority as TaskInput["priority"],
        dueDate: values.dueDate,
      });
      onClose();
    } catch (err) {
      if (err instanceof ApiError) {
        setErrors(err.fields);
        setFormError(Object.keys(err.fields).length ? null : err.message);
      } else {
        setFormError("Unable to save the task. Please try again.");
      }
      setSubmitting(false);
    }
  }

  const inputClass =
    "mt-1 w-full rounded-lg border border-zinc-300 bg-transparent px-3 py-2 text-sm outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 dark:border-zinc-700 dark:[color-scheme:dark]";

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-0 backdrop-blur-sm sm:items-center sm:p-4"
      onClick={onClose}
      role="presentation"
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label={task ? "Edit task" : "New task"}
        onClick={(e) => e.stopPropagation()}
        className="max-h-[90vh] w-full overflow-y-auto rounded-t-2xl bg-white p-6 shadow-xl sm:max-w-lg sm:rounded-2xl dark:bg-zinc-900"
      >
        <h2 className="text-lg font-semibold">
          {task ? "Edit task" : "New task"}
        </h2>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4" noValidate>
          <div>
            <label htmlFor="task-title" className="block text-sm font-medium">
              Title <span className="text-red-500">*</span>
            </label>
            <input
              id="task-title"
              value={values.title}
              onChange={(e) => set("title", e.target.value)}
              className={inputClass}
              placeholder="What needs to be done?"
              autoFocus
            />
            {errors.title && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                {errors.title}
              </p>
            )}
          </div>

          <div>
            <label
              htmlFor="task-description"
              className="block text-sm font-medium"
            >
              Description
            </label>
            <textarea
              id="task-description"
              value={values.description}
              onChange={(e) => set("description", e.target.value)}
              rows={3}
              className={inputClass}
              placeholder="Optional details…"
            />
            {errors.description && (
              <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                {errors.description}
              </p>
            )}
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <div>
              <label
                htmlFor="task-status"
                className="block text-sm font-medium"
              >
                Status
              </label>
              <select
                id="task-status"
                value={values.status}
                onChange={(e) => set("status", e.target.value)}
                className={inputClass}
              >
                <option value="todo">To do</option>
                <option value="in_progress">In progress</option>
                <option value="done">Done</option>
              </select>
            </div>

            <div>
              <label
                htmlFor="task-priority"
                className="block text-sm font-medium"
              >
                Priority
              </label>
              <select
                id="task-priority"
                value={values.priority}
                onChange={(e) => set("priority", e.target.value)}
                className={inputClass}
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </div>

            <div>
              <label htmlFor="task-due" className="block text-sm font-medium">
                Due date
              </label>
              <input
                id="task-due"
                type="date"
                value={values.dueDate}
                onChange={(e) => set("dueDate", e.target.value)}
                className={inputClass}
              />
              {errors.dueDate && (
                <p className="mt-1 text-sm text-red-600 dark:text-red-400">
                  {errors.dueDate}
                </p>
              )}
            </div>
          </div>

          {formError && (
            <p
              role="alert"
              className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950/50 dark:text-red-300"
            >
              {formError}
            </p>
          )}

          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg border border-zinc-300 px-4 py-2 text-sm font-medium transition hover:bg-zinc-100 dark:border-zinc-700 dark:hover:bg-zinc-800"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-indigo-500 disabled:opacity-60"
            >
              {submitting ? "Saving…" : task ? "Save changes" : "Create task"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
