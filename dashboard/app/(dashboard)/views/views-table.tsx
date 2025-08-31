"use client"

import { Badge } from "@/components/ui/badge";
import { Database, MoreVertical, Trash2, Edit, Eye, Copy, Calendar, Filter, Hash, FileCode } from "lucide-react";
import { formatDistanceToNow } from 'date-fns';
import { useRouter } from "next/navigation";
import { CreateViewDialog } from "./view-client-components";
import { ExecuteViewDialog } from "./execute-view-dialog";
import { EditViewDialog } from "./edit-view-dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { viewsApi } from "@/lib/api";
import { useToast } from "@/hooks/use-toast";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { useState, useTransition } from "react";
import { deleteView } from "@/app/actions/views-actions";

interface View {
  id?: string;
  name: string;
  description?: string;
  source_collection?: string;
  collection?: string;
  pipeline?: any[];
  filter?: any;
  fields?: string[];
  sort?: any;
  created_at?: string;
  updated_at?: string;
}

interface ViewsTableProps {
  views: View[];
  collections: any[];
}

export function ViewsTable({ views, collections }: ViewsTableProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [executeDialogOpen, setExecuteDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [selectedView, setSelectedView] = useState<View | null>(null);
  const [isPending, startTransition] = useTransition();
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    if (!selectedView) return;
    
    setIsDeleting(true);
    startTransition(async () => {
      try {
        const result = await deleteView(selectedView.name);
        if (result.success) {
          toast({
            title: "View deleted",
            description: `${selectedView.name} has been deleted successfully.`,
          });
          setDeleteDialogOpen(false);
          router.refresh();
        } else {
          toast({
            title: "Error",
            description: result.error || "Failed to delete view",
            variant: "destructive",
          });
        }
      } catch (error) {
        toast({
          title: "Error",
          description: "Failed to delete view. Please try again.",
          variant: "destructive",
        });
      } finally {
        setIsDeleting(false);
      }
    });
  };

  const copyQuery = (view: View) => {
    const query = {
      collection: view.source_collection || view.collection,
      pipeline: view.pipeline,
      filter: view.filter,
      fields: view.fields,
      sort: view.sort
    };
    navigator.clipboard.writeText(JSON.stringify(query, null, 2));
    toast({
      title: "Copied",
      description: "View query copied to clipboard",
    });
  };

  return (
    <>
    <div className="border">
      <table className="w-full">
        <thead>
          <tr className="border-b bg-muted/50">
            <th className="text-left p-4 font-medium text-sm">View Name</th>
            <th className="text-left p-4 font-medium text-sm">Collection</th>
            <th className="text-left p-4 font-medium text-sm">Type</th>
            <th className="text-right p-4 font-medium text-sm">Created</th>
            <th className="w-10"></th>
          </tr>
        </thead>
        <tbody>
          {views.length > 0 ? (
            views.map((view: View, index: number) => (
              <tr 
                key={view.id || view.name} 
                className={`border-b hover:bg-muted/30 transition-colors ${
                  index === views.length - 1 ? 'border-b-0' : ''
                }`}
              >
                <td className="p-4">
                  <div>
                    <p className="font-medium">{view.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {view.description || "No description provided"}
                    </p>
                  </div>
                </td>
                <td className="p-4">
                  <Badge variant="outline">
                    <Database className="h-3 w-3 mr-1.5" />
                    {view.source_collection || view.collection || "Unknown"}
                  </Badge>
                </td>
                <td className="p-4">
                  <div className="flex gap-2">
                    {view.pipeline && view.pipeline.length > 0 && (
                      <Badge variant="secondary" className="font-normal">
                        <Hash className="h-3 w-3 mr-1.5" />
                        Pipeline
                      </Badge>
                    )}
                    {view.filter && Object.keys(view.filter).length > 0 && (
                      <Badge variant="secondary" className="font-normal">
                        <Filter className="h-3 w-3 mr-1.5" />
                        Filtered
                      </Badge>
                    )}
                    {view.fields && view.fields.length > 0 && (
                      <Badge variant="secondary" className="font-normal">
                        Projection
                      </Badge>
                    )}
                    {!view.pipeline?.length && !view.filter && !view.fields?.length && (
                      <Badge variant="outline" className="font-normal text-muted-foreground">
                        Simple
                      </Badge>
                    )}
                  </div>
                </td>
                <td className="p-4 text-right">
                  <span className="text-sm text-muted-foreground">
                    {view.created_at ? formatDistanceToNow(new Date(view.created_at), { addSuffix: true }) : '-'}
                  </span>
                </td>
                <td className="p-4">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuLabel>Actions</DropdownMenuLabel>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={() => {
                        setSelectedView(view);
                        setExecuteDialogOpen(true);
                      }}>
                        <Eye className="mr-2 h-4 w-4" />
                        Execute View
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => copyQuery(view)}>
                        <Copy className="mr-2 h-4 w-4" />
                        Copy Query
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => {
                        setSelectedView(view);
                        setEditDialogOpen(true);
                      }}>
                        <Edit className="mr-2 h-4 w-4" />
                        Edit View
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem 
                        className="text-destructive"
                        onClick={() => {
                          setSelectedView(view);
                          setDeleteDialogOpen(true);
                        }}
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
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
                    <h3 className="text-lg font-semibold">No views yet</h3>
                    <p className="text-sm text-muted-foreground max-w-sm">
                      Views are saved queries that provide quick access to filtered and projected data from your collections.
                    </p>
                  </div>
                  <CreateViewDialog collections={collections} />
                </div>
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>

    {/* Execute View Dialog */}
    {selectedView && (
      <ExecuteViewDialog
        view={selectedView}
        open={executeDialogOpen}
        onOpenChange={setExecuteDialogOpen}
      />
    )}

    {/* Edit View Dialog */}
    {selectedView && (
      <EditViewDialog
        view={selectedView}
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
      />
    )}

    {/* Delete Confirmation Dialog */}
    <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete View</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete the view "{selectedView?.name}"? 
            This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting || isPending}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={isDeleting || isPending}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {isDeleting || isPending ? "Deleting..." : "Delete"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
    </>
  );
}