"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { API_URL, apiFetch, getToken } from "@/lib/api";
import type { ActivityEntry, Attachment, Task } from "@/lib/types";
import { Spinner } from "./spinner";

interface TaskDetailModalProps {
  task: Task;
  onClose: () => void;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

export function TaskDetailModal({ task, onClose }: TaskDetailModalProps) {
  const [attachments, setAttachments] = useState<Attachment[] | null>(null);
  const [activity, setActivity] = useState<ActivityEntry[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const reload = useCallback(() => setRefreshKey((k) => k + 1), []);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const [att, act] = await Promise.all([
          apiFetch<{ data: Attachment[] }>(
            `/api/tasks/${task.id}/attachments`,
          ),
          apiFetch<{ data: ActivityEntry[] }>(`/api/tasks/${task.id}/activity`),
        ]);
        if (cancelled) return;
        setAttachments(att.data);
        setActivity(act.data);
        setError(null);
      } catch {
        if (!cancelled) setError("Failed to load task details.");
      }
    }
    load();
    return () => {
      cancelled = true;
    };
  }, [task.id, refreshKey]);

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onClose]);

  async function handleUpload(file: File) {
    setUploading(true);
    setError(null);
    const formData = new FormData();
    formData.append("file", file);
    try {
      await apiFetch(`/api/tasks/${task.id}/attachments`, {
        method: "POST",
        formData,
      });
      reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed.");
    } finally {
      setUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  async function handleDeleteAttachment(id: string) {
    try {
      await apiFetch<void>(`/api/attachments/${id}`, { method: "DELETE" });
      reload();
    } catch {
      setError("Failed to delete attachment.");
    }
  }

  function downloadUrl(id: string): string {
    const token = getToken() ?? "";
    return `${API_URL}/api/attachments/${id}/download?access_token=${encodeURIComponent(token)}`;
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/40 p-0 backdrop-blur-sm sm:items-center sm:p-4"
      onClick={onClose}
      role="presentation"
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label={`Details for ${task.title}`}
        onClick={(e) => e.stopPropagation()}
        className="max-h-[90vh] w-full overflow-y-auto rounded-t-2xl bg-white p-6 shadow-xl sm:max-w-xl sm:rounded-2xl dark:bg-zinc-900"
      >
        <div className="flex items-start justify-between gap-4">
          <h2 className="text-lg font-semibold">{task.title}</h2>
          <button
            type="button"
            onClick={onClose}
            aria-label="Close"
            className="rounded-lg p-1 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-800"
          >
            ✕
          </button>
        </div>

        {task.description && (
          <p className="mt-2 text-sm whitespace-pre-wrap text-zinc-600 dark:text-zinc-300">
            {task.description}
          </p>
        )}

        {error && (
          <p
            role="alert"
            className="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950/50 dark:text-red-300"
          >
            {error}
          </p>
        )}

        <section className="mt-6">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold tracking-wide text-zinc-500 uppercase dark:text-zinc-400">
              Attachments
            </h3>
            <label className="cursor-pointer rounded-lg border border-zinc-300 px-3 py-1.5 text-sm transition hover:bg-zinc-100 dark:border-zinc-700 dark:hover:bg-zinc-800">
              {uploading ? "Uploading…" : "Upload file"}
              <input
                ref={fileInputRef}
                type="file"
                className="hidden"
                disabled={uploading}
                accept="image/*,.pdf,.txt,.csv,.doc,.docx,.xls,.xlsx"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) handleUpload(file);
                }}
              />
            </label>
          </div>

          {attachments === null ? (
            <div className="mt-3">
              <Spinner label="Loading attachments…" />
            </div>
          ) : attachments.length === 0 ? (
            <p className="mt-3 text-sm text-zinc-500 dark:text-zinc-400">
              No attachments yet.
            </p>
          ) : (
            <ul className="mt-3 space-y-2">
              {attachments.map((a) => (
                <li
                  key={a.id}
                  className="flex items-center justify-between gap-3 rounded-lg border border-zinc-200 px-3 py-2 text-sm dark:border-zinc-800"
                >
                  <a
                    href={downloadUrl(a.id)}
                    className="min-w-0 flex-1 truncate font-medium text-indigo-600 hover:underline dark:text-indigo-400"
                    download={a.fileName}
                  >
                    {a.fileName}
                  </a>
                  <span className="shrink-0 text-xs text-zinc-400">
                    {formatBytes(a.sizeBytes)}
                  </span>
                  <button
                    type="button"
                    onClick={() => handleDeleteAttachment(a.id)}
                    aria-label={`Delete ${a.fileName}`}
                    className="shrink-0 rounded p-1 text-zinc-400 hover:text-red-600 dark:hover:text-red-400"
                  >
                    🗑️
                  </button>
                </li>
              ))}
            </ul>
          )}
        </section>

        <section className="mt-6">
          <h3 className="text-sm font-semibold tracking-wide text-zinc-500 uppercase dark:text-zinc-400">
            Activity
          </h3>
          {activity === null ? (
            <div className="mt-3">
              <Spinner label="Loading activity…" />
            </div>
          ) : activity.length === 0 ? (
            <p className="mt-3 text-sm text-zinc-500 dark:text-zinc-400">
              No activity recorded yet.
            </p>
          ) : (
            <ol className="mt-3 space-y-3 border-l border-zinc-200 pl-4 dark:border-zinc-800">
              {activity.map((entry) => (
                <li key={entry.id} className="text-sm">
                  <p className="text-zinc-700 dark:text-zinc-200">
                    {entry.detail || entry.action}
                  </p>
                  <p className="mt-0.5 text-xs text-zinc-400">
                    {entry.userEmail} · {formatDateTime(entry.createdAt)}
                  </p>
                </li>
              ))}
            </ol>
          )}
        </section>
      </div>
    </div>
  );
}
