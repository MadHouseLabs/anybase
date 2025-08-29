import { getViews, getCollections } from "@/lib/api-server";
import { getCurrentUser } from "@/lib/auth-server";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Eye, AlertCircle } from "lucide-react";
import { CreateViewDialog } from "./view-client-components";
import { ViewsListWithSearch } from "./views-list";

export default async function ViewsPage() {
  const [viewsData, collectionsData, currentUser] = await Promise.all([
    getViews(),
    getCollections(), 
    getCurrentUser()
  ]);

  // Ensure views is always an array
  let views: any[] = [];
  if (Array.isArray(viewsData)) {
    views = viewsData;
  } else if (viewsData?.views && Array.isArray(viewsData.views)) {
    views = viewsData.views;
  } else if (viewsData?.data && Array.isArray(viewsData.data)) {
    views = viewsData.data;
  }

  // Ensure collections is always an array
  let collections: any[] = [];
  if (Array.isArray(collectionsData)) {
    collections = collectionsData;
  } else if (collectionsData?.collections && Array.isArray(collectionsData.collections)) {
    collections = collectionsData.collections;
  } else if (collectionsData?.data && Array.isArray(collectionsData.data)) {
    collections = collectionsData.data;
  }
  
  // Check if current user has access (admin or developer)
  const hasAccess = currentUser?.role === "admin" || currentUser?.role === "developer";

  if (!hasAccess) {
    return (
      <div className="container mx-auto py-8 max-w-7xl">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
              <Eye className="h-8 w-8" />
              Views Management
            </h1>
            <p className="text-muted-foreground mt-1">
              Create and manage saved queries for your collections
            </p>
          </div>
        </div>
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Only administrators and developers can manage views.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8 max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Eye className="h-8 w-8" />
            Views Management
          </h1>
          <p className="text-muted-foreground mt-1">
            Create and manage saved queries for your collections
          </p>
        </div>
      </div>

      {/* Info Alert */}
      <Alert className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-900">
        <Eye className="h-4 w-4 text-blue-600" />
        <AlertDescription className="text-blue-900 dark:text-blue-100">
          <strong>About Views:</strong> Views are saved queries that allow you to filter and project data from a collection. 
          They provide a consistent way to access frequently used data subsets without writing the same queries repeatedly.
        </AlertDescription>
      </Alert>

      {/* Main Content */}
      <Card>
        <CardHeader>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="text-xl">Saved Views</CardTitle>
                <CardDescription>
                  Manage your collection views and saved queries
                </CardDescription>
              </div>
              <CreateViewDialog collections={collections} />
            </div>
            
            <ViewsListWithSearch views={views} collections={collections} />
          </div>
        </CardHeader>
      </Card>
    </div>
  );
}