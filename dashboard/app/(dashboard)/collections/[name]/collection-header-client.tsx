"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { 
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Edit2, MoreVertical, Copy, Download, Trash2, RefreshCw, Save, X } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { useRouter } from "next/navigation";
import { collectionsApi } from "@/lib/api";

interface CollectionHeaderProps {
  collectionName: string;
  collection?: any;
}

export function CollectionHeader({ collectionName, collection }: CollectionHeaderProps) {
  const { toast } = useToast();
  const router = useRouter();
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  
  // Edit form states
  const [editData, setEditData] = useState({
    description: collection?.description || "",
    settings: {
      versioning: collection?.settings?.versioning || false,
      soft_delete: collection?.settings?.soft_delete || false,
      auditing: collection?.settings?.auditing || false,
    }
  });

  const handleRefresh = async () => {
    setIsRefreshing(true);
    router.refresh();
    setTimeout(() => {
      setIsRefreshing(false);
      toast({
        title: "Refreshed",
        description: "Collection data has been refreshed",
      });
    }, 500);
  };

  const handleExport = () => {
    toast({
      title: "Export Started",
      description: "Your collection export will be ready shortly",
    });
  };

  const handleDuplicate = () => {
    toast({
      title: "Duplicate Collection",
      description: "This feature will be available soon",
    });
  };

  const handleDelete = async () => {
    if (confirm(`Are you sure you want to delete the collection "${collectionName}"? This action cannot be undone.`)) {
      try {
        await collectionsApi.delete(collectionName);
        toast({
          title: "Success",
          description: "Collection deleted successfully",
        });
        router.push("/collections");
      } catch (error) {
        toast({
          title: "Error",
          description: "Failed to delete collection",
          variant: "destructive",
        });
      }
    }
  };

  const handleUpdate = async () => {
    try {
      setIsUpdating(true);
      await collectionsApi.update(collectionName, editData);
      
      toast({
        title: "Success",
        description: "Collection updated successfully",
      });
      
      setEditDialogOpen(false);
      router.refresh();
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update collection",
        variant: "destructive",
      });
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <Button
        variant="outline"
        size="icon"
        onClick={handleRefresh}
        disabled={isRefreshing}
      >
        <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
      </Button>
      
      <Button variant="outline" onClick={() => setEditDialogOpen(true)}>
        <Edit2 className="h-4 w-4 mr-2" />
        Edit
      </Button>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="icon">
            <MoreVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-56">
          <DropdownMenuItem onClick={handleExport}>
            <Download className="h-4 w-4 mr-2" />
            Export Collection
          </DropdownMenuItem>
          <DropdownMenuItem onClick={handleDuplicate}>
            <Copy className="h-4 w-4 mr-2" />
            Duplicate Collection
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem 
            className="text-destructive hover:bg-destructive/10"
            onClick={handleDelete}
          >
            <Trash2 className="h-4 w-4 mr-2" />
            Delete Collection
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Edit Collection Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit Collection</DialogTitle>
            <DialogDescription>
              Update collection settings and configuration
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-6">
            {/* Description */}
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="Describe what this collection stores..."
                value={editData.description}
                onChange={(e) => setEditData({ ...editData, description: e.target.value })}
                className="resize-none"
                rows={3}
              />
            </div>

            {/* Settings */}
            <div className="space-y-4">
              <Label>Collection Features</Label>
              
              <div className="space-y-4 rounded-lg border p-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="versioning" className="text-base font-medium">
                      Document Versioning
                    </Label>
                    <p className="text-sm text-muted-foreground">
                      Keep a complete history of all document changes
                    </p>
                  </div>
                  <Switch
                    id="versioning"
                    checked={editData.settings.versioning}
                    onCheckedChange={(checked) => 
                      setEditData({
                        ...editData,
                        settings: { ...editData.settings, versioning: checked }
                      })
                    }
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="soft-delete" className="text-base font-medium">
                      Soft Delete
                    </Label>
                    <p className="text-sm text-muted-foreground">
                      Mark documents as deleted instead of permanently removing them
                    </p>
                  </div>
                  <Switch
                    id="soft-delete"
                    checked={editData.settings.soft_delete}
                    onCheckedChange={(checked) => 
                      setEditData({
                        ...editData,
                        settings: { ...editData.settings, soft_delete: checked }
                      })
                    }
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="auditing" className="text-base font-medium">
                      Audit Logging
                    </Label>
                    <p className="text-sm text-muted-foreground">
                      Track all operations performed on documents for compliance
                    </p>
                  </div>
                  <Switch
                    id="auditing"
                    checked={editData.settings.auditing}
                    onCheckedChange={(checked) => 
                      setEditData({
                        ...editData,
                        settings: { ...editData.settings, auditing: checked }
                      })
                    }
                  />
                </div>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setEditDialogOpen(false);
                // Reset to original values
                setEditData({
                  description: collection?.description || "",
                  settings: {
                    versioning: collection?.settings?.versioning || false,
                    soft_delete: collection?.settings?.soft_delete || false,
                    auditing: collection?.settings?.auditing || false,
                  }
                });
              }}
              disabled={isUpdating}
            >
              <X className="h-4 w-4 mr-2" />
              Cancel
            </Button>
            <Button onClick={handleUpdate} disabled={isUpdating}>
              {isUpdating ? (
                <>Updating...</>
              ) : (
                <>
                  <Save className="h-4 w-4 mr-2" />
                  Save Changes
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}