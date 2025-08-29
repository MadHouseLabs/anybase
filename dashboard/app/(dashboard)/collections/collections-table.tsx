"use client"

import { Badge } from "@/components/ui/badge";
import { Clock, Shield, Activity, Database } from "lucide-react";
import { format } from 'date-fns';
import { useRouter } from "next/navigation";
import { CreateCollectionButton } from "./collection-client-components";

interface CollectionsTableProps {
  collections: any[]
}

export function CollectionsTable({ collections }: CollectionsTableProps) {
  const router = useRouter();

  return (
    <div className="border">
      <table className="w-full">
        <thead>
          <tr className="border-b bg-muted/50">
            <th className="text-left p-4 font-medium text-sm">Collection</th>
            <th className="text-left p-4 font-medium text-sm">Features</th>
            <th className="text-right p-4 font-medium text-sm">Documents</th>
            <th className="text-right p-4 font-medium text-sm">Created</th>
            <th className="w-10"></th>
          </tr>
        </thead>
        <tbody>
          {collections.length > 0 ? (
            collections.map((collection: any, index: number) => (
              <tr 
                key={collection.name} 
                className={`border-b hover:bg-muted/30 transition-colors cursor-pointer ${
                  index === collections.length - 1 ? 'border-b-0' : ''
                }`}
                onClick={() => router.push(`/collections/${collection.name}`)}
              >
                <td className="p-4">
                  <div>
                    <p className="font-medium">{collection.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {collection.description || "No description provided"}
                    </p>
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex gap-2">
                    {collection.settings?.versioning && (
                      <Badge variant="secondary" className="font-normal">
                        <Clock className="h-3 w-3 mr-1.5" />
                        Versioning
                      </Badge>
                    )}
                    {collection.settings?.soft_delete && (
                      <Badge variant="secondary" className="font-normal">
                        <Shield className="h-3 w-3 mr-1.5" />
                        Soft Delete
                      </Badge>
                    )}
                    {collection.settings?.auditing && (
                      <Badge variant="secondary" className="font-normal">
                        <Shield className="h-3 w-3 mr-1.5" />
                        Auditing
                      </Badge>
                    )}
                    {!collection.settings?.versioning && !collection.settings?.soft_delete && !collection.settings?.auditing && (
                      <Badge variant="outline" className="font-normal text-muted-foreground">
                        Standard
                      </Badge>
                    )}
                  </div>
                </td>
                <td className="p-4 text-right">
                  <span className="font-mono font-medium">
                    {(collection.document_count || 0).toLocaleString()}
                  </span>
                </td>
                <td className="p-4 text-right">
                  <span className="text-sm text-muted-foreground">
                    {collection.created_at ? format(new Date(collection.created_at), 'MMM d, yyyy') : '-'}
                  </span>
                </td>
                <td className="p-4">
                  <Activity className="h-4 w-4 text-muted-foreground" />
                </td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan={5} className="p-12 text-center">
                <div className="flex flex-col items-center space-y-4">
                  <div className="p-4 rounded-full bg-muted">
                    <Database className="h-12 w-12 text-muted-foreground" />
                  </div>
                  <div className="space-y-2">
                    <h3 className="text-lg font-semibold">No collections yet</h3>
                    <p className="text-sm text-muted-foreground max-w-sm">
                      Collections are containers for your documents. Create your first collection to start storing data.
                    </p>
                  </div>
                  <CreateCollectionButton />
                </div>
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}