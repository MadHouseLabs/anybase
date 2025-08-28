"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";
import { ExecuteViewButton, EditViewButton, CopyViewQueryButton, DeleteViewButton } from "./view-client-components";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Database, Calendar } from "lucide-react";
import { format } from "date-fns";

interface View {
  id?: string;
  name: string;
  description?: string;
  collection: string;
  filter?: any;
  fields?: string[];
  sort?: any;
  created_at?: string;
}

interface Collection {
  name: string;
  description?: string;
}

interface ViewsListWithSearchProps {
  views: View[];
  collections: Collection[];
}

export function ViewsListWithSearch({ views, collections }: ViewsListWithSearchProps) {
  const [searchQuery, setSearchQuery] = useState("");

  // Ensure views is an array
  const viewsArray = Array.isArray(views) ? views : [];

  const filteredViews = viewsArray.filter(view => 
    searchQuery === "" || 
    view.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    view.collection?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    view.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <>
      <div className="relative">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
        <Input
          placeholder="Search views by name, collection, or description..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-10"
        />
      </div>
      <div className="space-y-4 mt-4">
        {filteredViews.length === 0 ? (
          <div className="text-center py-12">
            <Database className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-semibold mb-2">
              {viewsArray.length > 0 && filteredViews.length === 0 ? "No matching views" : "No views yet"}
            </h3>
            <p className="text-muted-foreground">
              {viewsArray.length > 0 && filteredViews.length === 0 
                ? "Try adjusting your search criteria"
                : "Create your first view to get started"}
            </p>
          </div>
        ) : (
          filteredViews.map((view) => (
            <Card key={view.id || view.name} className="hover:shadow-md transition-shadow">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <CardTitle className="flex items-center gap-2">
                      <Database className="h-5 w-5" />
                      {view.name}
                    </CardTitle>
                    <CardDescription>
                      {view.description || "No description"}
                    </CardDescription>
                  </div>
                  <Badge variant="outline">{view.collection}</Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {view.filter && (
                    <div className="text-sm">
                      <span className="font-medium">Filter:</span>
                      <code className="ml-2 text-xs bg-muted px-2 py-1 rounded">
                        {JSON.stringify(view.filter)}
                      </code>
                    </div>
                  )}
                  {view.fields && view.fields.length > 0 && (
                    <div className="text-sm">
                      <span className="font-medium">Fields:</span>
                      <span className="ml-2">{view.fields.join(", ")}</span>
                    </div>
                  )}
                  {view.created_at && (
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Calendar className="h-4 w-4" />
                      Created {format(new Date(view.created_at), "MMM d, yyyy")}
                    </div>
                  )}
                  <div className="flex gap-2 pt-2">
                    <ExecuteViewButton view={view} />
                    <EditViewButton view={view} collections={collections} />
                    <CopyViewQueryButton view={view} />
                    <DeleteViewButton viewName={view.name} />
                  </div>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </>
  );
}