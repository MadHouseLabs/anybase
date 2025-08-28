import { cn } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";
import { SidebarNav } from "./sidebar-nav";

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {}

export function Sidebar({ className }: SidebarProps) {
  const mainRoutes = [
    {
      label: "Dashboard",
      iconName: "home",
      href: "/",
    },
    {
      label: "Collections",
      iconName: "database",
      href: "/collections",
    },
    {
      label: "Views",
      iconName: "eye",
      href: "/views",
    },
    {
      label: "Users",
      iconName: "users",
      href: "/users",
    },
    {
      label: "Access Keys",
      iconName: "key",
      href: "/access-keys",
    },
  ];

  const bottomRoutes = [
    {
      label: "API Docs",
      iconName: "book-open",
      href: "/api-docs",
    },
    {
      label: "Settings",
      iconName: "settings",
      href: "/settings",
    },
  ];

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
            <SidebarNav routes={mainRoutes} />
            <div className="my-4 border-t" />
            <SidebarNav routes={bottomRoutes} />
          </ScrollArea>
        </div>
      </div>
    </div>
  );
}