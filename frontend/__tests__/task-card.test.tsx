import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskCard } from "@/components/task-card";
import type { Task } from "@/lib/types";

const task: Task = {
  id: "t1",
  userId: "u1",
  userEmail: "owner@example.com",
  title: "Review PR",
  description: "Check the new API handlers",
  status: "todo",
  priority: "high",
  dueDate: "2026-07-01T00:00:00Z",
  createdAt: "2026-06-01T00:00:00Z",
  updatedAt: "2026-06-01T00:00:00Z",
};

function renderCard(overrides: Partial<Task> = {}, props = {}) {
  const handlers = {
    onToggleComplete: jest.fn(),
    onEdit: jest.fn(),
    onDelete: jest.fn(),
    onOpen: jest.fn(),
  };
  render(
    <ul>
      <TaskCard task={{ ...task, ...overrides }} {...handlers} {...props} />
    </ul>,
  );
  return handlers;
}

describe("TaskCard", () => {
  it("renders title, description, priority and status", () => {
    renderCard();
    expect(screen.getByText("Review PR")).toBeInTheDocument();
    expect(screen.getByText("Check the new API handlers")).toBeInTheDocument();
    expect(screen.getByText("high")).toBeInTheDocument();
    expect(screen.getByText("To do")).toBeInTheDocument();
  });

  it("checkbox reflects done state and triggers toggle", async () => {
    const handlers = renderCard({ status: "done" });
    const checkbox = screen.getByRole("checkbox");
    expect(checkbox).toBeChecked();
    await userEvent.click(checkbox);
    expect(handlers.onToggleComplete).toHaveBeenCalledTimes(1);
  });

  it("calls onDelete when delete button clicked", async () => {
    const handlers = renderCard();
    await userEvent.click(
      screen.getByRole("button", { name: /delete review pr/i }),
    );
    expect(handlers.onDelete).toHaveBeenCalledTimes(1);
  });

  it("calls onEdit when edit button clicked", async () => {
    const handlers = renderCard();
    await userEvent.click(
      screen.getByRole("button", { name: /edit review pr/i }),
    );
    expect(handlers.onEdit).toHaveBeenCalledTimes(1);
  });

  it("opens detail when title clicked", async () => {
    const handlers = renderCard();
    await userEvent.click(screen.getByText("Review PR"));
    expect(handlers.onOpen).toHaveBeenCalledTimes(1);
  });

  it("shows owner email when showOwner is set", () => {
    renderCard({}, { showOwner: true });
    expect(screen.getByText("owner@example.com")).toBeInTheDocument();
  });

  it("marks overdue tasks", () => {
    renderCard({ dueDate: "2020-01-01T00:00:00Z" });
    expect(screen.getByText(/overdue/i)).toBeInTheDocument();
  });
});
