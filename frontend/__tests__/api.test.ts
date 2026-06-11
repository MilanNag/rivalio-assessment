import { ApiError, apiFetch, getToken, setToken } from "@/lib/api";

const fetchMock = jest.fn();
global.fetch = fetchMock as unknown as typeof fetch;

function mockResponse(status: number, body: unknown) {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: async () => body,
  };
}

beforeEach(() => {
  fetchMock.mockReset();
  window.localStorage.clear();
});

describe("token storage", () => {
  it("persists and clears the token", () => {
    expect(getToken()).toBeNull();
    setToken("abc");
    expect(getToken()).toBe("abc");
    setToken(null);
    expect(getToken()).toBeNull();
  });
});

describe("apiFetch", () => {
  it("sends JSON bodies with auth header", async () => {
    setToken("my-token");
    fetchMock.mockResolvedValue(mockResponse(200, { data: 1 }));

    await apiFetch("/api/tasks", { method: "POST", body: { title: "x" } });

    const [url, options] = fetchMock.mock.calls[0];
    expect(url).toContain("/api/tasks");
    expect(options.method).toBe("POST");
    expect(options.headers.Authorization).toBe("Bearer my-token");
    expect(options.headers["Content-Type"]).toBe("application/json");
    expect(JSON.parse(options.body)).toEqual({ title: "x" });
  });

  it("returns parsed JSON on success", async () => {
    fetchMock.mockResolvedValue(mockResponse(200, { data: [1, 2] }));
    await expect(apiFetch("/api/tasks")).resolves.toEqual({ data: [1, 2] });
  });

  it("returns undefined for 204 responses", async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      status: 204,
      json: async () => {
        throw new Error("no body");
      },
    });
    await expect(apiFetch("/api/tasks/1")).resolves.toBeUndefined();
  });

  it("throws ApiError with envelope details on failure", async () => {
    fetchMock.mockResolvedValue(
      mockResponse(422, {
        error: {
          code: "validation_error",
          message: "One or more fields are invalid.",
          fields: { title: "Title is required." },
        },
      }),
    );

    try {
      await apiFetch("/api/tasks", { method: "POST", body: {} });
      throw new Error("expected ApiError");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      const apiErr = err as ApiError;
      expect(apiErr.status).toBe(422);
      expect(apiErr.code).toBe("validation_error");
      expect(apiErr.fields.title).toBe("Title is required.");
    }
  });

  it("handles non-JSON error bodies", async () => {
    fetchMock.mockResolvedValue({
      ok: false,
      status: 502,
      json: async () => {
        throw new Error("not json");
      },
    });

    await expect(apiFetch("/api/tasks")).rejects.toMatchObject({
      status: 502,
      code: "unknown_error",
    });
  });

  it("does not attach auth header without a token", async () => {
    fetchMock.mockResolvedValue(mockResponse(200, {}));
    await apiFetch("/api/auth/login", { method: "POST", body: {} });
    const [, options] = fetchMock.mock.calls[0];
    expect(options.headers.Authorization).toBeUndefined();
  });
});
