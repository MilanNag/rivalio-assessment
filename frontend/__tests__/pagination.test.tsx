import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Pagination } from "@/components/pagination";

describe("Pagination", () => {
  it("renders nothing when there is a single page", () => {
    const { container } = render(
      <Pagination
        meta={{ page: 1, limit: 10, total: 5, totalPages: 1 }}
        onPageChange={jest.fn()}
      />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("shows page summary and navigates", async () => {
    const onPageChange = jest.fn();
    render(
      <Pagination
        meta={{ page: 2, limit: 10, total: 35, totalPages: 4 }}
        onPageChange={onPageChange}
      />,
    );
    expect(screen.getByText(/page 2 of 4/i)).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: /previous/i }));
    expect(onPageChange).toHaveBeenCalledWith(1);
    await userEvent.click(screen.getByRole("button", { name: /next/i }));
    expect(onPageChange).toHaveBeenCalledWith(3);
  });

  it("disables previous on first page and next on last page", () => {
    const { rerender } = render(
      <Pagination
        meta={{ page: 1, limit: 10, total: 35, totalPages: 4 }}
        onPageChange={jest.fn()}
      />,
    );
    expect(screen.getByRole("button", { name: /previous/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /next/i })).toBeEnabled();

    rerender(
      <Pagination
        meta={{ page: 4, limit: 10, total: 35, totalPages: 4 }}
        onPageChange={jest.fn()}
      />,
    );
    expect(screen.getByRole("button", { name: /previous/i })).toBeEnabled();
    expect(screen.getByRole("button", { name: /next/i })).toBeDisabled();
  });
});
