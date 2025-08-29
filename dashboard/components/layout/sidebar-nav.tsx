"use client";

import { usePathname } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { 
  Database, 
  Settings,
  Home,
  Eye,
  Users,
  Key,
  BookOpen,
  Code2,
  Inbox,
  HardDrive,
  Clock,
  Radio,
  LucideIcon
} from "lucide-react";

// Map icon names to components
const iconMap: Record<string, LucideIcon> = {
  "home": Home,
  "database": Database,
  "eye": Eye,
  "code-2": Code2,
  "inbox": Inbox,
  "hard-drive": HardDrive,
  "clock": Clock,
  "radio": Radio,
  "users": Users,
  "key": Key,
  "book-open": BookOpen,
  "settings": Settings,
};

interface NavRoute {
  label: string;
  iconName: string;
  href: string;
}

interface SidebarNavProps {
  routes: NavRoute[];
}

export function SidebarNav({ routes }: SidebarNavProps) {
  const pathname = usePathname();

  return (
    <div className="flex flex-col gap-0.5">
      {routes.map((route) => {
        const Icon = iconMap[route.iconName] || Home;
        const isActive = pathname === route.href;
        
        return (
          <Link
            key={route.href}
            href={route.href}
            className={`
              flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium
              transition-all duration-150 group relative
              ${isActive 
                ? 'bg-primary/10 text-primary' 
                : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
              }
            `}
          >
            {isActive && (
              <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-5 bg-primary rounded-r-full" />
            )}
            <Icon className={`h-4 w-4 flex-shrink-0 ${isActive ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground'}`} />
            <span className="flex-1">{route.label}</span>
            {isActive && (
              <div className="h-1.5 w-1.5 rounded-full bg-primary animate-pulse" />
            )}
          </Link>
        );
      })}
    </div>
  );
}