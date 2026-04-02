"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import {
  ArrowLeft,
  GitMerge,
  GitPullRequest,
  XCircle,
  ExternalLink,
  FileCode,
  GitCommit,
  MessageSquare,
  Clock,
  Loader2,
  Plus,
  Minus,
} from "lucide-react";
import Link from "next/link";
import { Header } from "@/components/layout/Header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { usePR, useCommentPR } from "@/hooks/usePRs";
import {
  cn,
  formatDateTime,
  formatHours,
  reviewStateLabel,
} from "@/lib/utils";
import type { ReviewState } from "@/types";

const reviewStateStyles: Record<string, string> = {
  approved: "text-emerald-500",
  APPROVED: "text-emerald-500",
  changes_requested: "text-amber-500",
  CHANGES_REQUESTED: "text-amber-500",
  commented: "text-blue-500",
  COMMENTED: "text-blue-500",
  dismissed: "text-muted-foreground",
  DISMISSED: "text-muted-foreground",
};

export default function PRDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data, isLoading } = usePR(id);
  const commentMutation = useCommentPR();
  const [comment, setComment] = useState("");

  if (isLoading) {
    return (
      <div>
        <Header title="Pull Request" />
        <div className="space-y-4 p-6">
          <Skeleton className="h-8 w-3/4" />
          <Skeleton className="h-4 w-1/2" />
          <Skeleton className="h-32 w-full" />
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div>
        <Header title="Pull Request" />
        <div className="p-6 text-sm text-muted-foreground">
          Pull request not found.
        </div>
      </div>
    );
  }

  const { pull_request: pr, time_to_merge_hours, time_to_review_hours } = data;

  const statusConfig = {
    open: { icon: GitPullRequest, label: "Open", variant: "info" as const, iconClass: "text-blue-500" },
    merged: { icon: GitMerge, label: "Merged", variant: "success" as const, iconClass: "text-emerald-500" },
    closed: { icon: XCircle, label: "Closed", variant: "muted" as const, iconClass: "text-muted-foreground" },
  };

  const cfg = statusConfig[pr.status] ?? statusConfig.open;
  const StatusIcon = cfg.icon;

  const handleComment = () => {
    if (!comment.trim()) return;
    commentMutation.mutate(
      { id: pr.id, body: comment },
      { onSuccess: () => setComment("") }
    );
  };

  return (
    <div className="flex flex-col">
      <Header title="Pull Request" />

      <div className="flex-1 p-6 space-y-5">
        {/* Back link */}
        <Link
          href="/prs"
          className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to pull requests
        </Link>

        {/* Title */}
        <div>
          <div className="flex items-start gap-3">
            <StatusIcon className={cn("mt-1 h-5 w-5 shrink-0", cfg.iconClass)} />
            <div className="flex-1">
              <h1 className="text-xl font-bold leading-tight">{pr.title}</h1>
              <div className="mt-2 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                <Badge variant={cfg.variant}>{cfg.label}</Badge>
                {pr.repo && (
                  <a
                    href={pr.repo.html_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="font-medium text-foreground hover:underline"
                  >
                    {pr.repo.full_name}
                  </a>
                )}
                <span>#{pr.number}</span>
                <span>opened {formatDateTime(pr.github_created_at)}</span>
                <a
                  href={pr.html_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-1 hover:text-foreground"
                >
                  View on GitHub <ExternalLink className="h-3 w-3" />
                </a>
              </div>
            </div>
          </div>
        </div>

        <div className="grid gap-5 lg:grid-cols-3">
          {/* Main content */}
          <div className="space-y-5 lg:col-span-2">
            {/* Body */}
            {pr.body && (
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-semibold">Description</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="whitespace-pre-wrap text-sm text-muted-foreground">
                    {pr.body}
                  </p>
                </CardContent>
              </Card>
            )}

            {/* Reviews */}
            {(pr.reviews?.length ?? 0) > 0 && (
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-semibold">
                    Reviews ({pr.reviews?.length})
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  {pr.reviews?.map((review) => (
                    <div key={review.id} className="flex items-start gap-3">
                      <Avatar className="h-7 w-7">
                        <AvatarImage src={review.reviewer_avatar} />
                        <AvatarFallback className="text-xs">
                          {review.reviewer.charAt(0).toUpperCase()}
                        </AvatarFallback>
                      </Avatar>
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium">
                            {review.reviewer}
                          </span>
                          <span
                            className={cn(
                              "text-xs font-medium",
                              reviewStateStyles[review.state] ?? "text-muted-foreground"
                            )}
                          >
                            {reviewStateLabel(review.state as ReviewState)}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {formatDateTime(review.submitted_at)}
                          </span>
                        </div>
                        {review.body && (
                          <p className="mt-1 text-sm text-muted-foreground">
                            {review.body}
                          </p>
                        )}
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>
            )}

            {/* Add comment */}
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-semibold">Add a Comment</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <textarea
                  className="w-full rounded-md border bg-background px-3 py-2 text-sm resize-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  rows={4}
                  placeholder="Write a comment..."
                  value={comment}
                  onChange={(e) => setComment(e.target.value)}
                />
                <Button
                  onClick={handleComment}
                  disabled={!comment.trim() || commentMutation.isPending}
                  size="sm"
                >
                  {commentMutation.isPending && (
                    <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  )}
                  Post comment
                </Button>
              </CardContent>
            </Card>
          </div>

          {/* Sidebar */}
          <div className="space-y-4">
            {/* Author */}
            <Card>
              <CardContent className="p-4 space-y-3">
                <div className="flex items-center gap-2.5">
                  <Avatar className="h-8 w-8">
                    <AvatarImage src={pr.author_avatar_url} />
                    <AvatarFallback>{pr.author.charAt(0).toUpperCase()}</AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="text-xs text-muted-foreground">Author</p>
                    <p className="text-sm font-medium">{pr.author}</p>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3 border-t pt-3 text-xs">
                  <div>
                    <p className="text-muted-foreground">Branch</p>
                    <p className="font-mono font-medium truncate">{pr.head_branch}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Target</p>
                    <p className="font-mono font-medium truncate">{pr.base_branch}</p>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Metrics */}
            <Card>
              <CardContent className="p-4 space-y-3">
                <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                  Metrics
                </h3>
                <div className="space-y-2 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="flex items-center gap-1.5 text-muted-foreground">
                      <FileCode className="h-3.5 w-3.5" />
                      Files changed
                    </span>
                    <span className="font-medium">{pr.changed_files}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="flex items-center gap-1.5 text-muted-foreground">
                      <GitCommit className="h-3.5 w-3.5" />
                      Commits
                    </span>
                    <span className="font-medium">{pr.commit_count}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="flex items-center gap-1.5 text-emerald-500">
                      <Plus className="h-3.5 w-3.5" />
                      Additions
                    </span>
                    <span className="font-medium text-emerald-500">
                      +{pr.additions}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="flex items-center gap-1.5 text-destructive">
                      <Minus className="h-3.5 w-3.5" />
                      Deletions
                    </span>
                    <span className="font-medium text-destructive">
                      -{pr.deletions}
                    </span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="flex items-center gap-1.5 text-muted-foreground">
                      <MessageSquare className="h-3.5 w-3.5" />
                      Comments
                    </span>
                    <span className="font-medium">{pr.comment_count}</span>
                  </div>
                  {time_to_merge_hours != null && (
                    <div className="flex items-center justify-between">
                      <span className="flex items-center gap-1.5 text-muted-foreground">
                        <Clock className="h-3.5 w-3.5" />
                        Time to merge
                      </span>
                      <span className="font-medium">
                        {formatHours(time_to_merge_hours)}
                      </span>
                    </div>
                  )}
                  {time_to_review_hours != null && (
                    <div className="flex items-center justify-between">
                      <span className="flex items-center gap-1.5 text-muted-foreground">
                        <Clock className="h-3.5 w-3.5" />
                        First review
                      </span>
                      <span className="font-medium">
                        {formatHours(time_to_review_hours)}
                      </span>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
