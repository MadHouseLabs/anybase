import { getCollections } from "@/lib/api-server";
import { CollectionsTable } from "./collections-table";
import { CreateCollectionButton } from "./collection-client-components";
import { Input } from "@/components/ui/input";
import { Search, Database } from "lucide-react";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

export default async function CollectionsPage() {
  const data = await getCollections();
  const collections = data?.collections || data || [];

  // Calculate statistics
  const totalCollections = collections.length;
  const totalDocuments = collections.reduce((sum: number, col: any) => sum + (col.document_count || 0), 0);
  const withVersioning = collections.filter((col: any) => col.settings?.versioning).length;
  const withAuditing = collections.filter((col: any) => col.settings?.auditing).length;

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
                <BreadcrumbPage>Collections</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-semibold">Collections</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Manage your NoSQL document collections
              </p>
            </div>
            <CreateCollectionButton />
          </div>

          {/* Metrics */}
          <div className="flex items-center gap-6 pt-4 border-t">
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{totalCollections}</span>
              <span className="text-sm text-muted-foreground">Collections</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{totalDocuments.toLocaleString()}</span>
              <span className="text-sm text-muted-foreground">Documents</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{withVersioning}</span>
              <span className="text-sm text-muted-foreground">With Versioning</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{withAuditing}</span>
              <span className="text-sm text-muted-foreground">With Auditing</span>
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
              placeholder="Search collections..." 
              className="pl-10"
              disabled
            />
          </div>
        </div>

        {/* Collections Table */}
        <CollectionsTable collections={collections} />
      </div>
    </div>
  );
}