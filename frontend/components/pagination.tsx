"use client";

import type { ListMeta } from "@/lib/types";

interface PaginationProps {
  meta: ListMeta;
  onPageChange: (page: number) => void;
}

export function Pagination({ meta, onPageChange }: PaginationProps) {
  if (meta.totalPages <= 1) return null;

  const buttonClass =
    "rounded-lg border border-zinc-300 px-3 py-1.5 text-sm transition hover:bg-zinc-100 disabled:opacity-40 disabled:hover:bg-transparent dark:border-zinc-700 dark:hover:bg-zinc-800";

  return (
    <nav
      aria-label="Pagination"
      className="flex items-center justify-between gap-2"
    >
      <p className="text-sm text-zinc-500 dark:text-zinc-400">
        Page {meta.page} of {meta.totalPages} · {meta.total} task
        {meta.total === 1 ? "" : "s"}
      </p>
      <div className="flex gap-2">
        <button
          type="button"
          disabled={meta.page <= 1}
          onClick={() => onPageChange(meta.page - 1)}
          className={buttonClass}
        >
          Previous
        </button>
        <button
          type="button"
          disabled={meta.page >= meta.totalPages}
          onClick={() => onPageChange(meta.page + 1)}
          className={buttonClass}
        >
          Next
        </button>
      </div>
    </nav>
  );
}
