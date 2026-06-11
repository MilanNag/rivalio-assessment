import { act, renderHook, waitFor } from "@testing-library/react";
import { DEFAULT_QUERY, useTasks } from "@/lib/use-tasks";
import type { Task } from "@/lib/types";

const fetchMock = jest.fn();
global.fetch = fetchMock as unknown as typeof fetch;

// EventSource is not implemented in jsdom.
class FakeEventSource {
  onmessage: ((e: MessageEvent) => void) | null = null;
  close = jest.fn();
}
(global as Record<string, unknown>).EventSource = FakeEventSource;

function makeTask(id: string, overrides: Partial<Task> = {}): Task {
  return {
    id,
    userId: "u1",
    title: `Task ${id}`,
    description: "",
    status: "todo",
    priority: "medium",
    dueDate: null,
    createdAt: "2026-06-01T00:00:00Z",
    updatedAt: "2026-06-01T00:00:00Z",
    ...overrides,
  };
}

function jsonResponse(status: number, body: unknown) {
  return { ok: status < 300, status, json: async () => body };
}

const listResponse = (tasks: Task[]) =>
  jsonResponse(200, {
    data: tasks,
    meta: { page: 1, limit: 10, total: tasks.length, totalPages: 1 },
  });

beforeEach(() => {
  fetchMock.mockReset();
  window.localStorage.clear();
});

describe("useTasks", () => {
  it("loads tasks and exposes meta", async () => {
    fetchMock.mockResolvedValue(listResponse([makeTask("1"), makeTask("2")]));

    const { result } = renderHook(() => useTasks(DEFAULT_QUERY));
    expect(result.current.loading).toBe(true);

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.tasks).toHaveLength(2);
    expect(result.current.meta?.total).toBe(2);
    expect(result.current.error).toBeNull();
  });

  it("surfaces fetch errors", async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(500, {
        error: { code: "internal_error", message: "Something went wrong." },
      }),
    );

    const { result } = renderHook(() => useTasks(DEFAULT_QUERY));
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.error).toBeTruthy();
  });

  it("optimistically toggles status and keeps it on success", async () => {
    const task = makeTask("1");
    fetchMock.mockResolvedValue(listResponse([task]));

    const { result } = renderHook(() => useTasks(DEFAULT_QUERY));
    await waitFor(() => expect(result.current.tasks).toHaveLength(1));

    fetchMock.mockResolvedValueOnce(
      jsonResponse(200, { data: { ...task, status: "done" } }),
    );

    await act(async () => {
      await result.current.toggleComplete(result.current.tasks[0]);
    });
    expect(result.current.tasks[0].status).toBe("done");
  });

  it("rolls back the optimistic toggle on failure", async () => {
    const task = makeTask("1");
    fetchMock.mockResolvedValue(listResponse([task]));

    const { result } = renderHook(() => useTasks(DEFAULT_QUERY));
    await waitFor(() => expect(result.current.tasks).toHaveLength(1));

    fetchMock.mockResolvedValueOnce(
      jsonResponse(500, {
        error: { code: "internal_error", message: "boom" },
      }),
    );

    await act(async () => {
      await expect(
        result.current.toggleComplete(result.current.tasks[0]),
      ).rejects.toThrow();
    });
    expect(result.current.tasks[0].status).toBe("todo");
  });

  it("optimistically removes a deleted task", async () => {
    const tasks = [makeTask("1"), makeTask("2")];
    fetchMock.mockResolvedValue(listResponse(tasks));

    const { result } = renderHook(() => useTasks(DEFAULT_QUERY));
    await waitFor(() => expect(result.current.tasks).toHaveLength(2));

    fetchMock.mockResolvedValueOnce({
      ok: true,
      status: 204,
      json: async () => undefined,
    });
    // Refetch triggered after delete.
    fetchMock.mockResolvedValueOnce(listResponse([tasks[1]]));

    await act(async () => {
      await result.current.deleteTask(result.current.tasks[0]);
    });
    expect(
      result.current.tasks.find((t) => t.id === "1"),
    ).toBeUndefined();
  });

  it("restores the list when delete fails", async () => {
    const tasks = [makeTask("1"), makeTask("2")];
    fetchMock.mockResolvedValue(listResponse(tasks));

    const { result } = renderHook(() => useTasks(DEFAULT_QUERY));
    await waitFor(() => expect(result.current.tasks).toHaveLength(2));

    fetchMock.mockResolvedValueOnce(
      jsonResponse(500, { error: { code: "internal_error", message: "x" } }),
    );

    await act(async () => {
      await expect(
        result.current.deleteTask(result.current.tasks[0]),
      ).rejects.toThrow();
    });
    expect(result.current.tasks).toHaveLength(2);
  });

  it("passes filters, search, sort and pagination to the API", async () => {
    fetchMock.mockResolvedValue(listResponse([]));

    renderHook(() =>
      useTasks({
        status: "done",
        q: "report",
        sort: "due_date",
        order: "asc",
        page: 3,
        limit: 25,
        all: true,
      }),
    );

    await waitFor(() => expect(fetchMock).toHaveBeenCalled());
    const url = String(fetchMock.mock.calls[0][0]);
    expect(url).toContain("status=done");
    expect(url).toContain("q=report");
    expect(url).toContain("sort=due_date");
    expect(url).toContain("order=asc");
    expect(url).toContain("page=3");
    expect(url).toContain("limit=25");
    expect(url).toContain("all=true");
  });
});
