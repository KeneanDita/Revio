"use client";

import { useState } from "react";
import {
  BookOpen,
  Plus,
  RefreshCw,
  Trash2,
  Lock,
  Globe,
  Loader2,
  ExternalLink,
} from "lucide-react";
import { Header } from "@/components/layout/Header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useRepos,
  useConnectRepo,
  useDisconnectRepo,
  useSyncRepo,
} from "@/hooks/useRepos";
import { formatRelativeTime } from "@/lib/utils";

export default function RepositoriesPage() {
  const [repoInput, setRepoInput] = useState("");
  const [syncingId, setSyncingId] = useState<string | null>(null);

  const { data, isLoading } = useRepos();
  const connectRepo = useConnectRepo();
  const disconnectRepo = useDisconnectRepo();
  const syncRepo = useSyncRepo();

  const repos = data?.repositories ?? [];

  const handleConnect = () => {
    const trimmed = repoInput.trim();
    if (!trimmed) return;
    connectRepo.mutate(trimmed, {
      onSuccess: () => setRepoInput(""),
    });
  };

  const handleSync = async (id: string) => {
    setSyncingId(id);
    await syncRepo.mutateAsync(id).finally(() => setSyncingId(null));
  };

  const connectError = (connectRepo.error as { response?: { data?: { error?: string } } })
    ?.response?.data?.error;

  return (
    <div className="flex flex-col">
      <Header title="Repositories" />

      <div className="flex-1 space-y-6 p-6">
        {/* Connect new repo */}
        <Card>
          <CardContent className="p-5">
            <h2 className="mb-1 text-sm font-semibold">Connect a Repository</h2>
            <p className="mb-4 text-xs text-muted-foreground">
              Enter the full name of a GitHub repository (e.g.{" "}
              <code className="rounded bg-muted px-1 py-0.5 font-mono text-xs">
                owner/repo-name
              </code>
              )
            </p>
            <div className="flex gap-2">
              <Input
                placeholder="owner/repository"
                value={repoInput}
                onChange={(e) => setRepoInput(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleConnect()}
                className="max-w-sm"
              />
              <Button
                onClick={handleConnect}
                disabled={!repoInput.trim() || connectRepo.isPending}
              >
                {connectRepo.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Plus className="h-4 w-4" />
                )}
                Connect
              </Button>
            </div>
            {connectError && (
              <p className="mt-2 text-xs text-destructive">{connectError}</p>
            )}
          </CardContent>
        </Card>

        {/* Repository list */}
        <div>
          <h2 className="mb-3 text-sm font-semibold text-muted-foreground uppercase tracking-wider">
            Connected Repositories ({repos.length})
          </h2>

          {isLoading ? (
            <div className="space-y-2">
              {[...Array(3)].map((_, i) => (
                <Skeleton key={i} className="h-20 rounded-lg" />
              ))}
            </div>
          ) : repos.length === 0 ? (
            <div className="rounded-lg border border-dashed py-12 text-center">
              <BookOpen className="mx-auto h-8 w-8 text-muted-foreground/50" />
              <p className="mt-3 text-sm text-muted-foreground">
                No repositories connected yet
              </p>
            </div>
          ) : (
            <div className="space-y-2">
              {repos.map((repo) => (
                <div
                  key={repo.id}
                  className="flex items-center justify-between rounded-lg border bg-card px-4 py-3"
                >
                  <div className="flex items-center gap-3">
                    {repo.private ? (
                      <Lock className="h-4 w-4 text-muted-foreground" />
                    ) : (
                      <Globe className="h-4 w-4 text-muted-foreground" />
                    )}
                    <div>
                      <div className="flex items-center gap-2">
                        <a
                          href={repo.html_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-sm font-medium hover:text-primary hover:underline"
                        >
                          {repo.full_name}
                        </a>
                        <ExternalLink className="h-3 w-3 text-muted-foreground" />
                        <Badge variant={repo.private ? "muted" : "info"} className="text-[10px]">
                          {repo.private ? "Private" : "Public"}
                        </Badge>
                      </div>
                      {repo.description && (
                        <p className="mt-0.5 text-xs text-muted-foreground line-clamp-1">
                          {repo.description}
                        </p>
                      )}
                      <p className="mt-0.5 text-xs text-muted-foreground">
                        {repo.last_sync_at
                          ? `Last synced ${formatRelativeTime(repo.last_sync_at)}`
                          : "Never synced"}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleSync(repo.id)}
                      disabled={syncingId === repo.id}
                      className="gap-1.5 text-xs"
                    >
                      <RefreshCw
                        className={`h-3.5 w-3.5 ${syncingId === repo.id ? "animate-spin" : ""}`}
                      />
                      Sync
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => disconnectRepo.mutate(repo.id)}
                      className="gap-1.5 text-xs text-destructive hover:text-destructive"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                      Remove
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
