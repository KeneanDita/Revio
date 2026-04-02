"use client";

import { useState } from "react";
import {
  GitMerge,
  Clock,
  Users,
  TrendingUp,
  GitPullRequest,
} from "lucide-react";
import { Header } from "@/components/layout/Header";
import { StatsCard } from "@/components/dashboard/StatsCard";
import { MergeTimeChart } from "@/components/analytics/MergeTimeChart";
import { ReviewerChart, AuthorTable } from "@/components/analytics/ContributorChart";
import { PRChart } from "@/components/dashboard/PRChart";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { useAnalytics } from "@/hooks/useAnalytics";
import { useRepos } from "@/hooks/useRepos";
import { formatHours } from "@/lib/utils";
import { format, subDays } from "date-fns";
import type { AnalyticsFilters } from "@/types";

type Preset = "7d" | "30d" | "90d" | "custom";

function presetDates(preset: Preset): { from: string; to: string } {
  const today = new Date();
  const to = format(today, "yyyy-MM-dd");
  const presetMap: Record<Exclude<Preset, "custom">, number> = {
    "7d": 7,
    "30d": 30,
    "90d": 90,
  };
  if (preset === "custom") return { from: "", to };
  const days = presetMap[preset];
  return { from: format(subDays(today, days), "yyyy-MM-dd"), to };
}

export default function AnalyticsPage() {
  const [preset, setPreset] = useState<Preset>("30d");
  const [filters, setFilters] = useState<AnalyticsFilters>(() => presetDates("30d"));

  const { data: repos } = useRepos();
  const { data: analytics, isLoading } = useAnalytics(filters);

  const applyPreset = (p: Preset) => {
    setPreset(p);
    if (p !== "custom") {
      setFilters((f) => ({ ...f, ...presetDates(p) }));
    }
  };

  return (
    <div className="flex flex-col">
      <Header title="Analytics" />

      <div className="flex-1 space-y-6 p-6">
        {/* Filters bar */}
        <div className="flex flex-wrap items-center gap-2">
          {(["7d", "30d", "90d", "custom"] as Preset[]).map((p) => (
            <Button
              key={p}
              variant={preset === p ? "default" : "outline"}
              size="sm"
              onClick={() => applyPreset(p)}
            >
              {p === "7d"
                ? "Last 7 days"
                : p === "30d"
                ? "Last 30 days"
                : p === "90d"
                ? "Last 90 days"
                : "Custom"}
            </Button>
          ))}

          {preset === "custom" && (
            <>
              <Input
                type="date"
                className="w-40 text-sm"
                value={filters.from ?? ""}
                onChange={(e) =>
                  setFilters((f) => ({ ...f, from: e.target.value }))
                }
              />
              <span className="text-xs text-muted-foreground">to</span>
              <Input
                type="date"
                className="w-40 text-sm"
                value={filters.to ?? ""}
                onChange={(e) =>
                  setFilters((f) => ({ ...f, to: e.target.value }))
                }
              />
            </>
          )}

          <Select
            value={filters.repo_id || "all"}
            onValueChange={(v) =>
              setFilters((f) => ({ ...f, repo_id: v === "all" ? "" : v }))
            }
          >
            <SelectTrigger className="w-52">
              <SelectValue placeholder="All repositories" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All repositories</SelectItem>
              {repos?.repositories?.map((r) => (
                <SelectItem key={r.id} value={r.id}>
                  {r.full_name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Summary stats */}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          {isLoading ? (
            [...Array(4)].map((_, i) => (
              <Skeleton key={i} className="h-28 rounded-lg" />
            ))
          ) : (
            <>
              <StatsCard
                title="Total PRs"
                value={analytics?.summary.total_prs ?? 0}
                icon={GitPullRequest}
                variant="primary"
              />
              <StatsCard
                title="Merged PRs"
                value={analytics?.summary.merged_prs ?? 0}
                subtitle={`${analytics?.summary.merge_rate?.toFixed(1) ?? 0}% merge rate`}
                icon={GitMerge}
                variant="success"
              />
              <StatsCard
                title="Avg. Merge Time"
                value={formatHours(analytics?.summary.avg_merge_time_hours)}
                icon={Clock}
              />
              <StatsCard
                title="Avg. Time to Review"
                value={formatHours(analytics?.summary.avg_review_time_hours)}
                icon={TrendingUp}
                variant="warning"
              />
            </>
          )}
        </div>

        {/* Charts */}
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {isLoading ? (
            <>
              <Skeleton className="h-72 rounded-lg" />
              <Skeleton className="h-72 rounded-lg" />
            </>
          ) : (
            <>
              <PRChart data={analytics?.daily_trend ?? []} />
              <MergeTimeChart data={analytics?.daily_trend ?? []} />
            </>
          )}
        </div>

        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {isLoading ? (
            <>
              <Skeleton className="h-64 rounded-lg" />
              <Skeleton className="h-64 rounded-lg" />
            </>
          ) : (
            <>
              <ReviewerChart data={analytics?.top_reviewers ?? []} />
              <AuthorTable data={analytics?.top_authors ?? []} />
            </>
          )}
        </div>
      </div>
    </div>
  );
}
