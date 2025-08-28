import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { 
  Database, 
  FileJson, 
  Settings,
  Home,
  Eye,
  Users,
  Key,
  BookOpen
} from "lucide-react"
"use client";

import Link from "next/link";
import { usePathname } from "next/navigation"

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {}

export function Sidebar({ className }: SidebarProps) {
  const pathname = usePathname()

  const mainRoutes = [
    {
      label: "Dashboard",
      icon: Home,
      href: "/",
    },
    {
      label: "Collections",
      icon: Database,
      href: "/collections",
    },
    {
      label: "Views",
      icon: Eye,
      href: "/views",
    },
    {
      label: "Users",
      icon: Users,
      href: "/users",
    },
    {
      label: "Access Keys",
      icon: Key,
      href: "/access-keys",
    },
  ]

  const bottomRoutes = [
    {
      label: "API Docs",
      icon: BookOpen,
      href: "/api-docs",
    },
    {
      label: "Settings",
      icon: Settings,
      href: "/settings",
    },
  ]

  return (
    <div className={cn("pb-12 w-64 h-full", className)}>
      <div className="space-y-4 py-4 flex flex-col h-full">
        <div className="px-3 py-2 flex-1">
          <div className="mb-8">
            <h2 className="mb-2 px-4 text-2xl font-semibold tracking-tight">
              AnyBase
            </h2>
            <p className="px-4 text-sm text-muted-foreground">
              Database Management
            </p>
          </div>
          <ScrollArea className="h-[calc(100vh-200px)]">
            <div className="flex flex-col gap-1">
              {mainRoutes.map((route) => (
                <Link
                  key={route.href}
                  to={route.href}
                  className="block"
                >
                  <Button
                    variant={pathname === route.href ? "secondary" : "ghost"}
                    className="w-full justify-start h-10 mb-1"
                  >
                    <route.icon className="mr-3 h-4 w-4" />
                    {route.label}
                  </Button>
                </Link>
              ))}
              
              <div className="my-4 border-t" />
              
              {bottomRoutes.map((route) => (
                <Link
                  key={route.href}
                  to={route.href}
                  className="block"
                >
                  <Button
                    variant={pathname === route.href ? "secondary" : "ghost"}
                    className="w-full justify-start h-10 mb-1"
                  >
                    <route.icon className="mr-3 h-4 w-4" />
                    {route.label}
                  </Button>
                </Link>
              ))}
            </div>
          </ScrollArea>
        </div>
      </div>
    </div>
  )
}