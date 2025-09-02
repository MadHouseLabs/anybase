import { cn } from "@/lib/utils";
import { ScrollArea } from "@/components/ui/scroll-area";
import { SidebarNav } from "./sidebar-nav";

interface SidebarProps extends React.HTMLAttributes<HTMLDivElement> {}

export function Sidebar({ className }: SidebarProps) {
  const dashboardRoute = {
    label: "Dashboard",
    iconName: "home",
    href: "/",
  };

  const dataRoutes = [
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
      label: "Storage",
      iconName: "hard-drive",
      href: "/storage",
    },
  ];

  const computeRoutes = [
    {
      label: "Functions",
      iconName: "code-2",
      href: "/functions",
    },
    {
      label: "Event Store",
      iconName: "inbox",
      href: "/event-store",
    },
    {
      label: "Cron",
      iconName: "clock",
      href: "/cron",
    },
    {
      label: "Realtime",
      iconName: "radio",
      href: "/realtime",
    },
  ];

  const securityRoutes = [
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
      label: "Integrations",
      iconName: "plug",
      href: "/integrations",
    },
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
    <div className={cn("w-64 h-full border-r bg-background", className)}>
      <div className="flex flex-col h-full">
        {/* Logo Section */}
        <div className="h-16 flex items-center px-6 border-b">
          <div className="flex items-center gap-3">
            <div className="h-8 w-8 rounded-lg bg-primary flex items-center justify-center">
              <span className="text-primary-foreground font-bold text-sm">A</span>
            </div>
            <div>
              <h2 className="font-semibold text-base leading-none">AnyBase</h2>
              <p className="text-xs text-muted-foreground mt-0.5">Cloud Database</p>
            </div>
          </div>
        </div>

        {/* Navigation */}
        <ScrollArea className="flex-1 px-3 py-4">
          {/* Dashboard */}
          <div className="px-2">
            <SidebarNav routes={[dashboardRoute]} />
          </div>
          
          {/* Data Management */}
          <div className="mt-6">
            <h3 className="px-4 mb-2 text-xs font-medium text-muted-foreground/70">
              DATA MANAGEMENT
            </h3>
            <div className="px-2">
              <SidebarNav routes={dataRoutes} />
            </div>
          </div>
          
          {/* Compute & Processing */}
          <div className="mt-6">
            <h3 className="px-4 mb-2 text-xs font-medium text-muted-foreground/70">
              COMPUTE & PROCESSING
            </h3>
            <div className="px-2">
              <SidebarNav routes={computeRoutes} />
            </div>
          </div>
          
          {/* Security */}
          <div className="mt-6">
            <h3 className="px-4 mb-2 text-xs font-medium text-muted-foreground/70">
              SECURITY & ACCESS
            </h3>
            <div className="px-2">
              <SidebarNav routes={securityRoutes} />
            </div>
          </div>
        </ScrollArea>

        {/* Bottom section */}
        <div className="border-t p-3">
          <div className="px-2">
            <SidebarNav routes={bottomRoutes} />
          </div>
        </div>
      </div>
    </div>
  );
}