"use client"

import { useState, useTransition } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";
import { queryView } from "@/app/actions/views-actions";
import { Play, Filter, SortAsc, ChevronLeft, ChevronRight, Eye, Copy, Loader2, Database, FileCode } from "lucide-react";
import { Textarea } from "@/components/ui/textarea";

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
}

interface ExecuteViewDialogProps {
  view: View;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ExecuteViewDialog({ view, open, onOpenChange }: ExecuteViewDialogProps) {
  const { toast } = useToast();
  const [isPending, startTransition] = useTransition();
  const [results, setResults] = useState<any[]>([]);
  const [totalCount, setTotalCount] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [additionalFilter, setAdditionalFilter] = useState("{}");
  const [sortField, setSortField] = useState("");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");
  const [activeTab, setActiveTab] = useState<"results" | "json" | "query">("results");
  const [isExecuting, setIsExecuting] = useState(false);

  const executeView = () => {
    setIsExecuting(true);
    startTransition(async () => {
      try {
        let additionalFilterObj = {};
        try {
          if (additionalFilter.trim() && additionalFilter !== "{}") {
            additionalFilterObj = JSON.parse(additionalFilter);
          }
        } catch {
          toast({
            title: "Error",
            description: "Invalid filter JSON",
            variant: "destructive",
          });
          setIsExecuting(false);
          return;
        }

        const params: any = {
          limit: pageSize,
          skip: (currentPage - 1) * pageSize,
        };

        // Combine view filter with additional filter
        if (view.filter || Object.keys(additionalFilterObj).length > 0) {
          params.filter = { ...view.filter, ...additionalFilterObj };
        }

        // Add sort
        if (sortField) {
          params.sort = { [sortField]: sortOrder === "asc" ? 1 : -1 };
        } else if (view.sort) {
          params.sort = view.sort;
        }

        const result = await queryView(view.name, params);
        
        if (result.success) {
          setResults(result.data || []);
          setTotalCount(result.total || result.data?.length || 0);
          
          if (result.data && result.data.length === 0) {
            toast({
              title: "No results",
              description: "The view query returned no documents",
            });
          }
        } else {
          toast({
            title: "Error",
            description: result.error || "Failed to execute view",
            variant: "destructive",
          });
          setResults([]);
        }
      } catch (error) {
        toast({
          title: "Error",
          description: "Failed to execute view",
          variant: "destructive",
        });
        setResults([]);
      } finally {
        setIsExecuting(false);
      }
    });
  };

  const copyResults = () => {
    navigator.clipboard.writeText(JSON.stringify(results, null, 2));
    toast({
      title: "Copied",
      description: "Results copied to clipboard",
    });
  };

  const copyQuery = () => {
    const query = {
      view: view.name,
      collection: view.source_collection || view.collection,
      pipeline: view.pipeline,
      filter: { ...view.filter, ...JSON.parse(additionalFilter || "{}") },
      fields: view.fields,
      sort: sortField ? { [sortField]: sortOrder === "asc" ? 1 : -1 } : view.sort,
      limit: pageSize,
      skip: (currentPage - 1) * pageSize,
    };
    navigator.clipboard.writeText(JSON.stringify(query, null, 2));
    toast({
      title: "Copied",
      description: "Query copied to clipboard",
    });
  };

  const totalPages = Math.ceil(totalCount / pageSize);

  const formatValue = (value: any): string => {
    if (value === null || value === undefined) return "null";
    if (typeof value === "object") return JSON.stringify(value);
    return String(value);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-5xl max-h-[80vh]">
        <DialogHeader>
          <DialogTitle>Execute View: {view.name}</DialogTitle>
          <DialogDescription>
            {view.description || "Run the view query with optional filters and pagination"}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Query Parameters */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>Additional Filter (JSON)</Label>
              <Textarea
                placeholder='{"status": "active"}'
                value={additionalFilter}
                onChange={(e) => setAdditionalFilter(e.target.value)}
                className="font-mono text-sm h-20"
              />
              <p className="text-xs text-muted-foreground">
                Will be combined with view's existing filter
              </p>
            </div>
            
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-2">
                <div className="space-y-2">
                  <Label>Sort Field</Label>
                  <Input
                    placeholder="e.g., price"
                    value={sortField}
                    onChange={(e) => setSortField(e.target.value)}
                  />
                </div>
                <div className="space-y-2">
                  <Label>Sort Order</Label>
                  <Select value={sortOrder} onValueChange={(v) => setSortOrder(v as "asc" | "desc")}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="asc">Ascending</SelectItem>
                      <SelectItem value="desc">Descending</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
              
              <div className="grid grid-cols-2 gap-2">
                <div className="space-y-2">
                  <Label>Page Size</Label>
                  <Select value={pageSize.toString()} onValueChange={(v) => setPageSize(parseInt(v))}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="10">10</SelectItem>
                      <SelectItem value="25">25</SelectItem>
                      <SelectItem value="50">50</SelectItem>
                      <SelectItem value="100">100</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex items-end">
                  <Button 
                    onClick={executeView} 
                    disabled={isPending || isExecuting}
                    className="w-full"
                  >
                    {isExecuting ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Executing...
                      </>
                    ) : (
                      <>
                        <Play className="mr-2 h-4 w-4" />
                        Execute
                      </>
                    )}
                  </Button>
                </div>
              </div>
            </div>
          </div>

          {/* Results Section */}
          {results.length > 0 && (
            <>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Badge variant="secondary">
                    {totalCount} total results
                  </Badge>
                  <Badge variant="outline">
                    Page {currentPage} of {totalPages}
                  </Badge>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                    disabled={currentPage === 1 || isPending}
                  >
                    <ChevronLeft className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                    disabled={currentPage === totalPages || isPending}
                  >
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>

              <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as any)}>
                <div className="flex items-center justify-between">
                  <TabsList>
                    <TabsTrigger value="results">
                      <Eye className="mr-2 h-4 w-4" />
                      Table View
                    </TabsTrigger>
                    <TabsTrigger value="json">
                      <FileCode className="mr-2 h-4 w-4" />
                      JSON View
                    </TabsTrigger>
                    <TabsTrigger value="query">
                      <Database className="mr-2 h-4 w-4" />
                      Query Details
                    </TabsTrigger>
                  </TabsList>
                  
                  <Button variant="outline" size="sm" onClick={copyResults}>
                    <Copy className="mr-2 h-4 w-4" />
                    Copy Results
                  </Button>
                </div>

                <TabsContent value="results" className="mt-4">
                  <ScrollArea className="h-[300px] w-full border rounded-md">
                    <table className="w-full">
                      <thead className="sticky top-0 bg-background border-b">
                        <tr>
                          <th className="text-left p-3 font-medium text-sm w-[200px]">ID</th>
                          <th className="text-left p-3 font-medium text-sm">Data</th>
                        </tr>
                      </thead>
                      <tbody>
                        {results.map((doc, idx) => {
                          const { _id, ...restData } = doc;
                          const dataString = JSON.stringify(restData);
                          const truncatedData = dataString.length > 100 
                            ? dataString.substring(0, 100) + "..." 
                            : dataString;
                          
                          return (
                            <tr key={idx} className="border-b hover:bg-muted/50">
                              <td className="p-3 text-sm">
                                <span className="font-mono text-xs text-muted-foreground">
                                  {_id || "N/A"}
                                </span>
                              </td>
                              <td className="p-3 text-sm">
                                <span className="font-mono text-xs" title={dataString}>
                                  {truncatedData}
                                </span>
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </ScrollArea>
                </TabsContent>

                <TabsContent value="json" className="mt-4">
                  <ScrollArea className="h-[300px] w-full border rounded-md p-4">
                    <pre className="text-sm">
                      <code>{JSON.stringify(results, null, 2)}</code>
                    </pre>
                  </ScrollArea>
                </TabsContent>

                <TabsContent value="query" className="mt-4">
                  <ScrollArea className="h-[300px] w-full border rounded-md p-4">
                    <div className="space-y-4">
                      <div>
                        <h4 className="text-sm font-medium mb-2">Collection</h4>
                        <Badge variant="outline">
                          {view.source_collection || view.collection}
                        </Badge>
                      </div>
                      
                      {view.pipeline && view.pipeline.length > 0 && (
                        <div>
                          <h4 className="text-sm font-medium mb-2">Pipeline</h4>
                          <pre className="text-xs bg-muted p-2 rounded">
                            <code>{JSON.stringify(view.pipeline, null, 2)}</code>
                          </pre>
                        </div>
                      )}
                      
                      {(view.filter || additionalFilter !== "{}") && (
                        <div>
                          <h4 className="text-sm font-medium mb-2">Combined Filter</h4>
                          <pre className="text-xs bg-muted p-2 rounded">
                            <code>{JSON.stringify({ ...view.filter, ...JSON.parse(additionalFilter || "{}") }, null, 2)}</code>
                          </pre>
                        </div>
                      )}
                      
                      {view.fields && view.fields.length > 0 && (
                        <div>
                          <h4 className="text-sm font-medium mb-2">Projection Fields</h4>
                          <div className="flex flex-wrap gap-1">
                            {view.fields.map(field => (
                              <Badge key={field} variant="secondary" className="text-xs">
                                {field}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      )}
                      
                      <Button variant="outline" size="sm" onClick={copyQuery} className="w-full">
                        <Copy className="mr-2 h-4 w-4" />
                        Copy Full Query
                      </Button>
                    </div>
                  </ScrollArea>
                </TabsContent>
              </Tabs>
            </>
          )}
          
          {!results.length && !isExecuting && (
            <div className="text-center py-8 border rounded-md">
              <Database className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-sm text-muted-foreground">
                Click Execute to run the view query
              </p>
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}