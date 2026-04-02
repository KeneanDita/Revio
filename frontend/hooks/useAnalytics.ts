"use client";

import { useQuery } from "@tanstack/react-query";
import { analyticsApi } from "@/lib/api";
import type { AnalyticsFilters } from "@/types";

export function useAnalytics(filters: AnalyticsFilters = {}) {
  return useQuery({
    queryKey: ["analytics", filters],
    queryFn: async () => {
      const res = await analyticsApi.get(filters);
      return res.data;
    },
    staleTime: 2 * 60 * 1000,
  });
}
