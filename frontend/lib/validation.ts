export interface TaskFormValues {
  title: string;
  description: string;
  status: string;
  priority: string;
  dueDate: string;
}

export const TITLE_MAX = 200;
export const DESCRIPTION_MAX = 5000;

/** Client-side mirror of the server's task validation rules. */
export function validateTaskForm(
  values: TaskFormValues,
): Record<string, string> {
  const errors: Record<string, string> = {};
  if (!values.title.trim()) {
    errors.title = "Title is required.";
  } else if (values.title.trim().length > TITLE_MAX) {
    errors.title = `Title must be at most ${TITLE_MAX} characters.`;
  }
  if (values.description.length > DESCRIPTION_MAX) {
    errors.description = `Description must be at most ${DESCRIPTION_MAX} characters.`;
  }
  if (values.dueDate && Number.isNaN(Date.parse(values.dueDate))) {
    errors.dueDate = "Due date is not a valid date.";
  }
  return errors;
}

export function validateCredentials(
  email: string,
  password: string,
): Record<string, string> {
  const errors: Record<string, string> = {};
  if (!email.trim()) {
    errors.email = "Email is required.";
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) {
    errors.email = "Enter a valid email address.";
  }
  if (!password) {
    errors.password = "Password is required.";
  } else if (password.length < 8) {
    errors.password = "Password must be at least 8 characters.";
  }
  return errors;
}
