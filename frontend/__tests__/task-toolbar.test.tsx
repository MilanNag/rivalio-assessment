import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TaskToolbar } from "@/components/task-toolbar";
import { DEFAULT_QUERY } from "@/lib/use-tasks";

describe("TaskToolbar", () => {
  it("emits search changes resetting to page 1", async () => {
    const onChange = jest.fn();
    render(
      <TaskToolbar query={DEFAULT_QUERY} isAdmin={false} onChange={onChange} />,
    );
    await userEvent.type(screen.getByLabelText(/search/i), "a");
    expect(onChange).toHaveBeenCalledWith({ q: "a", page: 1 });
  });

  it("emits status filter changes", async () => {
    const onChange = jest.fn();
    render(
      <TaskToolbar query={DEFAULT_QUERY} isAdmin={false} onChange={onChange} />,
    );
    await userEvent.selectOptions(
      screen.getByLabelText(/filter by status/i),
      "done",
    );
    expect(onChange).toHaveBeenCalledWith({ status: "done", page: 1 });
  });

  it("emits sort changes", async () => {
    const onChange = jest.fn();
    render(
      <TaskToolbar query={DEFAULT_QUERY} isAdmin={false} onChange={onChange} />,
    );
    await userEvent.selectOptions(screen.getByLabelText(/sort by/i), "priority");
    expect(onChange).toHaveBeenCalledWith({ sort: "priority", page: 1 });
  });

  it("toggles sort order", async () => {
    const onChange = jest.fn();
    render(
      <TaskToolbar query={DEFAULT_QUERY} isAdmin={false} onChange={onChange} />,
    );
    await userEvent.click(screen.getByRole("button", { name: /descending/i }));
    expect(onChange).toHaveBeenCalledWith({ order: "asc", page: 1 });
  });

  it("hides the all-users toggle for non-admins and shows it for admins", () => {
    const { rerender } = render(
      <TaskToolbar
        query={DEFAULT_QUERY}
        isAdmin={false}
        onChange={jest.fn()}
      />,
    );
    expect(screen.queryByText(/all users/i)).not.toBeInTheDocument();

    rerender(
      <TaskToolbar query={DEFAULT_QUERY} isAdmin={true} onChange={jest.fn()} />,
    );
    expect(screen.getByText(/all users/i)).toBeInTheDocument();
  });

  it("emits the all-users toggle", async () => {
    const onChange = jest.fn();
    render(
      <TaskToolbar query={DEFAULT_QUERY} isAdmin={true} onChange={onChange} />,
    );
    await userEvent.click(screen.getByRole("checkbox"));
    expect(onChange).toHaveBeenCalledWith({ all: true, page: 1 });
  });
});
