import { getAccessKeys } from "@/lib/api-server";
import { getCurrentUser } from "@/lib/auth-server";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Key, Shield, AlertCircle } from "lucide-react";
import { AccessKeysStats } from "./access-keys-stats";
import { AccessKeyCardDisplay } from "./access-key-card-server";
import { CreateAccessKeyButton, ToggleAccessKeyButton, RegenerateKeyButton, DeleteAccessKeyButton } from "./access-keys-client-components";
import { AccessKeysListWithSearch } from "./access-keys-list";

export default async function AccessKeysPage() {
  const [accessKeysData, currentUser] = await Promise.all([
    getAccessKeys(),
    getCurrentUser()
  ]);

  const accessKeys = accessKeysData?.access_keys || accessKeysData || [];
  
  // Check if current user has access (admin or developer)
  const hasAccess = currentUser?.role === "admin" || currentUser?.role === "developer";

  // Calculate statistics
  const stats = {
    total: accessKeys.length,
    active: accessKeys.filter((k: any) => k.active).length,
    inactive: accessKeys.filter((k: any) => !k.active).length,
    expiring: accessKeys.filter((k: any) => {
      if (!k.expires_at) return false;
      const daysUntilExpiry = Math.ceil((new Date(k.expires_at).getTime() - Date.now()) / (1000 * 60 * 60 * 24));
      return daysUntilExpiry > 0 && daysUntilExpiry <= 30;
    }).length
  };

  if (!hasAccess) {
    return (
      <div className="container mx-auto py-6 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
              <Key className="h-8 w-8" />
              Access Keys Management
            </h1>
            <p className="text-muted-foreground mt-1">
              Create and manage API access keys for programmatic access to your resources
            </p>
          </div>
        </div>
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Only administrators and developers can manage access keys.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Key className="h-8 w-8" />
            Access Keys Management
          </h1>
          <p className="text-muted-foreground mt-1">
            Create and manage API access keys for programmatic access to your resources
          </p>
        </div>
        <CreateAccessKeyButton />
      </div>

      {/* Statistics Cards */}
      <AccessKeysStats stats={stats} />

      {/* Security Notice */}
      <Alert className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-900">
        <Shield className="h-4 w-4 text-blue-600" />
        <AlertDescription className="text-blue-900 dark:text-blue-100">
          <strong>Security Best Practices:</strong> Never share or expose your API keys. Store them securely in environment variables or secret management systems. Rotate keys regularly and use the minimum required permissions.
        </AlertDescription>
      </Alert>

      {/* Main Content */}
      <Card>
        <CardHeader>
          <div className="space-y-4">
            <div>
              <CardTitle className="text-xl">API Keys</CardTitle>
              <CardDescription>
                Manage your API access keys and their permissions
              </CardDescription>
            </div>
            
            <AccessKeysListWithSearch accessKeys={accessKeys} />
          </div>
        </CardHeader>
      </Card>
    </div>
  );
}