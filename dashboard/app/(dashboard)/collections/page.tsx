import { getCollections } from "@/lib/api-server";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Database, FileJson, Shield, Settings, Calendar, Hash, Plus, FolderOpen, ArrowRight, Info, CheckCircle, Clock, Activity, TrendingUp } from "lucide-react";
import { CreateCollectionButton, DeleteCollectionButton } from "./collection-client-components";
import Link from "next/link";
import { format } from 'date-fns';

export default async function CollectionsPage() {
  const data = await getCollections();
  const collections = data?.collections || data || [];

  // Calculate statistics
  const stats = {
    total: collections.length,
    documents: collections.reduce((sum: number, col: any) => sum + (col.document_count || 0), 0),
    withVersioning: collections.filter((col: any) => col.settings?.versioning).length,
    withAuditing: collections.filter((col: any) => col.settings?.auditing).length,
  };

  // Sort collections by document count for "Top Collections" section
  const topCollections = [...collections]
    .sort((a: any, b: any) => (b.document_count || 0) - (a.document_count || 0))
    .slice(0, 3);

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Database className="h-8 w-8" />
            Collections Management
          </h1>
          <p className="text-muted-foreground mt-1">
            Organize and manage your data with powerful NoSQL collections
          </p>
        </div>
        <CreateCollectionButton />
      </div>

      {/* Statistics Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Collections</CardTitle>
            <div className="p-2 rounded-lg bg-blue-100">
              <FolderOpen className="h-4 w-4 text-blue-600" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total}</div>
            <p className="text-xs text-muted-foreground">
              Active collections
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Documents</CardTitle>
            <div className="p-2 rounded-lg bg-green-100">
              <FileJson className="h-4 w-4 text-green-600" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.documents.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">
              Across all collections
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">With Versioning</CardTitle>
            <div className="p-2 rounded-lg bg-purple-100">
              <Clock className="h-4 w-4 text-purple-600" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.withVersioning}</div>
            <p className="text-xs text-muted-foreground">
              Collections with history
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">With Auditing</CardTitle>
            <div className="p-2 rounded-lg bg-orange-100">
              <Shield className="h-4 w-4 text-orange-600" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.withAuditing}</div>
            <p className="text-xs text-muted-foreground">
              Tracking all changes
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Info Alert */}
      <Alert className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-900">
        <Info className="h-4 w-4 text-blue-600" />
        <AlertDescription className="text-blue-900 dark:text-blue-100">
          <strong>Pro Tip:</strong> Enable versioning to track document history, soft-delete to preserve data, and auditing for compliance requirements.
          Collections support advanced features like schema validation, indexes, and custom permissions.
        </AlertDescription>
      </Alert>

      {/* Collections Section Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold">All Collections</h2>
          <p className="text-sm text-muted-foreground">
            {collections.length > 0 
              ? `${collections.length} collection${collections.length === 1 ? '' : 's'} in your database`
              : 'Get started by creating your first collection'}
          </p>
        </div>
        {collections.length > 5 && (
          <Badge variant="outline">{collections.length} total</Badge>
        )}
      </div>

      {/* Collections Grid */}
      {collections.length > 0 ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {collections.map((collection: any) => (
            <div key={collection.name} className="group relative">
              <Link href={`/collections/${collection.name}`}>
                <div className="border-2 rounded-lg p-6 space-y-4 hover:shadow-lg hover:border-primary/50 transition-all cursor-pointer bg-card">
                  {/* Collection Header */}
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div className="p-2 rounded-lg bg-primary/10">
                        <Database className="h-5 w-5 text-primary" />
                      </div>
                      <div>
                        <h3 className="font-semibold text-lg">{collection.name}</h3>
                        <p className="text-sm text-muted-foreground line-clamp-1">
                          {collection.description || "No description provided"}
                        </p>
                      </div>
                    </div>
                  </div>

                  {/* Stats Row */}
                  <div className="flex items-center justify-between py-3 border-y">
                    <div className="flex items-center gap-2">
                      <FileJson className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">Documents</span>
                    </div>
                    <span className="font-semibold">{collection.document_count || 0}</span>
                  </div>

                  {/* Features */}
                  <div className="flex flex-wrap gap-2">
                    {collection.settings?.versioning && (
                      <Badge variant="secondary" className="text-xs">
                        <Clock className="h-3 w-3 mr-1" />
                        Versioning
                      </Badge>
                    )}
                    {collection.settings?.soft_delete && (
                      <Badge variant="secondary" className="text-xs">
                        <Shield className="h-3 w-3 mr-1" />
                        Soft Delete
                      </Badge>
                    )}
                    {collection.settings?.auditing && (
                      <Badge variant="secondary" className="text-xs">
                        <CheckCircle className="h-3 w-3 mr-1" />
                        Auditing
                      </Badge>
                    )}
                    {!collection.settings?.versioning && !collection.settings?.soft_delete && !collection.settings?.auditing && (
                      <Badge variant="outline" className="text-xs text-muted-foreground">
                        Standard Collection
                      </Badge>
                    )}
                  </div>

                  {/* Footer */}
                  {collection.created_at && (
                    <div className="flex items-center justify-between pt-3 text-xs text-muted-foreground">
                      <div className="flex items-center gap-1">
                        <Calendar className="h-3 w-3" />
                        Created {format(new Date(collection.created_at), 'MMM d, yyyy')}
                      </div>
                      <ArrowRight className="h-4 w-4 group-hover:translate-x-1 transition-transform" />
                    </div>
                  )}
                </div>
              </Link>
              
              {/* Delete Button - appears on hover */}
              <div className="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity">
                <DeleteCollectionButton collectionName={collection.name} />
              </div>
            </div>
          ))}
        </div>
      ) : (
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center justify-center py-12 space-y-4">
            <div className="p-4 rounded-full bg-muted">
              <Database className="h-12 w-12 text-muted-foreground" />
            </div>
            <div className="text-center space-y-2">
              <h3 className="text-lg font-semibold">No collections yet</h3>
              <p className="text-sm text-muted-foreground max-w-sm">
                Collections are containers for your documents. Create your first collection to start storing data.
              </p>
            </div>
            <CreateCollectionButton />
          </CardContent>
        </Card>
      )}

      {/* Quick Start Guide for new users */}
      {collections.length === 0 && (
        <Card className="border-2 border-dashed">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Info className="h-5 w-5" />
              Quick Start Guide
            </CardTitle>
            <CardDescription>Get up and running with collections in minutes</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-3">
              <div className="space-y-2">
                <div className="flex items-center gap-2 font-medium">
                  <Badge className="px-2">1</Badge>
                  Create a Collection
                </div>
                <p className="text-sm text-muted-foreground">
                  Define your data structure with a name and optional settings
                </p>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2 font-medium">
                  <Badge className="px-2">2</Badge>
                  Add Documents
                </div>
                <p className="text-sm text-muted-foreground">
                  Insert JSON documents with automatic validation
                </p>
              </div>
              <div className="space-y-2">
                <div className="flex items-center gap-2 font-medium">
                  <Badge className="px-2">3</Badge>
                  Query Your Data
                </div>
                <p className="text-sm text-muted-foreground">
                  Use powerful queries to retrieve and analyze your data
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}