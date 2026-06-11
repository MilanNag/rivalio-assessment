import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ThemeProvider } from "@/lib/theme-context";
import { ThemeToggle } from "@/components/theme-toggle";

function renderToggle() {
  return render(
    <ThemeProvider>
      <ThemeToggle />
    </ThemeProvider>,
  );
}

beforeEach(() => {
  window.localStorage.clear();
  document.documentElement.classList.remove("dark");
  // jsdom lacks matchMedia.
  window.matchMedia = jest.fn().mockReturnValue({
    matches: false,
    addListener: jest.fn(),
    removeListener: jest.fn(),
  }) as unknown as typeof window.matchMedia;
});

describe("theme toggle", () => {
  it("defaults to light and toggles to dark, persisting the choice", async () => {
    renderToggle();
    expect(document.documentElement).not.toHaveClass("dark");

    await userEvent.click(
      screen.getByRole("button", { name: /switch to dark mode/i }),
    );

    expect(document.documentElement).toHaveClass("dark");
    expect(window.localStorage.getItem("taskflow.theme")).toBe("dark");
  });

  it("restores a persisted dark preference", () => {
    window.localStorage.setItem("taskflow.theme", "dark");
    renderToggle();
    expect(document.documentElement).toHaveClass("dark");
    expect(
      screen.getByRole("button", { name: /switch to light mode/i }),
    ).toBeInTheDocument();
  });

  it("toggles back to light", async () => {
    window.localStorage.setItem("taskflow.theme", "dark");
    renderToggle();
    await userEvent.click(
      screen.getByRole("button", { name: /switch to light mode/i }),
    );
    expect(document.documentElement).not.toHaveClass("dark");
    expect(window.localStorage.getItem("taskflow.theme")).toBe("light");
  });
});
