import Link from "next/link";
import { GitPullRequest, GitMerge, XCircle, MessageSquare, Clock } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { cn, formatRelativeTime, formatHours } from "@/lib/utils";
import type { PullRequest } from "@/types";

interface PRCardProps {
  pr: PullRequest;
}

const statusConfig = {
  open: {
    label: "Open",
    icon: GitPullRequest,
    badgeVariant: "info" as const,
    iconClass: "text-blue-500",
  },
  merged: {
    label: "Merged",
    icon: GitMerge,
    badgeVariant: "success" as const,
    iconClass: "text-emerald-500",
  },
  closed: {
    label: "Closed",
    icon: XCircle,
    badgeVariant: "muted" as const,
    iconClass: "text-muted-foreground",
  },
};

export function PRCard({ pr }: PRCardProps) {
  const config = statusConfig[pr.status] ?? statusConfig.open;
  const StatusIcon = config.icon;

  const mergeTimeHours =
    pr.merged_at && pr.github_created_at
      ? (new Date(pr.merged_at).getTime() -
          new Date(pr.github_created_at).getTime()) /
        3_600_000
      : null;

  return (
    <Link
      href={`/prs/${pr.id}`}
      className="group flex items-start gap-3 rounded-lg border bg-card px-4 py-3 transition-colors hover:bg-accent/30"
    >
      <StatusIcon
        className={cn("mt-0.5 h-4 w-4 shrink-0", config.iconClass)}
      />

      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-2">
          <p className="truncate text-sm font-medium group-hover:text-primary">
            {pr.title}
          </p>
          <Badge variant={config.badgeVariant} className="shrink-0">
            {config.label}
          </Badge>
        </div>

        <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
          {pr.Repo && (
            <span className="font-medium text-foreground/70">
              {pr.Repo.full_name}
            </span>
          )}
          <span>#{pr.number}</span>

          <span className="flex items-center gap-1">
            <Avatar className="h-4 w-4">
              <AvatarImage src={pr.author_avatar_url} />
              <AvatarFallback className="text-[8px]">
                {pr.author.charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            {pr.author}
          </span>

          <span>{formatRelativeTime(pr.github_created_at)}</span>

          {pr.comment_count > 0 && (
            <span className="flex items-center gap-1">
              <MessageSquare className="h-3 w-3" />
              {pr.comment_count}
            </span>
          )}

          {mergeTimeHours !== null && (
            <span className="flex items-center gap-1">
              <Clock className="h-3 w-3" />
              {formatHours(mergeTimeHours)}
            </span>
          )}
        </div>
      </div>
    </Link>
  );
}
