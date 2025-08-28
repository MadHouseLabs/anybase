import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Eye, Database, Calendar } from "lucide-react";
import { format } from 'date-fns';

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

interface ViewCardDisplayProps {
  view: View;
  children?: React.ReactNode;
}

export function ViewCardDisplay({ view, children }: ViewCardDisplayProps) {
  return (
    <div className="border rounded-lg p-6 space-y-4 hover:bg-muted/50 transition-colors">
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary/10 rounded-lg">
              <Eye className="h-5 w-5 text-primary" />
            </div>
            <div>
              <h3 className="text-lg font-semibold">{view.name}</h3>
              <div className="flex items-center gap-2 mt-1">
                <Badge variant="outline" className="text-xs">
                  <Database className="h-3 w-3 mr-1" />
                  {view.collection}
                </Badge>
                {view.created_at && (
                  <span className="text-xs text-muted-foreground flex items-center gap-1">
                    <Calendar className="h-3 w-3" />
                    Created {format(new Date(view.created_at), 'MMM d, yyyy')}
                  </span>
                )}
              </div>
            </div>
          </div>
          {view.description && (
            <p className="text-sm text-muted-foreground mt-2">{view.description}</p>
          )}
        </div>
        <div className="flex gap-2">
          {children}
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4 pt-4 border-t">
        <div>
          <Label className="text-xs text-muted-foreground">Filter</Label>
          <code className="text-xs block mt-1 p-2 bg-muted rounded">
            {view.filter ? JSON.stringify(view.filter) : 'No filter'}
          </code>
        </div>
        <div>
          <Label className="text-xs text-muted-foreground">Fields</Label>
          <code className="text-xs block mt-1 p-2 bg-muted rounded">
            {view.fields?.length ? view.fields.join(', ') : 'All fields'}
          </code>
        </div>
        <div>
          <Label className="text-xs text-muted-foreground">Sort</Label>
          <code className="text-xs block mt-1 p-2 bg-muted rounded">
            {view.sort ? JSON.stringify(view.sort) : 'No sorting'}
          </code>
        </div>
      </div>
    </div>
  );
}