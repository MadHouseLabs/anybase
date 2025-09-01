"use client"

import { Badge } from "@/components/ui/badge";
import { Clock, Shield, Activity, Database, MoreVertical, Trash2, Edit, Eye } from "lucide-react";
import { format } from 'date-fns';
import { formatLocalDateOnly } from '@/lib/date-utils';
import { useRouter } from "next/navigation";
import { CreateCollectionButton } from "./collection-client-components";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { collectionsApi } from "@/lib/api";
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
import { useState } from "react";

interface CollectionsTableProps {
  collections: any[]
}

export function CollectionsTable({ collections }: CollectionsTableProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedCollection, setSelectedCollection] = useState<any>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    if (!selectedCollection) return;
    
    setIsDeleting(true);
    try {
      await collectionsApi.delete(selectedCollection.name);
      toast({
        title: "Collection deleted",
        description: `${selectedCollection.name} has been deleted successfully.`,
      });
      setDeleteDialogOpen(false);
      router.refresh();
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete collection. Please try again.",
        variant: "destructive",
      });
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <>
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
                key={collection.id || `${collection.name}-${index}`} 
                className={`border-b hover:bg-muted/30 transition-colors ${
                  index === collections.length - 1 ? 'border-b-0' : ''
                }`}
              >
                <td className="p-4 cursor-pointer" onClick={() => router.push(`/collections/${collection.name}`)}>
                  <div>
                    <p className="font-medium">{collection.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {collection.description || "No description provided"}
                    </p>
                  </div>
                </td>
                <td className="p-4 cursor-pointer" onClick={() => router.push(`/collections/${collection.name}`)}>
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
                <td className="p-4 text-right cursor-pointer" onClick={() => router.push(`/collections/${collection.name}`)}>
                  <span className="font-mono font-medium">
                    {(collection.document_count || 0).toLocaleString()}
                  </span>
                </td>
                <td className="p-4 text-right cursor-pointer" onClick={() => router.push(`/collections/${collection.name}`)}>
                  <span className="text-sm text-muted-foreground">
                    {formatLocalDateOnly(collection.created_at) || '-'}
                  </span>
                </td>
                <td className="p-4">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                      <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuLabel>Actions</DropdownMenuLabel>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={() => router.push(`/collections/${collection.name}`)}>
                        <Eye className="mr-2 h-4 w-4" />
                        View Details
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => router.push(`/collections/${collection.name}`)}>
                        <Edit className="mr-2 h-4 w-4" />
                        Edit Settings
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem 
                        className="text-destructive"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedCollection(collection);
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

    {/* Delete Confirmation Dialog */}
    <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Collection</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete the collection "{selectedCollection?.name}"? 
            This action cannot be undone and will permanently delete all documents in this collection.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={isDeleting}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {isDeleting ? "Deleting..." : "Delete"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
    </>
  );
}