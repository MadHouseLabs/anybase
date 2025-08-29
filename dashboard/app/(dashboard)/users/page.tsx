import { getUsers } from "@/lib/api-server";
import { getCurrentUser } from "@/lib/auth-server";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle, Users, Shield, UserCheck, UserX, Search, Code } from "lucide-react";
import { UsersTable } from "./users-table";
import { CreateUserButton } from "./users-client-components";
import { Input } from "@/components/ui/input";
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
                  <BreadcrumbPage>Users</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
            
            <div className="flex items-center justify-between mb-6">
              <div>
                <h1 className="text-2xl font-semibold">Users</h1>
                <p className="text-sm text-muted-foreground mt-1">
                  Manage system users and their access permissions
                </p>
              </div>
            </div>
          </div>
        </div>
        
        <div className="container mx-auto px-6 py-6 max-w-7xl">
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              You don't have permission to view users.
            </AlertDescription>
          </Alert>
        </div>
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

          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-semibold">Users</h1>
              <p className="text-sm text-muted-foreground mt-1">
                {canEditUsers ? "Manage system users and their access permissions" : "View system users (read-only access)"}
              </p>
            </div>
            <CreateUserButton canEditUsers={canEditUsers} />
          </div>

          {/* Metrics */}
          <div className="flex items-center gap-6 pt-4 border-t">
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{stats.total}</span>
              <span className="text-sm text-muted-foreground">Total Users</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <Shield className="h-4 w-4 text-blue-600" />
              <span className="text-2xl font-semibold">{stats.admins}</span>
              <span className="text-sm text-muted-foreground">Admins</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <Code className="h-4 w-4 text-purple-600" />
              <span className="text-2xl font-semibold">{stats.developers}</span>
              <span className="text-sm text-muted-foreground">Developers</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <UserCheck className="h-4 w-4 text-green-600" />
              <span className="text-2xl font-semibold">{stats.active}</span>
              <span className="text-sm text-muted-foreground">Active</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <UserX className="h-4 w-4 text-red-500" />
              <span className="text-2xl font-semibold">{stats.inactive}</span>
              <span className="text-sm text-muted-foreground">Inactive</span>
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="container mx-auto px-6 py-6 max-w-7xl">
        {/* Search */}
        <div className="mb-6">
          <div className="relative max-w-md">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <Input 
              placeholder="Search users..." 
              className="pl-10"
              disabled
            />
          </div>
        </div>
        
        {/* Users Table */}
        <UsersTable 
          users={users}
          canEditUsers={canEditUsers}
          currentUserEmail={currentUser?.email || ""}
        />
        
        {users.length > 0 && (
          <div className="flex justify-end mt-4">
            <p className="text-sm text-muted-foreground">
              Showing {users.length} user{users.length === 1 ? '' : 's'}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}