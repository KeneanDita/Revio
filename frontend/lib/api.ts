import axios, { AxiosError } from "axios";
import type {
  AuthResponse,
  PRListResponse,
  PRDetailResponse,
  RepoListResponse,
  Repository,
  AnalyticsResponse,
  NotificationListResponse,
  PRFilters,
  AnalyticsFilters,
} from "@/types";
import { buildQueryString } from "./utils";

const api = axios.create({
  baseURL: "/api",
  withCredentials: true,
  headers: { "Content-Type": "application/json" },
});

api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("access_token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

api.interceptors.response.use(
  (res) => res,
  (error: AxiosError) => {
    if (error.response?.status === 401 && typeof window !== "undefined") {
      localStorage.removeItem("access_token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);

// ─── Auth ─────────────────────────────────────────────────────────────────────

export const authApi = {
  signup: (email: string, password: string, name: string) =>
    api.post<AuthResponse>("/auth/signup", { email, password, name }),

  login: (email: string, password: string) =>
    api.post<AuthResponse>("/auth/login", { email, password }),

  logout: () => api.post("/auth/logout"),

  me: () => api.get<AuthResponse["user"]>("/auth/me"),

  githubLoginUrl: () => `/api/auth/github`,
};

// ─── Repositories ─────────────────────────────────────────────────────────────

export const reposApi = {
  list: () => api.get<RepoListResponse>("/repos"),

  listGitHub: () => api.get<RepoListResponse>("/repos/github"),

  connect: (fullName: string) =>
    api.post<Repository>("/repos/connect", { full_name: fullName }),

  disconnect: (id: string) => api.delete(`/repos/${id}`),

  sync: (id: string) => api.post(`/repos/${id}/sync`),
};

// ─── Pull Requests ────────────────────────────────────────────────────────────

export const prsApi = {
  list: (filters: PRFilters = {}) =>
    api.get<PRListResponse>(`/prs${buildQueryString(filters as Record<string, string | number | boolean | undefined | null>)}`),

  get: (id: string) => api.get<PRDetailResponse>(`/prs/${id}`),

  comment: (id: string, body: string) =>
    api.post(`/prs/${id}/comment`, { body }),
};

// ─── Analytics ────────────────────────────────────────────────────────────────

export const analyticsApi = {
  get: (filters: AnalyticsFilters = {}) =>
    api.get<AnalyticsResponse>(
      `/analytics${buildQueryString(filters as Record<string, string | number | boolean | undefined | null>)}`
    ),
};

// ─── Notifications ────────────────────────────────────────────────────────────

export const notificationsApi = {
  list: () => api.get<NotificationListResponse>("/notifications"),

  markRead: (id: string) => api.patch(`/notifications/${id}/read`),

  markAllRead: () => api.patch("/notifications/read-all"),
};

export default api;
