import { getUsers } from "@/lib/api-server";
import { getCurrentUser } from "@/lib/auth-server";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { AlertCircle, Users } from "lucide-react";
import { UserStats } from "./user-stats";
import { UsersTable } from "./users-table";
import { CreateUserButton } from "./users-client-components";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

export default async function UsersPage() {
  const [usersData, currentUser] = await Promise.all([
    getUsers(),
    getCurrentUser()
  ]);

  const users = usersData?.users || usersData || [];
  
  // Check if current user is admin or developer
  const isAdmin = currentUser?.role === "admin";
  const isDeveloper = currentUser?.role === "developer";
  const canViewUsers = isAdmin || isDeveloper;
  const canEditUsers = isAdmin; // Only admins can edit

  if (!canViewUsers) {
    return (
      <div className="container mx-auto py-8 max-w-7xl">
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            You don't have permission to view users.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  const stats = {
    total: users.length,
    admins: users.filter((u: any) => u.role === "admin").length,
    developers: users.filter((u: any) => u.role === "developer").length,
    active: users.filter((u: any) => u.active).length,
    inactive: users.filter((u: any) => !u.active).length
  };

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
                <BreadcrumbPage>Users</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold">User Management</h1>
              <p className="text-sm text-muted-foreground mt-1">
                {canEditUsers ? "Manage system users and their access permissions" : "View system users (read-only access)"}
              </p>
            </div>
            <CreateUserButton canEditUsers={canEditUsers} />
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="container mx-auto px-6 py-6 max-w-7xl">
        {/* Statistics Cards */}
        <UserStats stats={stats} />

        {/* Users Table Section */}
        <div className="space-y-4">
          <div>
            <h2 className="text-xl font-semibold">All Users</h2>
            <p className="text-sm text-muted-foreground">
              Complete list of registered users and their access levels
            </p>
          </div>
          
          <UsersTable 
            users={users}
            canEditUsers={canEditUsers}
            currentUserEmail={currentUser?.email || ""}
          />
          
          {users.length > 0 && (
            <div className="flex justify-end">
              <p className="text-sm text-muted-foreground">
                Showing {users.length} user{users.length === 1 ? '' : 's'}
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}