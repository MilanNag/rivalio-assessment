export function Spinner({ label }: { label?: string }) {
  return (
    <div
      className="flex items-center gap-2 text-zinc-500 dark:text-zinc-400"
      role="status"
    >
      <span className="size-5 animate-spin rounded-full border-2 border-zinc-300 border-t-indigo-600 dark:border-zinc-700 dark:border-t-indigo-400" />
      {label && <span className="text-sm">{label}</span>}
    </div>
  );
}
