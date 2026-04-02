import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { formatDistanceToNow, format, parseISO } from "date-fns";
import type { PRStatus, ReviewState } from "@/types";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatRelativeTime(dateStr: string | null | undefined): string {
  if (!dateStr) return "—";
  try {
    return formatDistanceToNow(parseISO(dateStr), { addSuffix: true });
  } catch {
    return "—";
  }
}

export function formatDate(dateStr: string | null | undefined): string {
  if (!dateStr) return "—";
  try {
    return format(parseISO(dateStr), "MMM d, yyyy");
  } catch {
    return "—";
  }
}

export function formatDateTime(dateStr: string | null | undefined): string {
  if (!dateStr) return "—";
  try {
    return format(parseISO(dateStr), "MMM d, yyyy 'at' h:mm a");
  } catch {
    return "—";
  }
}

export function formatHours(hours: number | null | undefined): string {
  if (hours == null) return "—";
  if (hours < 1) return `${Math.round(hours * 60)}m`;
  if (hours < 24) return `${hours.toFixed(1)}h`;
  const days = Math.floor(hours / 24);
  const remaining = hours % 24;
  if (remaining < 1) return `${days}d`;
  return `${days}d ${remaining.toFixed(0)}h`;
}

export function prStatusLabel(status: PRStatus): string {
  const labels: Record<PRStatus, string> = {
    open: "Open",
    closed: "Closed",
    merged: "Merged",
  };
  return labels[status] ?? status;
}

export function reviewStateLabel(state: ReviewState): string {
  const labels: Record<string, string> = {
    approved: "Approved",
    APPROVED: "Approved",
    changes_requested: "Changes Requested",
    CHANGES_REQUESTED: "Changes Requested",
    commented: "Commented",
    COMMENTED: "Commented",
    dismissed: "Dismissed",
    DISMISSED: "Dismissed",
  };
  return labels[state] ?? state;
}

export function truncate(str: string, length: number): string {
  if (str.length <= length) return str;
  return str.slice(0, length) + "...";
}

export function buildQueryString(
  params: Record<string, string | number | boolean | undefined | null>
): string {
  const filtered = Object.entries(params).filter(
    ([, v]) => v !== undefined && v !== null && v !== ""
  );
  if (filtered.length === 0) return "";
  return "?" + new URLSearchParams(filtered.map(([k, v]) => [k, String(v)])).toString();
}
