export type TaskStatus = "todo" | "in_progress" | "done";
export type TaskPriority = "low" | "medium" | "high";
export type UserRole = "user" | "admin";

export interface User {
  id: string;
  email: string;
  role: UserRole;
  createdAt: string;
}

export interface Task {
  id: string;
  userId: string;
  userEmail?: string;
  title: string;
  description: string;
  status: TaskStatus;
  priority: TaskPriority;
  dueDate: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface Attachment {
  id: string;
  taskId: string;
  fileName: string;
  contentType: string;
  sizeBytes: number;
  createdAt: string;
}

export interface ActivityEntry {
  id: number;
  taskId: string;
  userId: string;
  userEmail: string;
  action: string;
  detail: string;
  createdAt: string;
}

export interface ListMeta {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface TaskListResponse {
  data: Task[];
  meta: ListMeta;
}

export interface ApiErrorBody {
  error: {
    code: string;
    message: string;
    fields?: Record<string, string>;
  };
}

export interface TaskQuery {
  status: TaskStatus | "";
  q: string;
  sort: "created_at" | "due_date" | "priority";
  order: "asc" | "desc";
  page: number;
  limit: number;
  all: boolean;
}

export interface TaskInput {
  title?: string;
  description?: string;
  status?: TaskStatus;
  priority?: TaskPriority;
  dueDate?: string;
}
