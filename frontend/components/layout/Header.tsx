"use client";

import { Bell } from "lucide-react";
import { ThemeToggle } from "./ThemeToggle";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import { useNotifications, useMarkAllRead } from "@/hooks/useNotifications";
import { cn, formatRelativeTime } from "@/lib/utils";
import Link from "next/link";

interface HeaderProps {
  title: string;
}

export function Header({ title }: HeaderProps) {
  const { data } = useNotifications();
  const markAllRead = useMarkAllRead();

  const unreadCount = data?.unread_count ?? 0;
  const notifications = data?.notifications?.slice(0, 8) ?? [];

  return (
    <header className="sticky top-0 z-40 flex h-16 items-center justify-between border-b bg-background/95 px-6 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <h1 className="text-xl font-semibold">{title}</h1>

      <div className="flex items-center gap-2">
        <ThemeToggle />

        {/* Notifications */}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="relative h-8 w-8">
              <Bell className="h-4 w-4" />
              {unreadCount > 0 && (
                <span className="absolute -right-0.5 -top-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-primary text-[10px] font-bold text-primary-foreground">
                  {unreadCount > 9 ? "9+" : unreadCount}
                </span>
              )}
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-80">
            <div className="flex items-center justify-between px-3 py-2">
              <span className="text-sm font-semibold">Notifications</span>
              {unreadCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-auto px-2 py-0 text-xs"
                  onClick={() => markAllRead.mutate()}
                >
                  Mark all read
                </Button>
              )}
            </div>
            <DropdownMenuSeparator />
            {notifications.length === 0 ? (
              <div className="px-3 py-6 text-center text-sm text-muted-foreground">
                No notifications
              </div>
            ) : (
              notifications.map((n) => (
                <DropdownMenuItem key={n.id} asChild>
                  <Link
                    href={n.link || "#"}
                    className={cn(
                      "flex flex-col gap-0.5 px-3 py-2",
                      !n.read && "bg-accent/50"
                    )}
                  >
                    <span
                      className={cn(
                        "text-sm",
                        !n.read && "font-medium"
                      )}
                    >
                      {n.title}
                    </span>
                    <span className="text-xs text-muted-foreground line-clamp-1">
                      {n.body}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {formatRelativeTime(n.created_at)}
                    </span>
                  </Link>
                </DropdownMenuItem>
              ))
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </header>
  );
}
