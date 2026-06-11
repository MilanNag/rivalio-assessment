import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskFormModal } from "@/components/task-form-modal";
import type { Task } from "@/lib/types";

const baseTask: Task = {
  id: "t1",
  userId: "u1",
  title: "Existing task",
  description: "Some details",
  status: "in_progress",
  priority: "high",
  dueDate: "2026-07-01T00:00:00Z",
  createdAt: "2026-06-01T00:00:00Z",
  updatedAt: "2026-06-01T00:00:00Z",
};

describe("TaskFormModal", () => {
  it("renders create mode with defaults", () => {
    render(
      <TaskFormModal task={null} onClose={jest.fn()} onSubmit={jest.fn()} />,
    );
    expect(screen.getByRole("dialog", { name: "New task" })).toBeInTheDocument();
    expect(screen.getByLabelText(/title/i)).toHaveValue("");
    expect(screen.getByLabelText(/status/i)).toHaveValue("todo");
    expect(screen.getByLabelText(/priority/i)).toHaveValue("medium");
  });

  it("prefills fields in edit mode", () => {
    render(
      <TaskFormModal
        task={baseTask}
        onClose={jest.fn()}
        onSubmit={jest.fn()}
      />,
    );
    expect(screen.getByLabelText(/title/i)).toHaveValue("Existing task");
    expect(screen.getByLabelText(/description/i)).toHaveValue("Some details");
    expect(screen.getByLabelText(/status/i)).toHaveValue("in_progress");
    expect(screen.getByLabelText(/priority/i)).toHaveValue("high");
    expect(screen.getByLabelText(/due date/i)).toHaveValue("2026-07-01");
  });

  it("blocks submission and shows error when title is empty", async () => {
    const onSubmit = jest.fn();
    render(
      <TaskFormModal task={null} onClose={jest.fn()} onSubmit={onSubmit} />,
    );

    await userEvent.click(screen.getByRole("button", { name: /create task/i }));

    expect(await screen.findByText("Title is required.")).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("submits trimmed values and closes", async () => {
    const onSubmit = jest.fn().mockResolvedValue(undefined);
    const onClose = jest.fn();
    render(<TaskFormModal task={null} onClose={onClose} onSubmit={onSubmit} />);

    await userEvent.type(screen.getByLabelText(/title/i), "  Ship it  ");
    await userEvent.selectOptions(screen.getByLabelText(/priority/i), "high");
    await userEvent.click(screen.getByRole("button", { name: /create task/i }));

    await waitFor(() => expect(onSubmit).toHaveBeenCalledTimes(1));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Ship it", priority: "high" }),
    );
    await waitFor(() => expect(onClose).toHaveBeenCalled());
  });

  it("shows a form error when the server rejects", async () => {
    const onSubmit = jest.fn().mockRejectedValue(new Error("network down"));
    render(
      <TaskFormModal task={null} onClose={jest.fn()} onSubmit={onSubmit} />,
    );

    await userEvent.type(screen.getByLabelText(/title/i), "Task");
    await userEvent.click(screen.getByRole("button", { name: /create task/i }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      /unable to save/i,
    );
  });

  it("closes on Escape", async () => {
    const onClose = jest.fn();
    render(
      <TaskFormModal task={null} onClose={onClose} onSubmit={jest.fn()} />,
    );
    await userEvent.keyboard("{Escape}");
    expect(onClose).toHaveBeenCalled();
  });
});
