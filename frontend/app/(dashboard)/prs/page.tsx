"use client";

import { useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Header } from "@/components/layout/Header";
import { PRCard } from "@/components/prs/PRCard";
import { PRFiltersBar } from "@/components/prs/PRFilters";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { usePRList } from "@/hooks/usePRs";
import type { PRFilters } from "@/types";

export default function PRsPage() {
  const [filters, setFilters] = useState<PRFilters>({
    page: 1,
    per_page: 20,
  });

  const { data, isLoading, isFetching } = usePRList(filters);

  const prs = data?.pull_requests ?? [];
  const totalPages = data?.total_pages ?? 1;
  const currentPage = data?.page ?? 1;

  return (
    <div className="flex flex-col">
      <Header title="Pull Requests" />

      <div className="flex-1 space-y-4 p-6">
        {/* Filters */}
        <PRFiltersBar filters={filters} onChange={setFilters} />

        {/* Count */}
        {data && (
          <p className="text-xs text-muted-foreground">
            {data.total} pull request{data.total !== 1 ? "s" : ""}
            {isFetching && " — refreshing..."}
          </p>
        )}

        {/* List */}
        {isLoading ? (
          <div className="space-y-2">
            {[...Array(8)].map((_, i) => (
              <Skeleton key={i} className="h-16 rounded-lg" />
            ))}
          </div>
        ) : prs.length === 0 ? (
          <div className="rounded-lg border border-dashed py-16 text-center">
            <p className="text-sm text-muted-foreground">
              No pull requests found for the current filters
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {prs.map((pr) => (
              <PRCard key={pr.id} pr={pr} />
            ))}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-between pt-2">
            <p className="text-xs text-muted-foreground">
              Page {currentPage} of {totalPages}
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage <= 1}
                onClick={() =>
                  setFilters((f) => ({ ...f, page: (f.page ?? 1) - 1 }))
                }
              >
                <ChevronLeft className="h-4 w-4" />
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage >= totalPages}
                onClick={() =>
                  setFilters((f) => ({ ...f, page: (f.page ?? 1) + 1 }))
                }
              >
                Next
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
