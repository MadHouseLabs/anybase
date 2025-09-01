"use client";

import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { 
  FileJson, Key, Settings, FileCode, Search, Plus, Copy,
  MoreHorizontal, Eye, Edit, Trash2, Filter, Download,
  ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight,
  Info, Shield, Clock, Activity, Database, Layers, Cpu
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useToast } from "@/hooks/use-toast";
import { format } from 'date-fns';
import { documentsApi, collectionsApi } from "@/lib/api";
import { VectorFieldsManager } from "@/components/vector-fields-manager";
import { formatLocalDateTime, formatLocalDate } from "@/lib/date-utils";

interface CollectionTabsProps {
  collectionName: string;
  collection: any;
  documents: any[];
  totalDocuments: number;
  indexes: any[];
}

export function CollectionTabs({ 
  collectionName, 
  collection, 
  documents: initialDocuments, 
  totalDocuments,
  indexes 
}: CollectionTabsProps) {
  const { toast } = useToast();
  const [activeTab, setActiveTab] = useState("documents");
  const [searchQuery, setSearchQuery] = useState("");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  
  // Dialog states
  const [documentDialogOpen, setDocumentDialogOpen] = useState(false);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [schemaDialogOpen, setSchemaDialogOpen] = useState(false);
  const [newDocumentContent, setNewDocumentContent] = useState("");
  const [editDocumentContent, setEditDocumentContent] = useState("");
  const [editSchemaContent, setEditSchemaContent] = useState("");
  const [selectedDocument, setSelectedDocument] = useState<any>(null);
  const [isCreatingDocument, setIsCreatingDocument] = useState(false);
  const [isUpdatingDocument, setIsUpdatingDocument] = useState(false);
  const [isDeletingDocument, setIsDeletingDocument] = useState(false);
  const [isUpdatingSchema, setIsUpdatingSchema] = useState(false);
  
  // Filter documents based on search
  const filteredDocuments = initialDocuments.filter(doc => {
    if (!searchQuery) return true;
    const searchLower = searchQuery.toLowerCase();
    return JSON.stringify(doc).toLowerCase().includes(searchLower);
  });

  const totalPages = Math.ceil(totalDocuments / pageSize);
  const startIndex = (currentPage - 1) * pageSize;
  const endIndex = Math.min(startIndex + pageSize, totalDocuments);

  const copyDocumentId = (id: string) => {
    navigator.clipboard.writeText(id);
    toast({
      title: "Copied",
      description: "Document ID copied to clipboard",
    });
  };

  const handleViewDocument = (doc: any) => {
    setSelectedDocument(doc);
    setViewDialogOpen(true);
  };

  const handleEditDocument = (doc: any) => {
    setSelectedDocument(doc);
    // Extract the actual document data (remove metadata)
    // Check if document has a 'data' field (API structure) or is flat
    let actualData = {};
    if (doc.data && typeof doc.data === 'object') {
      actualData = doc.data;
    } else {
      // Remove system fields from flat structure
      const { 
        _id, id, collection, created_at, updated_at, 
        created_by, updated_by, _version, version,
        _created_at, _updated_at, _created_by, _updated_by,
        ...dataOnly 
      } = doc;
      actualData = dataOnly;
    }
    setEditDocumentContent(JSON.stringify(actualData, null, 2));
    setEditDialogOpen(true);
  };

  const handleUpdateDocument = async () => {
    if (!selectedDocument || !editDocumentContent.trim()) return;

    try {
      const updatedData = JSON.parse(editDocumentContent);
      setIsUpdatingDocument(true);
      
      const docId = selectedDocument.id || selectedDocument._id;
      await documentsApi.update(collectionName, docId, updatedData);
      
      toast({
        title: "Success",
        description: "Document updated successfully",
      });
      
      setEditDialogOpen(false);
      setEditDocumentContent("");
      setSelectedDocument(null);
      
      // Refresh the page to show updated document
      window.location.reload();
    } catch (error) {
      if (error instanceof SyntaxError) {
        toast({
          title: "Invalid JSON",
          description: "Please enter valid JSON format",
          variant: "destructive",
        });
      } else {
        toast({
          title: "Error",
          description: "Failed to update document",
          variant: "destructive",
        });
      }
    } finally {
      setIsUpdatingDocument(false);
    }
  };

  const handleDeleteDocument = async (doc: any) => {
    const docId = doc.id || doc._id;
    
    if (!confirm(`Are you sure you want to delete this document?\n\nDocument ID: ${docId}`)) {
      return;
    }

    try {
      setIsDeletingDocument(true);
      await documentsApi.delete(collectionName, docId);
      
      toast({
        title: "Success",
        description: "Document deleted successfully",
      });
      
      // Refresh the page to update the list
      window.location.reload();
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete document",
        variant: "destructive",
      });
    } finally {
      setIsDeletingDocument(false);
    }
  };

  const handleEditSchema = () => {
    const currentSchema = collection.schema || { type: "object", properties: {}, required: [] };
    setEditSchemaContent(JSON.stringify(currentSchema, null, 2));
    setSchemaDialogOpen(true);
  };

  const handleUpdateSchema = async () => {
    if (!editSchemaContent.trim()) {
      toast({
        title: "Error",
        description: "Schema cannot be empty",
        variant: "destructive",
      });
      return;
    }

    try {
      const schemaData = JSON.parse(editSchemaContent);
      setIsUpdatingSchema(true);
      
      await collectionsApi.update(collectionName, { schema: schemaData });
      
      toast({
        title: "Success",
        description: "Schema updated successfully",
      });
      
      setSchemaDialogOpen(false);
      setEditSchemaContent("");
      
      // Refresh the page to show updated schema
      window.location.reload();
    } catch (error) {
      if (error instanceof SyntaxError) {
        toast({
          title: "Invalid JSON",
          description: "Please enter valid JSON schema format",
          variant: "destructive",
        });
      } else {
        toast({
          title: "Error",
          description: "Failed to update schema",
          variant: "destructive",
        });
      }
    } finally {
      setIsUpdatingSchema(false);
    }
  };

  const handleDuplicateDocument = (doc: any) => {
    // Extract actual data without metadata
    let actualData = {};
    if (doc.data && typeof doc.data === 'object') {
      actualData = doc.data;
    } else {
      // Remove system fields from flat structure
      const { 
        _id, id, collection, created_at, updated_at, 
        created_by, updated_by, _version, version,
        _created_at, _updated_at, _created_by, _updated_by,
        ...dataOnly 
      } = doc;
      actualData = dataOnly;
    }
    setNewDocumentContent(JSON.stringify(actualData, null, 2));
    setDocumentDialogOpen(true);
  };

  const handleCreateDocument = async () => {
    if (!newDocumentContent.trim()) {
      toast({
        title: "Error",
        description: "Document content cannot be empty",
        variant: "destructive",
      });
      return;
    }

    try {
      const documentData = JSON.parse(newDocumentContent);
      setIsCreatingDocument(true);
      
      // Call API to create document
      await documentsApi.create(collectionName, documentData);
      
      toast({
        title: "Success",
        description: "Document created successfully",
      });
      
      setDocumentDialogOpen(false);
      setNewDocumentContent("");
      
      // Refresh the page to show new document
      window.location.reload();
    } catch (error) {
      if (error instanceof SyntaxError) {
        toast({
          title: "Invalid JSON",
          description: "Please enter valid JSON format",
          variant: "destructive",
        });
      } else {
        toast({
          title: "Error",
          description: "Failed to create document",
          variant: "destructive",
        });
      }
    } finally {
      setIsCreatingDocument(false);
    }
  };

  const tabsList = [
    { value: "documents", label: "Documents", icon: FileJson, count: totalDocuments },
    { value: "schema", label: "Schema", icon: FileCode },
    { value: "indexes", label: "Indexes", icon: Key, count: indexes.length },
    { value: "vectors", label: "Vector Fields", icon: Layers },
    { value: "settings", label: "Settings", icon: Settings },
  ];

  return (
    <>
    <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
      <TabsList className="h-12 p-1 bg-muted/50">
        {tabsList.map((tab) => {
          const IconComponent = tab.icon;
          return (
            <TabsTrigger 
              key={tab.value} 
              value={tab.value}
              className="flex items-center gap-2 data-[state=active]:bg-background data-[state=active]:shadow-sm"
            >
              <IconComponent className="h-4 w-4" />
              <span>{tab.label}</span>
              {tab.count !== undefined && (
                <Badge variant="secondary" className="ml-1.5 h-5 px-1.5 text-xs">
                  {tab.count}
                </Badge>
              )}
            </TabsTrigger>
          );
        })}
      </TabsList>

      <TabsContent value="documents" className="space-y-4 mt-6">
        {/* Documents Toolbar */}
        <Card className="rounded-none shadow-none">
          <CardContent className="p-4">
            <div className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-2 flex-1 max-w-md">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    placeholder="Search documents..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-9 h-9"
                  />
                </div>
                <Button variant="outline" size="sm">
                  <Filter className="h-4 w-4 mr-2" />
                  Filter
                </Button>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="outline" size="sm">
                  <Download className="h-4 w-4 mr-2" />
                  Export
                </Button>
                <Button size="sm" onClick={() => setDocumentDialogOpen(true)}>
                  <Plus className="h-4 w-4 mr-2" />
                  Add Document
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Documents Table */}
        <Card className="rounded-none shadow-none">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead className="w-[50px]">
                  <input type="checkbox" className="rounded border-gray-300" />
                </TableHead>
                <TableHead className="font-semibold">Document ID</TableHead>
                <TableHead className="font-semibold">Data Preview</TableHead>
                <TableHead className="font-semibold">Size</TableHead>
                <TableHead className="font-semibold">Created</TableHead>
                <TableHead className="font-semibold">Modified</TableHead>
                <TableHead className="text-right font-semibold">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredDocuments.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="h-32">
                    <div className="flex flex-col items-center justify-center">
                      <div className="p-3 bg-muted mb-3">
                        <FileJson className="h-8 w-8 text-muted-foreground" />
                      </div>
                      <p className="font-medium text-muted-foreground">
                        {searchQuery ? "No documents match your search" : "No documents yet"}
                      </p>
                      <p className="text-sm text-muted-foreground mt-1">
                        {searchQuery ? "Try adjusting your search terms" : "Add your first document to get started"}
                      </p>
                      {!searchQuery && (
                        <Button className="mt-4" size="sm" onClick={() => setDocumentDialogOpen(true)}>
                          <Plus className="h-4 w-4 mr-2" />
                          Add Document
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                filteredDocuments.slice(startIndex, endIndex).map((doc) => {
                  const docId = doc.id || doc._id || "";
                  const docSize = JSON.stringify(doc).length;
                  const sizeStr = docSize < 1024 ? `${docSize} B` : `${(docSize / 1024).toFixed(1)} KB`;
                  
                  return (
                    <TableRow key={docId} className="hover:bg-muted/30">
                      <TableCell>
                        <input type="checkbox" className="rounded border-gray-300" />
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <code className="text-xs bg-muted px-2 py-1 font-mono">
                            {docId.substring(0, 8)}...
                          </code>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity"
                            onClick={() => copyDocumentId(docId)}
                          >
                            <Copy className="h-3 w-3" />
                          </Button>
                        </div>
                      </TableCell>
                      <TableCell>
                        <code className="text-xs text-muted-foreground line-clamp-1 max-w-xs">
                          {JSON.stringify(doc.data || doc).substring(0, 100)}...
                        </code>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-muted-foreground">{sizeStr}</span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-muted-foreground">
                          {formatLocalDateTime(doc.created_at)}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-muted-foreground">
                          {formatLocalDateTime(doc.updated_at)}
                        </span>
                      </TableCell>
                      <TableCell className="text-right">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-8 w-8">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleViewDocument(doc)}>
                              <Eye className="h-4 w-4 mr-2" />
                              View
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleEditDocument(doc)}>
                              <Edit className="h-4 w-4 mr-2" />
                              Edit
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => handleDuplicateDocument(doc)}>
                              <Copy className="h-4 w-4 mr-2" />
                              Duplicate
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem 
                              className="text-destructive"
                              onClick={() => handleDeleteDocument(doc)}
                              disabled={isDeletingDocument}
                            >
                              <Trash2 className="h-4 w-4 mr-2" />
                              {isDeletingDocument ? "Deleting..." : "Delete"}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
          
          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between px-4 py-3 border-t">
              <div className="text-sm text-muted-foreground">
                Showing <span className="font-medium">{startIndex + 1}</span> to{" "}
                <span className="font-medium">{endIndex}</span> of{" "}
                <span className="font-medium">{totalDocuments}</span> documents
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setCurrentPage(1)}
                  disabled={currentPage === 1}
                >
                  <ChevronsLeft className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                  disabled={currentPage === 1}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <div className="flex items-center gap-1">
                  <span className="text-sm px-2">
                    Page {currentPage} of {totalPages}
                  </span>
                </div>
                <Button
                  variant="outline"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                  disabled={currentPage === totalPages}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => setCurrentPage(totalPages)}
                  disabled={currentPage === totalPages}
                >
                  <ChevronsRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </Card>
      </TabsContent>

      <TabsContent value="schema" className="space-y-4 mt-6">
        <Card className="rounded-none shadow-none">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Collection Schema</CardTitle>
                <CardDescription>
                  Define validation rules and structure for your documents
                </CardDescription>
              </div>
              <Button onClick={handleEditSchema}>
                <Edit className="h-4 w-4 mr-2" />
                Edit Schema
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="bg-muted/50 p-4">
              <pre className="text-sm overflow-x-auto">
                <code>{JSON.stringify(collection.schema || { type: "object", properties: {} }, null, 2)}</code>
              </pre>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="indexes" className="space-y-4 mt-6">
        <Card className="rounded-none shadow-none">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Collection Indexes</CardTitle>
                <CardDescription>
                  Optimize query performance with strategic indexes
                </CardDescription>
              </div>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Create Index
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  <TableHead className="font-semibold">Index Name</TableHead>
                  <TableHead className="font-semibold">Fields</TableHead>
                  <TableHead className="font-semibold">Type</TableHead>
                  <TableHead className="font-semibold">Properties</TableHead>
                  <TableHead className="text-right font-semibold">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {indexes.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={5} className="h-32 text-center">
                      <div className="flex flex-col items-center justify-center">
                        <div className="p-3 bg-muted mb-3">
                          <Key className="h-8 w-8 text-muted-foreground" />
                        </div>
                        <p className="font-medium text-muted-foreground">No indexes defined</p>
                        <p className="text-sm text-muted-foreground mt-1">
                          Create indexes to improve query performance
                        </p>
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  indexes.map((index: any) => (
                    <TableRow key={index.name}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <code className="font-medium">{index.name}</code>
                          {index.name === "_id_" && (
                            <Badge variant="outline" className="text-xs">Primary</Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <code className="text-xs bg-muted px-2 py-1">
                          {Object.entries(index.key || {}).map(([field, direction]) => 
                            `${field}${direction === -1 ? ' ↓' : ' ↑'}`
                          ).join(", ")}
                        </code>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline">
                          {index.unique ? "Unique" : "Standard"}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-1">
                          {index.sparse && <Badge variant="secondary" className="text-xs">Sparse</Badge>}
                          {index.background && <Badge variant="secondary" className="text-xs">Background</Badge>}
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        {index.name !== "_id_" && (
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Index Info */}
        <Card className="bg-blue-50/50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-900 rounded-none shadow-none">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium flex items-center gap-2">
              <Info className="h-4 w-4" />
              Index Best Practices
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="text-sm space-y-1.5">
              <li className="flex items-start gap-2">
                <span className="text-blue-600 dark:text-blue-400">•</span>
                Create indexes on fields used frequently in queries
              </li>
              <li className="flex items-start gap-2">
                <span className="text-blue-600 dark:text-blue-400">•</span>
                Use compound indexes for queries with multiple fields
              </li>
              <li className="flex items-start gap-2">
                <span className="text-blue-600 dark:text-blue-400">•</span>
                Unique indexes enforce data integrity constraints
              </li>
              <li className="flex items-start gap-2">
                <span className="text-blue-600 dark:text-blue-400">•</span>
                Monitor index usage and remove unused indexes
              </li>
            </ul>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="vectors" className="space-y-4 mt-6">
        <VectorFieldsManager 
          collectionName={collectionName}
          onFieldsUpdate={() => {
            // Optionally refresh collection data
          }}
        />
      </TabsContent>

      <TabsContent value="settings" className="space-y-4 mt-6">
        <div className="grid gap-4">
          <Card className="rounded-none shadow-none">
            <CardHeader>
              <CardTitle>Collection Features</CardTitle>
              <CardDescription>
                Configure advanced features for your collection
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {[
                {
                  name: "Document Versioning",
                  description: "Keep a complete history of all document changes",
                  enabled: collection.settings?.versioning,
                  icon: Clock,
                },
                {
                  name: "Soft Delete",
                  description: "Mark documents as deleted instead of permanently removing them",
                  enabled: collection.settings?.soft_delete,
                  icon: Shield,
                },
                {
                  name: "Audit Logging",
                  description: "Track all operations performed on documents for compliance",
                  enabled: collection.settings?.auditing,
                  icon: Activity,
                },
                {
                  name: "Schema Validation",
                  description: "Enforce document structure and data types",
                  enabled: !!collection.schema,
                  icon: Layers,
                },
              ].map((feature) => {
                const IconComponent = feature.icon;
                return (
                  <div key={feature.name} className="flex items-center justify-between p-4 border">
                    <div className="flex items-start gap-3">
                      <div className="p-2 bg-muted">
                        <IconComponent className="h-4 w-4" />
                      </div>
                      <div className="space-y-1">
                        <p className="font-medium">{feature.name}</p>
                        <p className="text-sm text-muted-foreground">
                          {feature.description}
                        </p>
                      </div>
                    </div>
                    <Badge variant={feature.enabled ? "default" : "secondary"}>
                      {feature.enabled ? "Enabled" : "Disabled"}
                    </Badge>
                  </div>
                );
              })}
            </CardContent>
          </Card>

          <Card className="rounded-none shadow-none">
            <CardHeader>
              <CardTitle>Collection Information</CardTitle>
              <CardDescription>
                Metadata and configuration details
              </CardDescription>
            </CardHeader>
            <CardContent>
              <dl className="space-y-3">
                <div className="flex justify-between py-2 border-b">
                  <dt className="text-sm text-muted-foreground">Collection Name</dt>
                  <dd className="text-sm font-medium">{collectionName}</dd>
                </div>
                <div className="flex justify-between py-2 border-b">
                  <dt className="text-sm text-muted-foreground">Created</dt>
                  <dd className="text-sm font-medium">
                    {formatLocalDate(collection.created_at)}
                  </dd>
                </div>
                <div className="flex justify-between py-2 border-b">
                  <dt className="text-sm text-muted-foreground">Last Modified</dt>
                  <dd className="text-sm font-medium">
                    {formatLocalDate(collection.updated_at)}
                  </dd>
                </div>
                <div className="flex justify-between py-2">
                  <dt className="text-sm text-muted-foreground">Storage Engine</dt>
                  <dd className="text-sm font-medium">DocumentDB</dd>
                </div>
              </dl>
            </CardContent>
          </Card>
        </div>
      </TabsContent>
    </Tabs>

    {/* Add Document Dialog */}
    <Dialog open={documentDialogOpen} onOpenChange={setDocumentDialogOpen}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Add New Document</DialogTitle>
          <DialogDescription>
            Enter your document data in JSON format. The document ID will be automatically generated.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="document-content">Document Data (JSON)</Label>
            <Textarea
              id="document-content"
              placeholder='{
  "name": "Example",
  "description": "Sample document",
  "value": 123
}'
              value={newDocumentContent}
              onChange={(e) => setNewDocumentContent(e.target.value)}
              className="min-h-[300px] font-mono text-sm"
            />
            <p className="text-xs text-muted-foreground">
              Enter valid JSON. Use double quotes for keys and string values.
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button 
            variant="outline" 
            onClick={() => {
              setDocumentDialogOpen(false);
              setNewDocumentContent("");
            }}
            disabled={isCreatingDocument}
          >
            Cancel
          </Button>
          <Button 
            onClick={handleCreateDocument}
            disabled={isCreatingDocument || !newDocumentContent.trim()}
          >
            {isCreatingDocument ? "Creating..." : "Create Document"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    {/* View Document Dialog */}
    <Dialog open={viewDialogOpen} onOpenChange={setViewDialogOpen}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>Document Details</DialogTitle>
          <DialogDescription>
            Full document data in JSON format
          </DialogDescription>
        </DialogHeader>
        <ScrollArea className="h-[400px] w-full rounded-md border p-4">
          <pre className="text-sm">
            <code>
              {(() => {
                if (!selectedDocument) return "";
                
                // Format document to show data and metadata separately
                let actualData = {};
                let metadata = {};
                
                if (selectedDocument.data && typeof selectedDocument.data === 'object') {
                  actualData = selectedDocument.data;
                  metadata = {
                    id: selectedDocument.id || selectedDocument._id,
                    collection: selectedDocument.collection,
                    created_at: selectedDocument.created_at,
                    updated_at: selectedDocument.updated_at,
                    created_by: selectedDocument.created_by,
                    updated_by: selectedDocument.updated_by,
                    version: selectedDocument.version || selectedDocument._version
                  };
                } else {
                  const { 
                    _id, id, collection, created_at, updated_at, 
                    created_by, updated_by, _version, version,
                    _created_at, _updated_at, _created_by, _updated_by,
                    ...dataOnly 
                  } = selectedDocument;
                  
                  actualData = dataOnly;
                  metadata = {
                    id: _id || id,
                    collection,
                    created_at: created_at || _created_at,
                    updated_at: updated_at || _updated_at,
                    created_by: created_by || _created_by,
                    updated_by: updated_by || _updated_by,
                    version: version || _version
                  };
                }
                
                // Clean up undefined values in metadata
                Object.keys(metadata).forEach(key => {
                  if (metadata[key] === undefined) {
                    delete metadata[key];
                  }
                });
                
                const displayDoc = {
                  data: actualData,
                  _metadata: metadata
                };
                
                return JSON.stringify(displayDoc, null, 2);
              })()}
            </code>
          </pre>
        </ScrollArea>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => {
              if (selectedDocument) {
                const docId = selectedDocument.id || selectedDocument._id;
                copyDocumentId(docId);
              }
            }}
          >
            <Copy className="h-4 w-4 mr-2" />
            Copy ID
          </Button>
          <Button
            variant="outline"
            onClick={() => {
              if (selectedDocument) {
                navigator.clipboard.writeText(JSON.stringify(selectedDocument, null, 2));
                toast({
                  title: "Copied",
                  description: "Document copied to clipboard",
                });
              }
            }}
          >
            <Copy className="h-4 w-4 mr-2" />
            Copy JSON
          </Button>
          <Button onClick={() => setViewDialogOpen(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    {/* Edit Document Dialog */}
    <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Edit Document</DialogTitle>
          <DialogDescription>
            Modify the document data in JSON format. Metadata fields will be preserved.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-document-content">Document Data (JSON)</Label>
            <Textarea
              id="edit-document-content"
              value={editDocumentContent}
              onChange={(e) => setEditDocumentContent(e.target.value)}
              className="min-h-[300px] font-mono text-sm"
              placeholder="Enter valid JSON..."
            />
            <p className="text-xs text-muted-foreground">
              Edit the document data. System fields like ID and timestamps are preserved.
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button 
            variant="outline" 
            onClick={() => {
              setEditDialogOpen(false);
              setEditDocumentContent("");
              setSelectedDocument(null);
            }}
            disabled={isUpdatingDocument}
          >
            Cancel
          </Button>
          <Button 
            onClick={handleUpdateDocument}
            disabled={isUpdatingDocument || !editDocumentContent.trim()}
          >
            {isUpdatingDocument ? "Updating..." : "Save Changes"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    {/* Edit Schema Dialog */}
    <Dialog open={schemaDialogOpen} onOpenChange={setSchemaDialogOpen}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>Edit Collection Schema</DialogTitle>
          <DialogDescription>
            Define the JSON schema for document validation. Use standard JSON Schema format.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-schema-content">Schema Definition (JSON Schema)</Label>
            <Textarea
              id="edit-schema-content"
              value={editSchemaContent}
              onChange={(e) => setEditSchemaContent(e.target.value)}
              className="min-h-[400px] font-mono text-sm"
              placeholder='{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "Name of the item"
    },
    "value": {
      "type": "number",
      "minimum": 0
    }
  },
  "required": ["name"]
}'
            />
            <p className="text-xs text-muted-foreground">
              Define properties, types, required fields, and validation rules for documents in this collection.
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button 
            variant="outline" 
            onClick={() => {
              setSchemaDialogOpen(false);
              setEditSchemaContent("");
            }}
            disabled={isUpdatingSchema}
          >
            Cancel
          </Button>
          <Button 
            onClick={handleUpdateSchema}
            disabled={isUpdatingSchema || !editSchemaContent.trim()}
          >
            {isUpdatingSchema ? "Updating..." : "Save Schema"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
    </>
  );
}