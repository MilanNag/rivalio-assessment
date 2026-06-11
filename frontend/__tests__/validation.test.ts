import {
  validateCredentials,
  validateTaskForm,
  TITLE_MAX,
  DESCRIPTION_MAX,
} from "@/lib/validation";

describe("validateTaskForm", () => {
  const valid = {
    title: "Write tests",
    description: "Cover validation",
    status: "todo",
    priority: "medium",
    dueDate: "2026-07-01",
  };

  it("accepts a valid form", () => {
    expect(validateTaskForm(valid)).toEqual({});
  });

  it("requires a title", () => {
    expect(validateTaskForm({ ...valid, title: "" })).toHaveProperty("title");
    expect(validateTaskForm({ ...valid, title: "   " })).toHaveProperty(
      "title",
    );
  });

  it("rejects overlong title", () => {
    const result = validateTaskForm({
      ...valid,
      title: "x".repeat(TITLE_MAX + 1),
    });
    expect(result).toHaveProperty("title");
  });

  it("rejects overlong description", () => {
    const result = validateTaskForm({
      ...valid,
      description: "d".repeat(DESCRIPTION_MAX + 1),
    });
    expect(result).toHaveProperty("description");
  });

  it("rejects an unparseable due date", () => {
    expect(validateTaskForm({ ...valid, dueDate: "not-a-date" })).toHaveProperty(
      "dueDate",
    );
  });

  it("allows an empty due date", () => {
    expect(validateTaskForm({ ...valid, dueDate: "" })).toEqual({});
  });
});

describe("validateCredentials", () => {
  it("accepts valid credentials", () => {
    expect(validateCredentials("a@b.com", "password123")).toEqual({});
  });

  it("requires email", () => {
    expect(validateCredentials("", "password123")).toHaveProperty("email");
  });

  it("rejects malformed email", () => {
    expect(validateCredentials("nope", "password123")).toHaveProperty("email");
  });

  it("requires password of at least 8 characters", () => {
    expect(validateCredentials("a@b.com", "")).toHaveProperty("password");
    expect(validateCredentials("a@b.com", "1234567")).toHaveProperty(
      "password",
    );
  });
});
