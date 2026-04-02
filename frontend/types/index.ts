export type Role = "user" | "admin";

export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string;
  role: Role;
  github_login: string;
  created_at: string;
}

export interface Repository {
  id: string;
  user_id: string;
  github_id: number;
  name: string;
  full_name: string;
  owner: string;
  description: string;
  private: boolean;
  html_url: string;
  last_sync_at: string | null;
  created_at: string;
  updated_at: string;
}

export type PRStatus = "open" | "closed" | "merged";

export interface PullRequest {
  id: string;
  repo_id: string;
  github_id: number;
  number: number;
  title: string;
  body: string;
  author: string;
  author_avatar_url: string;
  status: PRStatus;
  base_branch: string;
  head_branch: string;
  additions: number;
  deletions: number;
  changed_files: number;
  commit_count: number;
  comment_count: number;
  review_count: number;
  html_url: string;
  merged_at: string | null;
  closed_at: string | null;
  first_review_at: string | null;
  github_created_at: string;
  github_updated_at: string;
  created_at: string;
  updated_at: string;
  Repo?: Repository;
  reviews?: Review[];
}

export type ReviewState =
  | "approved"
  | "changes_requested"
  | "commented"
  | "dismissed"
  | "APPROVED"
  | "CHANGES_REQUESTED"
  | "COMMENTED"
  | "DISMISSED";

export interface Review {
  id: string;
  pr_id: string;
  github_id: number;
  reviewer: string;
  reviewer_avatar: string;
  state: ReviewState;
  body: string;
  html_url: string;
  submitted_at: string;
}

export interface Notification {
  id: string;
  user_id: string;
  type: string;
  title: string;
  body: string;
  link: string;
  read: boolean;
  created_at: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: number;
}

export interface AuthResponse {
  user: User;
  tokens: TokenPair;
}

// ─── API Response wrappers ────────────────────────────────────────────────────

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface PRListResponse {
  pull_requests: PullRequest[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface RepoListResponse {
  repositories: Repository[];
  total: number;
}

export interface PRDetailResponse {
  pull_request: PullRequest;
  time_to_merge_hours: number | null;
  time_to_review_hours: number | null;
}

// ─── Analytics ───────────────────────────────────────────────────────────────

export interface SummaryMetrics {
  total_prs: number;
  open_prs: number;
  merged_prs: number;
  closed_prs: number;
  avg_merge_time_hours: number | null;
  avg_review_time_hours: number | null;
  merge_rate: number;
}

export interface DailyDataPoint {
  date: string;
  opened: number;
  merged: number;
  closed: number;
}

export interface ReviewerMetric {
  reviewer: string;
  reviewer_avatar: string;
  review_count: number;
  approval_count: number;
}

export interface AuthorMetric {
  author: string;
  author_avatar: string;
  pr_count: number;
  merged_count: number;
}

export interface StatusBreakdown {
  open: number;
  merged: number;
  closed: number;
}

export interface AnalyticsResponse {
  summary: SummaryMetrics;
  daily_trend: DailyDataPoint[];
  top_reviewers: ReviewerMetric[];
  top_authors: AuthorMetric[];
  status_breakdown: StatusBreakdown;
}

export interface NotificationListResponse {
  notifications: Notification[];
  unread_count: number;
}

// ─── Filters ─────────────────────────────────────────────────────────────────

export interface PRFilters {
  status?: PRStatus | "";
  repo_id?: string;
  author?: string;
  from?: string;
  to?: string;
  page?: number;
  per_page?: number;
}

export interface AnalyticsFilters {
  from?: string;
  to?: string;
  repo_id?: string;
  author?: string;
}
