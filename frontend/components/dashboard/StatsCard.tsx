import { type LucideIcon } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface StatsCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: LucideIcon;
  trend?: {
    value: number;
    label: string;
  };
  variant?: "default" | "primary" | "success" | "warning";
}

const variantStyles = {
  default: "text-muted-foreground",
  primary: "text-primary",
  success: "text-emerald-500",
  warning: "text-amber-500",
};

export function StatsCard({
  title,
  value,
  subtitle,
  icon: Icon,
  trend,
  variant = "default",
}: StatsCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">{title}</p>
            <p className="text-3xl font-bold tracking-tight">{value}</p>
            {subtitle && (
              <p className="text-xs text-muted-foreground">{subtitle}</p>
            )}
            {trend && (
              <p
                className={cn(
                  "text-xs font-medium",
                  trend.value >= 0 ? "text-emerald-500" : "text-destructive"
                )}
              >
                {trend.value >= 0 ? "+" : ""}
                {trend.value}% {trend.label}
              </p>
            )}
          </div>
          <div
            className={cn(
              "rounded-md p-2",
              variant === "primary" && "bg-primary/10",
              variant === "success" && "bg-emerald-500/10",
              variant === "warning" && "bg-amber-500/10",
              variant === "default" && "bg-muted"
            )}
          >
            <Icon className={cn("h-5 w-5", variantStyles[variant])} />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
