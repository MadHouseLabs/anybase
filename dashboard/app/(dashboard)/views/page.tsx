import { getViews, getCollections } from "@/lib/api-server";
import { getCurrentUser } from "@/lib/auth-server";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Eye, AlertCircle, Database, Filter, SortAsc } from "lucide-react";
import { CreateViewDialog } from "./view-client-components";
import { ViewsTable } from "./views-table";
import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

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

  // Calculate statistics
  const totalViews = views.length;
  const collectionsWithViews = [...new Set(views.map((v: any) => v.collection))].length;
  const viewsWithFilters = views.filter((v: any) => v.filter && Object.keys(v.filter).length > 0).length;
  const viewsWithProjection = views.filter((v: any) => v.fields && v.fields.length > 0).length;

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
                  <BreadcrumbPage>Views</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
            
            <div className="flex items-center justify-between mb-6">
              <div>
                <h1 className="text-2xl font-semibold">Views</h1>
                <p className="text-sm text-muted-foreground mt-1">
                  Create and manage saved queries for your collections
                </p>
              </div>
            </div>
          </div>
        </div>
        
        <div className="container mx-auto px-6 py-6 max-w-7xl">
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              Only administrators and developers can manage views.
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
                <BreadcrumbPage>Views</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-semibold">Views</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Create and manage saved queries for your collections
              </p>
            </div>
            <CreateViewDialog collections={collections} />
          </div>

          {/* Metrics */}
          <div className="flex items-center gap-6 pt-4 border-t">
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{totalViews}</span>
              <span className="text-sm text-muted-foreground">Total Views</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{collectionsWithViews}</span>
              <span className="text-sm text-muted-foreground">Collections</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{viewsWithFilters}</span>
              <span className="text-sm text-muted-foreground">With Filters</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{viewsWithProjection}</span>
              <span className="text-sm text-muted-foreground">With Projection</span>
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
              placeholder="Search views..." 
              className="pl-10"
              disabled
            />
          </div>
        </div>

        {/* Views Table */}
        <ViewsTable views={views} collections={collections} />
      </div>
    </div>
  );
}