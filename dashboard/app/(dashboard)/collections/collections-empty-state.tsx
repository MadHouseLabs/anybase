import { Card, CardContent } from "@/components/ui/card";
import { Database } from "lucide-react";

export function CollectionsEmptyState() {
  return (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Database className="h-12 w-12 text-muted-foreground mb-4" />
        <p className="text-lg font-medium">No collections yet</p>
        <p className="text-sm text-muted-foreground">Create your first collection to get started</p>
      </CardContent>
    </Card>
  );
}