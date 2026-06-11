import { act, renderHook, waitFor } from "@testing-library/react";
import { AuthProvider, useAuth } from "@/lib/auth-context";
import { setToken, getToken } from "@/lib/api";
import type { ReactNode } from "react";

const fetchMock = jest.fn();
global.fetch = fetchMock as unknown as typeof fetch;

const wrapper = ({ children }: { children: ReactNode }) => (
  <AuthProvider>{children}</AuthProvider>
);

const user = {
  id: "u1",
  email: "alice@example.com",
  role: "user",
  createdAt: "2026-06-01T00:00:00Z",
};

function jsonResponse(status: number, body: unknown) {
  return { ok: status < 300, status, json: async () => body };
}

beforeEach(() => {
  fetchMock.mockReset();
  window.localStorage.clear();
});

describe("AuthProvider", () => {
  it("starts logged out when no token is stored", async () => {
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.initializing).toBe(false));
    expect(result.current.user).toBeNull();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("restores the session from a stored token", async () => {
    setToken("stored-token");
    fetchMock.mockResolvedValue(jsonResponse(200, { user }));

    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.initializing).toBe(false));
    expect(result.current.user?.email).toBe("alice@example.com");
  });

  it("clears an invalid stored token", async () => {
    setToken("expired-token");
    fetchMock.mockResolvedValue(
      jsonResponse(401, {
        error: { code: "unauthorized", message: "Invalid or expired token." },
      }),
    );

    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.initializing).toBe(false));
    expect(result.current.user).toBeNull();
    expect(getToken()).toBeNull();
  });

  it("login stores the token and user", async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(200, { token: "new-token", user }),
    );

    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.initializing).toBe(false));

    await act(async () => {
      await result.current.login("alice@example.com", "password123");
    });
    expect(result.current.user?.id).toBe("u1");
    expect(getToken()).toBe("new-token");
  });

  it("logout clears the session", async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(200, { token: "new-token", user }),
    );
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.initializing).toBe(false));

    await act(async () => {
      await result.current.login("alice@example.com", "password123");
    });
    act(() => result.current.logout());

    expect(result.current.user).toBeNull();
    expect(getToken()).toBeNull();
  });
});
