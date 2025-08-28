"use client";

import { useState, useTransition } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";
import { createView, updateView, deleteView, queryView } from "@/app/actions/views-actions";
import { Plus, Play, Edit, Trash2, Copy, CheckCircle, Database, FileJson } from "lucide-react";

interface Collection {
  name: string;
  description?: string;
}

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

export function CreateViewDialog({ collections }: { collections: Collection[] }) {
  const { toast } = useToast();
  const [open, setOpen] = useState(false);
  const [isPending, startTransition] = useTransition();
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    collection: "",
    filter: "{}",
    fields: ""
  });

  const handleSubmit = () => {
    if (!formData.name || !formData.collection) {
      toast({
        title: "Error",
        description: "View name and collection are required",
        variant: "destructive",
      });
      return;
    }

    startTransition(async () => {
      try {
        let filter = {};
        let fields: string[] = [];

        try {
          if (formData.filter.trim()) {
            filter = JSON.parse(formData.filter);
          }
        } catch {
          toast({
            title: "Error",
            description: "Invalid filter JSON",
            variant: "destructive",
          });
          return;
        }

        if (formData.fields.trim()) {
          fields = formData.fields.split(',').map(f => f.trim()).filter(f => f);
        }

        const result = await createView({
          name: formData.name,
          description: formData.description,
          collection: formData.collection,
          filter: Object.keys(filter).length > 0 ? filter : undefined,
          fields: fields.length > 0 ? fields : undefined,
        });

        if (result.success) {
          toast({
            title: "Success",
            description: `View "${formData.name}" created successfully`,
          });
          setOpen(false);
          setFormData({
            name: "",
            description: "",
            collection: "",
            filter: "{}",
            fields: ""
          });
        } else {
          toast({
            title: "Error",
            description: result.error || "Failed to create view",
            variant: "destructive",
          });
        }
      } catch (error) {
        toast({
          title: "Error",
          description: "Failed to create view",
          variant: "destructive",
        });
      }
    });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm">
          <Plus className="h-4 w-4 mr-2" />
          Create View
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Create View</DialogTitle>
          <DialogDescription>
            Create a saved query that can be executed on demand
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="name">View Name</Label>
              <Input
                id="name"
                placeholder="e.g., active_users"
                value={formData.name}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="collection">Collection</Label>
              <Select
                value={formData.collection}
                onValueChange={(value) => setFormData(prev => ({ ...prev, collection: value }))}
              >
                <SelectTrigger id="collection">
                  <SelectValue placeholder="Select collection" />
                </SelectTrigger>
                <SelectContent>
                  {collections.map((col) => (
                    <SelectItem key={col.name} value={col.name}>
                      {col.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Describe what this view returns"
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
            />
          </div>

          <Tabs defaultValue="filter" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="filter">Filter</TabsTrigger>
              <TabsTrigger value="fields">Fields (Projection)</TabsTrigger>
            </TabsList>
            
            <TabsContent value="filter" className="space-y-2">
              <Label htmlFor="filter">Query Filter (JSON)</Label>
              <Textarea
                id="filter"
                placeholder='{"status": "active", "age": {"$gte": 18}}'
                className="font-mono text-sm"
                rows={4}
                value={formData.filter}
                onChange={(e) => setFormData(prev => ({ ...prev, filter: e.target.value }))}
              />
              <p className="text-xs text-muted-foreground">
                MongoDB-style query filter to apply to the collection
              </p>
            </TabsContent>
            
            <TabsContent value="fields" className="space-y-2">
              <Label htmlFor="fields">Projected Fields</Label>
              <Input
                id="fields"
                placeholder="name, email, created_at"
                value={formData.fields}
                onChange={(e) => setFormData(prev => ({ ...prev, fields: e.target.value }))}
              />
              <p className="text-xs text-muted-foreground">
                Comma-separated list of fields to include in results (leave empty for all fields)
              </p>
            </TabsContent>
          </Tabs>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)} disabled={isPending}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isPending}>
            {isPending ? "Creating..." : "Create View"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function EditViewButton({ view, collections }: { view: View; collections: Collection[] }) {
  const { toast } = useToast();
  const [open, setOpen] = useState(false);
  const [isPending, startTransition] = useTransition();
  const [formData, setFormData] = useState({
    name: view.name,
    description: view.description || "",
    collection: view.collection,
    filter: JSON.stringify(view.filter || {}),
    fields: view.fields ? view.fields.join(", ") : ""
  });

  const handleSubmit = () => {
    startTransition(async () => {
      try {
        let filter = {};
        let fields: string[] = [];
        
        if (formData.filter.trim() && formData.filter !== "{}") {
          filter = JSON.parse(formData.filter);
        }
        
        if (formData.fields.trim()) {
          fields = formData.fields.split(',').map(f => f.trim()).filter(f => f);
        }
        
        const result = await updateView(view.name, {
          description: formData.description,
          filter,
          fields: fields.length > 0 ? fields : undefined
        });
        
        if (result.success) {
          toast({
            title: "Success",
            description: `View "${view.name}" updated successfully`,
          });
          setOpen(false);
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
      }
    });
  };

  return (
    <>
      <Button
        variant="ghost"
        size="sm"
        onClick={() => setOpen(true)}
      >
        <Edit className="h-4 w-4" />
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit View: {view.name}</DialogTitle>
            <DialogDescription>
              Update the view configuration
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>View Name</Label>
                <Input
                  value={formData.name}
                  disabled
                  className="bg-muted"
                />
                <p className="text-xs text-muted-foreground">
                  View name cannot be changed
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-collection">Collection</Label>
                <Select
                  value={formData.collection}
                  onValueChange={(value) => setFormData(prev => ({ ...prev, collection: value }))}
                >
                  <SelectTrigger id="edit-collection">
                    <SelectValue placeholder="Select collection" />
                  </SelectTrigger>
                  <SelectContent>
                    {collections.map((col) => (
                      <SelectItem key={col.name} value={col.name}>
                        <div className="flex items-center gap-2">
                          <Database className="h-3 w-3" />
                          {col.name}
                        </div>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="edit-description">Description</Label>
              <Textarea
                id="edit-description"
                placeholder="Describe what this view returns"
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              />
            </div>

            <Tabs defaultValue="filter" className="w-full">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="filter">Filter</TabsTrigger>
                <TabsTrigger value="fields">Fields (Projection)</TabsTrigger>
              </TabsList>
              
              <TabsContent value="filter" className="space-y-2">
                <Label htmlFor="edit-filter">Query Filter (JSON)</Label>
                <Textarea
                  id="edit-filter"
                  placeholder='{"status": "active", "age": {"$gte": 18}}'
                  className="font-mono text-sm"
                  rows={4}
                  value={formData.filter}
                  onChange={(e) => setFormData(prev => ({ ...prev, filter: e.target.value }))}
                />
                <p className="text-xs text-muted-foreground">
                  MongoDB-style query filter to apply to the collection
                </p>
              </TabsContent>
              
              <TabsContent value="fields" className="space-y-2">
                <Label htmlFor="edit-fields">Projected Fields</Label>
                <Input
                  id="edit-fields"
                  placeholder="name, email, created_at"
                  value={formData.fields}
                  onChange={(e) => setFormData(prev => ({ ...prev, fields: e.target.value }))}
                />
                <p className="text-xs text-muted-foreground">
                  Comma-separated list of fields to include in results (leave empty for all fields)
                </p>
              </TabsContent>
            </Tabs>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)} disabled={isPending}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={isPending}>
              {isPending ? "Updating..." : "Update View"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

export function DeleteViewButton({ viewName }: { viewName: string }) {
  const { toast } = useToast();
  const [isPending, startTransition] = useTransition();

  const handleDelete = () => {
    if (!confirm(`Are you sure you want to delete the view "${viewName}"?`)) return;

    startTransition(async () => {
      const result = await deleteView(viewName);
      
      if (result.success) {
        toast({
          title: "Success",
          description: `View "${viewName}" deleted successfully`,
        });
      } else {
        toast({
          title: "Error",
          description: result.error || "Failed to delete view",
          variant: "destructive",
        });
      }
    });
  };

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={handleDelete}
      disabled={isPending}
      className="hover:bg-red-50"
    >
      <Trash2 className="h-4 w-4 text-destructive" />
    </Button>
  );
}

export function ExecuteViewButton({ view }: { view: View }) {
  const { toast } = useToast();
  const [open, setOpen] = useState(false);
  const [isPending, startTransition] = useTransition();
  const [queryResults, setQueryResults] = useState<any[]>([]);
  const [queryOptions, setQueryOptions] = useState({
    filter: "{}",
    sort: "{}",
    limit: "10",
    skip: "0"
  });

  const handleExecute = () => {
    startTransition(async () => {
      try {
        let extraFilter = {};
        let sort = {};
        let limit = 10;
        let skip = 0;
        
        try {
          if (queryOptions.filter.trim() && queryOptions.filter !== "{}") {
            extraFilter = JSON.parse(queryOptions.filter);
          }
          if (queryOptions.sort.trim() && queryOptions.sort !== "{}") {
            sort = JSON.parse(queryOptions.sort);
          }
          limit = parseInt(queryOptions.limit) || 10;
          skip = parseInt(queryOptions.skip) || 0;
        } catch (e) {
          toast({
            title: "Error",
            description: "Invalid JSON in filter or sort",
            variant: "destructive",
          });
          return;
        }
        
        const result = await queryView(view.name, { 
          limit,
          skip,
          filter: Object.keys(extraFilter).length > 0 ? extraFilter : undefined,
          sort: Object.keys(sort).length > 0 ? sort : undefined
        });
        
        if (result.success) {
          setQueryResults(result.data?.documents || result.data?.data || []);
        } else {
          toast({
            title: "Error",
            description: result.error || "Failed to execute view",
            variant: "destructive",
          });
        }
      } catch (error) {
        toast({
          title: "Error",
          description: "Failed to execute view",
          variant: "destructive",
        });
      }
    });
  };

  return (
    <>
      <Button
        variant="default"
        size="sm"
        onClick={() => setOpen(true)}
      >
        <Play className="h-4 w-4 mr-1" />
        Execute
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Execute View: {view.name}</DialogTitle>
            <DialogDescription>
              Query {view.collection} collection with runtime filters and pagination
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="border rounded-lg p-4 space-y-4 bg-muted/30">
              <h3 className="text-sm font-semibold">Runtime Query Options</h3>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="runtime-filter" className="text-xs">Additional Filter (JSON)</Label>
                  <Textarea
                    id="runtime-filter"
                    placeholder='{"status": "pending"}'
                    className="font-mono text-xs"
                    rows={3}
                    value={queryOptions.filter}
                    onChange={(e) => setQueryOptions(prev => ({ ...prev, filter: e.target.value }))}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="runtime-sort" className="text-xs">Sort Order (JSON)</Label>
                  <Textarea
                    id="runtime-sort"
                    placeholder='{"created_at": -1}'
                    className="font-mono text-xs"
                    rows={3}
                    value={queryOptions.sort}
                    onChange={(e) => setQueryOptions(prev => ({ ...prev, sort: e.target.value }))}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="runtime-limit" className="text-xs">Limit</Label>
                  <Input
                    id="runtime-limit"
                    type="number"
                    placeholder="10"
                    value={queryOptions.limit}
                    onChange={(e) => setQueryOptions(prev => ({ ...prev, limit: e.target.value }))}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="runtime-skip" className="text-xs">Skip (Offset)</Label>
                  <Input
                    id="runtime-skip"
                    type="number"
                    placeholder="0"
                    value={queryOptions.skip}
                    onChange={(e) => setQueryOptions(prev => ({ ...prev, skip: e.target.value }))}
                  />
                </div>
              </div>
              <Button 
                onClick={handleExecute}
                className="w-full"
                disabled={isPending}
              >
                <Play className="h-4 w-4 mr-2" />
                {isPending ? "Executing..." : "Execute Query"}
              </Button>
            </div>
            
            {queryResults.length > 0 ? (
              <>
                <div className="flex items-center justify-between">
                  <Badge variant="outline">
                    <FileJson className="h-3 w-3 mr-1" />
                    {queryResults.length} documents
                  </Badge>
                </div>
                <div className="border rounded-lg overflow-hidden">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        {queryResults[0] && Object.keys(queryResults[0]).map((key) => (
                          <TableHead key={key} className="font-mono text-xs">
                            {key}
                          </TableHead>
                        ))}
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {queryResults.map((doc, index) => (
                        <TableRow key={index}>
                          {Object.values(doc).map((value: any, i) => (
                            <TableCell key={i} className="font-mono text-xs">
                              {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                            </TableCell>
                          ))}
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                No results found
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}

export function CopyViewQueryButton({ view }: { view: View }) {
  const { toast } = useToast();
  const [copied, setCopied] = useState(false);

  const generateQueryString = () => {
    const parts = [];
    if (view.filter) parts.push(`filter: ${JSON.stringify(view.filter)}`);
    if (view.fields?.length) parts.push(`fields: [${view.fields.join(', ')}]`);
    if (view.sort) parts.push(`sort: ${JSON.stringify(view.sort)}`);
    return parts.join('\n');
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(generateQueryString());
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
    toast({
      title: "Copied",
      description: "Query copied to clipboard",
    });
  };

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={handleCopy}
    >
      {copied ? (
        <CheckCircle className="h-4 w-4" />
      ) : (
        <Copy className="h-4 w-4" />
      )}
    </Button>
  );
}