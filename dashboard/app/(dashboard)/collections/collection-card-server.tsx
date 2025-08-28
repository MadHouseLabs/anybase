import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Database } from "lucide-react";

interface Collection {
  name: string;
  description?: string;
  settings?: {
    versioning?: boolean;
    soft_delete?: boolean;
    auditing?: boolean;
  };
  document_count?: number;
}

interface CollectionCardDisplayProps {
  collection: Collection;
  children?: React.ReactNode;
}

export function CollectionCardDisplay({ collection, children }: CollectionCardDisplayProps) {
  return (
    <Card className="group hover:shadow-lg transition-shadow cursor-pointer">
      <CardHeader>
        <div className="flex items-center justify-between">
          <Database className="h-8 w-8 text-muted-foreground" />
          <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            {children}
          </div>
        </div>
        <CardTitle>{collection.name}</CardTitle>
        <CardDescription>{collection.description || "No description"}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            {collection.settings?.versioning && (
              <Badge variant="secondary">Versioning</Badge>
            )}
            {collection.settings?.soft_delete && (
              <Badge variant="secondary">Soft Delete</Badge>
            )}
            {collection.settings?.auditing && (
              <Badge variant="secondary">Auditing</Badge>
            )}
          </div>
          {collection.document_count !== undefined && (
            <p className="text-sm text-muted-foreground">
              {collection.document_count} documents
            </p>
          )}
        </div>
      </CardContent>
    </Card>
  );
}