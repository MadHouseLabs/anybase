import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { 
  Database, 
  FileJson, 
  Settings, 
  LogOut,
  Home,
  Eye,
  Users,
  Key
} from "lucide-react"
import { Link, useLocation } from "react-router-dom"

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {}

export function Sidebar({ className }: SidebarProps) {
  const location = useLocation()

  const routes = [
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
              {routes.map((route) => (
                <Link
                  key={route.href}
                  to={route.href}
                  className="block"
                >
                  <Button
                    variant={location.pathname === route.href ? "secondary" : "ghost"}
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
        <div className="px-3 py-2">
          <Button
            variant="ghost"
            className="w-full justify-start text-destructive hover:text-destructive"
            onClick={() => {
              localStorage.removeItem('token')
              window.location.href = '/login'
            }}
          >
            <LogOut className="mr-2 h-4 w-4" />
            Logout
          </Button>
        </div>
      </div>
    </div>
  )
}