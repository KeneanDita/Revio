"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { reposApi } from "@/lib/api";

export function useRepos() {
  return useQuery({
    queryKey: ["repos"],
    queryFn: async () => {
      const res = await reposApi.list();
      return res.data;
    },
  });
}

export function useGitHubRepos() {
  return useQuery({
    queryKey: ["repos", "github"],
    queryFn: async () => {
      const res = await reposApi.listGitHub();
      return res.data;
    },
    staleTime: 30 * 1000,
  });
}

export function useConnectRepo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (fullName: string) => reposApi.connect(fullName),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["repos"] });
    },
  });
}

export function useDisconnectRepo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => reposApi.disconnect(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["repos"] });
    },
  });
}

export function useSyncRepo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => reposApi.sync(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["repos"] });
    },
  });
}
