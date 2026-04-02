"use client";

import { GitPullRequest, GitMerge, Clock, Users } from "lucide-react";
import { Header } from "@/components/layout/Header";
import { StatsCard } from "@/components/dashboard/StatsCard";
import { PRChart } from "@/components/dashboard/PRChart";
import { StatusDonut } from "@/components/dashboard/StatusDonut";
import { Skeleton } from "@/components/ui/skeleton";
import { useAnalytics } from "@/hooks/useAnalytics";
import { formatHours } from "@/lib/utils";
import { usePRList } from "@/hooks/usePRs";
import { PRCard } from "@/components/prs/PRCard";

export default function DashboardPage() {
  const { data: analytics, isLoading } = useAnalytics();
  const { data: recentPRs } = usePRList({ per_page: 5 });

  return (
    <div className="flex flex-col">
      <Header title="Dashboard" />

      <div className="flex-1 space-y-6 p-6">
        {/* Stats grid */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {isLoading ? (
            [...Array(4)].map((_, i) => (
              <Skeleton key={i} className="h-28 rounded-lg" />
            ))
          ) : (
            <>
              <StatsCard
                title="Total Pull Requests"
                value={analytics?.summary.total_prs ?? 0}
                icon={GitPullRequest}
                variant="primary"
              />
              <StatsCard
                title="Open PRs"
                value={analytics?.summary.open_prs ?? 0}
                subtitle={`${analytics?.summary.merge_rate?.toFixed(1) ?? 0}% merge rate`}
                icon={GitPullRequest}
                variant="warning"
              />
              <StatsCard
                title="Merged PRs"
                value={analytics?.summary.merged_prs ?? 0}
                icon={GitMerge}
                variant="success"
              />
              <StatsCard
                title="Avg. Merge Time"
                value={formatHours(analytics?.summary.avg_merge_time_hours)}
                subtitle={
                  analytics?.summary.avg_review_time_hours != null
                    ? `First review: ${formatHours(analytics.summary.avg_review_time_hours)}`
                    : undefined
                }
                icon={Clock}
              />
            </>
          )}
        </div>

        {/* Charts row */}
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
          <div className="lg:col-span-2">
            {isLoading ? (
              <Skeleton className="h-80 rounded-lg" />
            ) : (
              <PRChart data={analytics?.daily_trend ?? []} />
            )}
          </div>
          <div>
            {isLoading ? (
              <Skeleton className="h-80 rounded-lg" />
            ) : (
              <StatusDonut
                data={
                  analytics?.status_breakdown ?? {
                    open: 0,
                    merged: 0,
                    closed: 0,
                  }
                }
              />
            )}
          </div>
        </div>

        {/* Recent PRs */}
        <div>
          <h2 className="mb-3 text-sm font-semibold text-muted-foreground uppercase tracking-wider">
            Recent Pull Requests
          </h2>
          <div className="space-y-2">
            {recentPRs?.pull_requests?.map((pr) => (
              <PRCard key={pr.id} pr={pr} />
            ))}
            {!recentPRs?.pull_requests?.length && !isLoading && (
              <p className="py-8 text-center text-sm text-muted-foreground">
                No pull requests found. Connect a repository to get started.
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
