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
  LucideIcon
} from "lucide-react";

// Map icon names to components
const iconMap: Record<string, LucideIcon> = {
  "home": Home,
  "database": Database,
  "eye": Eye,
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
    <div className="flex flex-col gap-1">
      {routes.map((route) => {
        const Icon = iconMap[route.iconName] || Home;
        return (
          <Link
            key={route.href}
            href={route.href}
            className="block"
          >
            <Button
              variant={pathname === route.href ? "secondary" : "ghost"}
              className="w-full justify-start h-10 mb-1"
            >
              <Icon className="mr-3 h-4 w-4" />
              {route.label}
            </Button>
          </Link>
        );
      })}
    </div>
  );
}