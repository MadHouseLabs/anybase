"use client"

import { useState, useTransition, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { useToast } from "@/hooks/use-toast";
import { updateView } from "@/app/actions/views-actions";
import { Loader2, Save } from "lucide-react";
import { useRouter } from "next/navigation";

interface View {
  id?: string;
  name: string;
  description?: string;
  source_collection?: string;
  collection?: string;
  pipeline?: any[];
  filter?: any;
  fields?: string[];
}

interface EditViewDialogProps {
  view: View;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function EditViewDialog({ view, open, onOpenChange }: EditViewDialogProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [isPending, startTransition] = useTransition();
  const [isUpdating, setIsUpdating] = useState(false);
  
  const [formData, setFormData] = useState({
    description: "",
    filter: "{}",
    fields: ""
  });

  useEffect(() => {
    if (view && open) {
      setFormData({
        description: view.description || "",
        filter: view.filter ? JSON.stringify(view.filter, null, 2) : "{}",
        fields: view.fields ? view.fields.join(", ") : ""
      });
    }
  }, [view, open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    setIsUpdating(true);
    startTransition(async () => {
      try {
        // Parse and validate JSON fields
        let filterObj = {};
        
        try {
          if (formData.filter.trim() && formData.filter !== "{}") {
            filterObj = JSON.parse(formData.filter);
          }
        } catch {
          toast({
            title: "Error",
            description: "Invalid filter JSON",
            variant: "destructive",
          });
          setIsUpdating(false);
          return;
        }

        // Parse fields
        const fieldsArray = formData.fields
          .split(",")
          .map(f => f.trim())
          .filter(f => f.length > 0);

        const updateData = {
          description: formData.description || undefined,
          filter: Object.keys(filterObj).length > 0 ? filterObj : undefined,
          fields: fieldsArray.length > 0 ? fieldsArray : undefined
        };

        const result = await updateView(view.name, updateData);
        
        if (result.success) {
          toast({
            title: "View updated",
            description: `${view.name} has been updated successfully.`,
          });
          onOpenChange(false);
          router.refresh();
        } else {
          toast({
            title: "Error",
            description: result.error || "Failed to update view",
            variant: "destructive",
          });
        }
      } catch (error) {
        toast({
          title: "Error",
          description: "Failed to update view",
          variant: "destructive",
        });
      } finally {
        setIsUpdating(false);
      }
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Edit View: {view?.name}</DialogTitle>
            <DialogDescription>
              Update the view configuration. The view name and source collection cannot be changed.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="Enter a description for this view"
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                className="min-h-[80px]"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="filter">Filter (JSON)</Label>
              <Textarea
                id="filter"
                placeholder='{"status": "active", "price": {"$gte": 100}}'
                value={formData.filter}
                onChange={(e) => setFormData(prev => ({ ...prev, filter: e.target.value }))}
                className="font-mono text-sm min-h-[100px]"
              />
              <p className="text-xs text-muted-foreground">
                MongoDB query filter to apply to the collection
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="fields">Projection Fields (comma-separated)</Label>
              <Input
                id="fields"
                placeholder="name, price, status"
                value={formData.fields}
                onChange={(e) => setFormData(prev => ({ ...prev, fields: e.target.value }))}
              />
              <p className="text-xs text-muted-foreground">
                Specify which fields to include in the results (leave empty for all fields)
              </p>
            </div>

            <div className="rounded-lg bg-muted p-3 space-y-1">
              <p className="text-sm font-medium">Source Collection</p>
              <p className="text-sm text-muted-foreground">
                {view?.source_collection || view?.collection} (cannot be changed)
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button 
              type="button" 
              variant="outline" 
              onClick={() => onOpenChange(false)}
              disabled={isUpdating || isPending}
            >
              Cancel
            </Button>
            <Button 
              type="submit" 
              disabled={isUpdating || isPending}
            >
              {isUpdating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Updating...
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Update View
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}