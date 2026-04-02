"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { prsApi } from "@/lib/api";
import type { PRFilters } from "@/types";

export function usePRList(filters: PRFilters = {}) {
  return useQuery({
    queryKey: ["prs", filters],
    queryFn: async () => {
      const res = await prsApi.list(filters);
      return res.data;
    },
  });
}

export function usePR(id: string) {
  return useQuery({
    queryKey: ["prs", id],
    queryFn: async () => {
      const res = await prsApi.get(id);
      return res.data;
    },
    enabled: !!id,
  });
}

export function useCommentPR() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: string }) =>
      prsApi.comment(id, body),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["prs", id] });
    },
  });
}
