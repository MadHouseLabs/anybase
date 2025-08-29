import { getAccessKeys } from "@/lib/api-server";
import { getCurrentUser } from "@/lib/auth-server";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Key, Shield, AlertCircle, Lock, Unlock, Clock, Search } from "lucide-react";
import { CreateAccessKeyButton } from "./access-keys-client-components";
import { AccessKeysListWithSearch } from "./access-keys-list";
import { Input } from "@/components/ui/input";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

export default async function AccessKeysPage() {
  const [accessKeysData, currentUser] = await Promise.all([
    getAccessKeys(),
    getCurrentUser()
  ]);

  const accessKeys = accessKeysData?.access_keys || accessKeysData || [];
  
  // Check if current user has access (admin or developer)
  const hasAccess = currentUser?.role === "admin" || currentUser?.role === "developer";

  // Calculate statistics
  const totalKeys = accessKeys.length;
  const activeKeys = accessKeys.filter((k: any) => k.active).length;
  const inactiveKeys = accessKeys.filter((k: any) => !k.active).length;
  const expiringKeys = accessKeys.filter((k: any) => {
    if (!k.expires_at) return false;
    const daysUntilExpiry = Math.ceil((new Date(k.expires_at).getTime() - Date.now()) / (1000 * 60 * 60 * 24));
    return daysUntilExpiry > 0 && daysUntilExpiry <= 30;
  }).length;

  if (!hasAccess) {
    return (
      <div className="flex flex-col min-h-screen">
        <div className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
          <div className="container mx-auto px-6 py-4 max-w-7xl">
            <Breadcrumb className="mb-4">
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbLink href="/">Dashboard</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  <BreadcrumbPage>Access Keys</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
            
            <div className="flex items-center justify-between mb-6">
              <div>
                <h1 className="text-2xl font-semibold">Access Keys</h1>
                <p className="text-sm text-muted-foreground mt-1">
                  Create and manage API access keys for programmatic access
                </p>
              </div>
            </div>
          </div>
        </div>
        
        <div className="container mx-auto px-6 py-6 max-w-7xl">
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Only administrators and developers can manage access keys.
            </AlertDescription>
          </Alert>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col min-h-screen">
      {/* Header */}
      <div className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container mx-auto px-6 py-4 max-w-7xl">
          {/* Breadcrumbs */}
          <Breadcrumb className="mb-4">
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink href="/">Dashboard</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>Access Keys</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-semibold">Access Keys</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Create and manage API access keys for programmatic access
              </p>
            </div>
            <CreateAccessKeyButton />
          </div>

          {/* Metrics */}
          <div className="flex items-center gap-6 pt-4 border-t">
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{totalKeys}</span>
              <span className="text-sm text-muted-foreground">Total Keys</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <Unlock className="h-4 w-4 text-green-600" />
              <span className="text-2xl font-semibold">{activeKeys}</span>
              <span className="text-sm text-muted-foreground">Active</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <Lock className="h-4 w-4 text-gray-400" />
              <span className="text-2xl font-semibold">{inactiveKeys}</span>
              <span className="text-sm text-muted-foreground">Inactive</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-orange-500" />
              <span className="text-2xl font-semibold">{expiringKeys}</span>
              <span className="text-sm text-muted-foreground">Expiring Soon</span>
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="container mx-auto px-6 py-6 max-w-7xl">
        {/* Security Notice */}
        <Alert className="mb-6">
          <Shield className="h-4 w-4" />
          <AlertDescription>
            <strong>Security Best Practices:</strong> Never share or expose your API keys. Store them securely in environment variables or secret management systems. Rotate keys regularly and use the minimum required permissions.
          </AlertDescription>
        </Alert>

        {/* Search */}
        <div className="mb-6">
          <div className="relative max-w-md">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <Input 
              placeholder="Search access keys..." 
              className="pl-10"
              disabled
            />
          </div>
        </div>

        {/* Access Keys List */}
        <AccessKeysListWithSearch accessKeys={accessKeys} />
      </div>
    </div>
  );
}