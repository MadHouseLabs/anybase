import { Eye } from "lucide-react";

interface ViewsEmptyStateProps {
  hasSearchQuery: boolean;
}

export function ViewsEmptyState({ hasSearchQuery }: ViewsEmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12">
      <Eye className="h-12 w-12 text-muted-foreground mb-4" />
      <p className="text-lg font-medium mb-2">
        {hasSearchQuery ? "No views found" : "No views yet"}
      </p>
      <p className="text-sm text-muted-foreground">
        {hasSearchQuery ? "Try adjusting your search" : "Create your first view to get started"}
      </p>
    </div>
  );
}