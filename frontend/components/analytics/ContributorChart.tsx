"use client";

import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type { ReviewerMetric, AuthorMetric } from "@/types";

interface ReviewerChartProps {
  data: ReviewerMetric[];
}

interface AuthorChartProps {
  data: AuthorMetric[];
}

export function ReviewerChart({ data }: ReviewerChartProps) {
  const sliced = data.slice(0, 8);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base font-semibold">Top Reviewers</CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={240}>
          <BarChart
            data={sliced}
            layout="vertical"
            margin={{ top: 0, right: 12, left: 0, bottom: 0 }}
          >
            <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="hsl(var(--border))" />
            <XAxis
              type="number"
              tick={{ fontSize: 11, fill: "hsl(var(--muted-foreground))" }}
              tickLine={false}
              axisLine={false}
              allowDecimals={false}
            />
            <YAxis
              type="category"
              dataKey="reviewer"
              width={90}
              tick={{ fontSize: 11, fill: "hsl(var(--muted-foreground))" }}
              tickLine={false}
              axisLine={false}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: "hsl(var(--popover))",
                border: "1px solid hsl(var(--border))",
                borderRadius: "6px",
                fontSize: 12,
              }}
            />
            <Bar dataKey="review_count" name="Reviews" fill="#3b82f6" radius={[0, 3, 3, 0]} />
            <Bar dataKey="approval_count" name="Approvals" fill="#10b981" radius={[0, 3, 3, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}

export function AuthorTable({ data }: AuthorChartProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base font-semibold">Top Authors</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {data.slice(0, 8).map((author) => (
            <div
              key={author.author}
              className="flex items-center gap-3 rounded-md px-2 py-1.5 hover:bg-muted/50"
            >
              <Avatar className="h-7 w-7">
                <AvatarImage src={author.author_avatar} />
                <AvatarFallback className="text-xs">
                  {author.author.charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="flex-1 min-w-0">
                <p className="truncate text-sm font-medium">{author.author}</p>
              </div>
              <div className="text-right text-xs text-muted-foreground">
                <p>{author.pr_count} PRs</p>
                <p className="text-emerald-500">{author.merged_count} merged</p>
              </div>
            </div>
          ))}
          {data.length === 0 && (
            <p className="py-6 text-center text-sm text-muted-foreground">
              No data for the selected period
            </p>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
