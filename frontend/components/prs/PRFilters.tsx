"use client";

import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useRepos } from "@/hooks/useRepos";
import type { PRFilters } from "@/types";

const ALL = "all";

interface PRFiltersProps {
  filters: PRFilters;
  onChange: (filters: PRFilters) => void;
}

export function PRFiltersBar({ filters, onChange }: PRFiltersProps) {
  const { data } = useRepos();
  const repos = data?.repositories ?? [];

  return (
    <div className="flex flex-wrap items-center gap-2">
      <div className="relative">
        <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Filter by author..."
          className="pl-8 w-48"
          value={filters.author ?? ""}
          onChange={(e) =>
            onChange({ ...filters, author: e.target.value, page: 1 })
          }
        />
      </div>

      <Select
        value={filters.status || ALL}
        onValueChange={(v) =>
          onChange({
            ...filters,
            status: v === ALL ? "" : (v as PRFilters["status"]),
            page: 1,
          })
        }
      >
        <SelectTrigger className="w-36">
          <SelectValue placeholder="All statuses" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL}>All statuses</SelectItem>
          <SelectItem value="open">Open</SelectItem>
          <SelectItem value="merged">Merged</SelectItem>
          <SelectItem value="closed">Closed</SelectItem>
        </SelectContent>
      </Select>

      <Select
        value={filters.repo_id || ALL}
        onValueChange={(v) =>
          onChange({ ...filters, repo_id: v === ALL ? "" : v, page: 1 })
        }
      >
        <SelectTrigger className="w-52">
          <SelectValue placeholder="All repositories" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={ALL}>All repositories</SelectItem>
          {repos.map((r) => (
            <SelectItem key={r.id} value={r.id}>
              {r.full_name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Input
        type="date"
        className="w-40 text-sm"
        value={filters.from ?? ""}
        onChange={(e) => onChange({ ...filters, from: e.target.value, page: 1 })}
        aria-label="From date"
      />
      <span className="text-xs text-muted-foreground">to</span>
      <Input
        type="date"
        className="w-40 text-sm"
        value={filters.to ?? ""}
        onChange={(e) => onChange({ ...filters, to: e.target.value, page: 1 })}
        aria-label="To date"
      />

      {(filters.status || filters.repo_id || filters.author || filters.from || filters.to) && (
        <Button
          variant="ghost"
          size="sm"
          onClick={() => onChange({ page: 1, per_page: filters.per_page })}
          className="text-xs"
        >
          Clear filters
        </Button>
      )}
    </div>
  );
}
