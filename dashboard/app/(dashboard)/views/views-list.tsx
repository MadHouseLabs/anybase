"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";
import { ViewCardDisplay } from "./view-card-server";
import { ExecuteViewButton, EditViewButton, CopyViewQueryButton, DeleteViewButton } from "./view-client-components";
import { ViewsEmptyState } from "./views-empty-state";

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
          <ViewsEmptyState hasSearchQuery={viewsArray.length > 0 && filteredViews.length === 0} />
        ) : (
          filteredViews.map((view) => (
            <ViewCardDisplay key={view.id || view.name} view={view}>
              <ExecuteViewButton view={view} />
              <EditViewButton view={view} collections={collections} />
              <CopyViewQueryButton view={view} />
              <DeleteViewButton viewName={view.name} />
            </ViewCardDisplay>
          ))
        )}
      </div>
    </>
  );
}