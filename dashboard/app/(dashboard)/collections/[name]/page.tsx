"use client"

import { useState, useEffect } from "react"
import { useParams, useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { collectionsApi, documentsApi } from "@/lib/api"
import { 
  ArrowLeft, Plus, Database, Trash2, Edit2, Save, X, FileJson, 
  Key, Settings, Search, Copy, CheckCircle, Info, Code, 
  Calendar, Hash, MoreVertical, ExternalLink, Activity, Edit, FileCode
} from "lucide-react"
import { useToast } from "@/hooks/use-toast"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Separator } from "@/components/ui/separator"
import { ScrollArea } from "@/components/ui/scroll-area"
import { SchemaEditor } from "@/components/schema-editor"
import { validateDocument, type ValidationError } from "@/lib/schema-validator"

export default function CollectionDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { toast } = useToast()
  
  const collectionName = params?.name as string
  
  const [collection, setCollection] = useState<any>(null)
  const [documents, setDocuments] = useState<any[]>([])
  const [indexes, setIndexes] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState("documents")
  
  // Edit mode states
  const [editMode, setEditMode] = useState(false)
  const [editedCollection, setEditedCollection] = useState<any>(null)
  const [schemaEditMode, setSchemaEditMode] = useState(false)
  const [editedSchema, setEditedSchema] = useState<any>({ type: 'object', properties: {}, required: [] })
  
  // Dialog states
  const [indexDialogOpen, setIndexDialogOpen] = useState(false)
  const [documentDialogOpen, setDocumentDialogOpen] = useState(false)
  const [documentViewDialogOpen, setDocumentViewDialogOpen] = useState(false)
  const [selectedDocument, setSelectedDocument] = useState<any>(null)
  const [editingDocument, setEditingDocument] = useState(false)
  const [editedDocumentContent, setEditedDocumentContent] = useState("")
  const [newIndex, setNewIndex] = useState({
    name: "",
    fields: "",
    unique: false,
    sparse: false,
  })
  const [newDocument, setNewDocument] = useState("")
  const [searchQuery, setSearchQuery] = useState("")
  const [documentValidationErrors, setDocumentValidationErrors] = useState<ValidationError[]>([])
  const [editValidationErrors, setEditValidationErrors] = useState<ValidationError[]>([])
  
  // Pagination states
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [totalDocuments, setTotalDocuments] = useState(0)
  const [totalPages, setTotalPages] = useState(0)

  useEffect(() => {
    if (collectionName) {
      loadCollection()
      loadDocuments()
      loadIndexes()
    }
  }, [collectionName])
  
  useEffect(() => {
    if (collectionName) {
      loadDocuments()
    }
  }, [currentPage, pageSize])

  const loadCollection = async () => {
    try {
      const response = await collectionsApi.get(collectionName!)
      setCollection(response)
      setEditedCollection(response)
      // Initialize schema with default if not present
      const defaultSchema = { type: 'object', properties: {}, required: [] }
      setEditedSchema(response?.schema || defaultSchema)
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load collection",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const loadDocuments = async () => {
    try {
      const response = await documentsApi.list(collectionName!, {
        page: currentPage,
        limit: pageSize,
      })
      setDocuments(response.documents || response.data || [])
      setTotalDocuments(response.total || 0)
      setTotalPages(response.totalPages || 1)
    } catch (error) {
      console.error("Failed to load documents:", error)
    }
  }

  const loadIndexes = async () => {
    try {
      const response = await collectionsApi.getIndexes(collectionName!)
      setIndexes(response.indexes || [])
    } catch (error) {
      console.error("Failed to load indexes:", error)
    }
  }

  const handleSaveCollection = async () => {
    try {
      await collectionsApi.update(collectionName!, {
        description: editedCollection.description,
        settings: editedCollection.settings,
      })
      toast({
        title: "Success",
        description: "Collection updated successfully",
      })
      setCollection(editedCollection)
      setEditMode(false)
      loadCollection()
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update collection",
        variant: "destructive",
      })
    }
  }
  
  const saveSchema = async () => {
    try {
      await collectionsApi.update(collectionName!, {
        schema: editedSchema,
      })
      toast({
        title: "Success",
        description: "Schema updated successfully",
      })
      setCollection({ ...collection, schema: editedSchema })
      setSchemaEditMode(false)
      loadCollection()
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update schema",
        variant: "destructive",
      })
    }
  }

  const handleCreateIndex = async () => {
    try {
      const fields = newIndex.fields.split(',').reduce((acc, field) => {
        const trimmed = field.trim()
        if (trimmed) {
          // Support format like "email:1" or "createdAt:-1"
          const [fieldName, direction] = trimmed.split(':')
          const directionValue = direction ? parseInt(direction) : 1
          acc[fieldName.trim()] = directionValue
        }
        return acc
      }, {} as Record<string, number>)

      await collectionsApi.createIndex(collectionName!, {
        name: newIndex.name,
        keys: fields,
        options: {
          unique: newIndex.unique,
          sparse: newIndex.sparse,
        },
      })
      
      toast({
        title: "Success",
        description: "Index created successfully",
      })
      setIndexDialogOpen(false)
      setNewIndex({ name: "", fields: "", unique: false, sparse: false })
      loadIndexes()
    } catch (error: any) {
      console.error("Index creation error:", error)
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to create index",
        variant: "destructive",
      })
    }
  }

  const handleDeleteIndex = async (indexName: string) => {
    if (!confirm(`Are you sure you want to delete index "${indexName}"?`)) return
    
    try {
      await collectionsApi.deleteIndex(collectionName!, indexName)
      toast({
        title: "Success",
        description: "Index deleted successfully",
      })
      loadIndexes()
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete index",
        variant: "destructive",
      })
    }
  }

  const handleCreateDocument = async () => {
    try {
      const documentData = JSON.parse(newDocument)
      
      // Validate against schema
      const errors = validateDocument(documentData, collection?.schema)
      if (errors.length > 0) {
        setDocumentValidationErrors(errors)
        toast({
          title: "Validation Error",
          description: `Document has ${errors.length} validation error(s)`,
          variant: "destructive",
        })
        return
      }
      
      await documentsApi.create(collectionName!, documentData)
      toast({
        title: "Success",
        description: "Document created successfully",
      })
      setDocumentDialogOpen(false)
      setNewDocument("")
      setDocumentValidationErrors([])
      loadDocuments()
    } catch (error: any) {
      if (error instanceof SyntaxError) {
        toast({
          title: "Error",
          description: "Invalid JSON format",
          variant: "destructive",
        })
      } else {
        toast({
          title: "Error",
          description: error.response?.data?.error || "Failed to create document",
          variant: "destructive",
        })
      }
    }
  }

  const handleDeleteDocument = async (documentId: string) => {
    if (!confirm("Are you sure you want to delete this document?")) return
    
    try {
      await documentsApi.delete(collectionName!, documentId)
      toast({
        title: "Success",
        description: "Document deleted successfully",
      })
      loadDocuments()
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete document",
        variant: "destructive",
      })
    }
  }

  const viewDocument = (doc: any) => {
    setSelectedDocument(doc)
    
    // Documents from API have actual data in a 'data' field, metadata at top level
    let actualData = {}
    if (doc.data && typeof doc.data === 'object') {
      // Document structure from API: {id, collection, data: {...actualFields}, created_by, ...}
      actualData = doc.data
    } else {
      // Fallback: extract data by removing metadata fields
      const { 
        _id, id, collection, created_by, updated_by, 
        created_at, updated_at, version, _version,
        _created_at, _updated_at, _created_by, _updated_by,
        data, // exclude 'data' field if it exists
        ...dataOnly 
      } = doc
      actualData = dataOnly
    }
    
    setEditedDocumentContent(JSON.stringify(actualData, null, 2))
    setEditingDocument(false)
    setDocumentViewDialogOpen(true)
  }

  const copyDocumentId = (id: string) => {
    navigator.clipboard.writeText(id)
    toast({
      title: "Copied",
      description: "Document ID copied to clipboard",
    })
  }

  const handleUpdateDocument = async () => {
    try {
      const updatedData = JSON.parse(editedDocumentContent)
      const docId = selectedDocument.id || selectedDocument._id
      
      // Remove all metadata fields from the update data
      const { 
        _id, id, collection, created_by, updated_by, 
        created_at, updated_at, version, _version,
        _created_at, _updated_at, _created_by, _updated_by,
        ...updateData 
      } = updatedData
      
      // Validate against schema
      const errors = validateDocument(updateData, collection?.schema)
      if (errors.length > 0) {
        setEditValidationErrors(errors)
        toast({
          title: "Validation Error",
          description: `Document has ${errors.length} validation error(s)`,
          variant: "destructive",
        })
        return
      }
      
      await documentsApi.update(collectionName!, docId, updateData)
      
      toast({
        title: "Success",
        description: "Document updated successfully",
      })
      
      setDocumentViewDialogOpen(false)
      setEditingDocument(false)
      setEditValidationErrors([])
      loadDocuments()
    } catch (error: any) {
      if (error instanceof SyntaxError) {
        toast({
          title: "Error",
          description: "Invalid JSON format",
          variant: "destructive",
        })
      } else {
        toast({
          title: "Error",
          description: error.response?.data?.error || "Failed to update document",
          variant: "destructive",
        })
      }
    }
  }

  const filteredDocuments = documents.filter(doc => {
    if (!searchQuery) return true
    const searchLower = searchQuery.toLowerCase()
    return JSON.stringify(doc).toLowerCase().includes(searchLower)
  })

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (!collection) {
    return (
      <div className="container mx-auto py-6">
        <div className="text-center">
          <h2 className="text-2xl font-semibold">Collection not found</h2>
          <Button onClick={() => router.push('/collections')} className="mt-4">
            Back to Collections
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => router.push('/collections')}
                className="hover:bg-secondary"
              >
                <ArrowLeft className="h-4 w-4" />
              </Button>
              <div>
                <div className="flex items-center gap-2">
                  <Database className="h-5 w-5 text-muted-foreground" />
                  <h1 className="text-2xl font-bold">{collectionName}</h1>
                </div>
                <p className="text-sm text-muted-foreground mt-1">
                  {collection.description || "No description provided"}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              {editMode ? (
                <>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setEditedCollection(collection)
                      setEditMode(false)
                    }}
                  >
                    <X className="h-4 w-4 mr-2" />
                    Cancel
                  </Button>
                  <Button onClick={handleSaveCollection}>
                    <Save className="h-4 w-4 mr-2" />
                    Save Changes
                  </Button>
                </>
              ) : (
                <>
                  <Button variant="outline" onClick={() => setEditMode(true)}>
                    <Edit2 className="h-4 w-4 mr-2" />
                    Edit
                  </Button>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="outline" size="icon">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => {}}>
                        <ExternalLink className="h-4 w-4 mr-2" />
                        Export Collection
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => {}}>
                        <Copy className="h-4 w-4 mr-2" />
                        Duplicate Collection
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem 
                        className="text-destructive"
                        onClick={() => {
                          if (confirm(`Are you sure you want to delete ${collectionName}?`)) {
                            // Delete collection
                          }
                        }}
                      >
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete Collection
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </>
              )}
            </div>
          </div>
          
          {/* Stats Bar */}
          <div className="flex items-center gap-6 mt-4 text-sm">
            <div className="flex items-center gap-2">
              <FileJson className="h-4 w-4 text-muted-foreground" />
              <span className="text-muted-foreground">Documents:</span>
              <span className="font-semibold">{documents.length}</span>
            </div>
            <Separator orientation="vertical" className="h-4" />
            <div className="flex items-center gap-2">
              <Key className="h-4 w-4 text-muted-foreground" />
              <span className="text-muted-foreground">Indexes:</span>
              <span className="font-semibold">{indexes.length}</span>
            </div>
            <Separator orientation="vertical" className="h-4" />
            <div className="flex items-center gap-2">
              <Calendar className="h-4 w-4 text-muted-foreground" />
              <span className="text-muted-foreground">Created:</span>
              <span className="font-semibold">
                {collection.created_at ? new Date(collection.created_at).toLocaleDateString() : "N/A"}
              </span>
            </div>
            <Separator orientation="vertical" className="h-4" />
            <div className="flex items-center gap-2">
              <Activity className="h-4 w-4 text-muted-foreground" />
              <span className="text-muted-foreground">Status:</span>
              <Badge variant="default" className="h-5">Active</Badge>
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 container mx-auto px-6 py-6">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full flex flex-col">
          <TabsList className="grid w-full max-w-[500px] grid-cols-4">
            <TabsTrigger value="documents" className="flex items-center gap-2">
              <FileJson className="h-4 w-4" />
              Documents
            </TabsTrigger>
            <TabsTrigger value="schema" className="flex items-center gap-2">
              <FileCode className="h-4 w-4" />
              Schema
            </TabsTrigger>
            <TabsTrigger value="indexes" className="flex items-center gap-2">
              <Key className="h-4 w-4" />
              Indexes
            </TabsTrigger>
            <TabsTrigger value="settings" className="flex items-center gap-2">
              <Settings className="h-4 w-4" />
              Settings
            </TabsTrigger>
          </TabsList>

          <TabsContent value="documents" className="flex-1 mt-6">
            <div className="space-y-4">
              {/* Documents Toolbar */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 flex-1 max-w-sm">
                  <div className="relative flex-1">
                    <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      placeholder="Search documents..."
                      value={searchQuery}
                      onChange={(e) => setSearchQuery(e.target.value)}
                      className="pl-9"
                    />
                  </div>
                </div>
                <Button onClick={() => setDocumentDialogOpen(true)}>
                  <Plus className="h-4 w-4 mr-2" />
                  Add Document
                </Button>
              </div>

              {/* Documents Table */}
              <Card>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[200px]">Document ID</TableHead>
                      <TableHead>Data Preview</TableHead>
                      <TableHead className="w-[150px]">Created</TableHead>
                      <TableHead className="w-[100px] text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredDocuments.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={4} className="h-32 text-center">
                          <div className="flex flex-col items-center justify-center">
                            <FileJson className="h-8 w-8 text-muted-foreground mb-2" />
                            <p className="text-sm text-muted-foreground">
                              {searchQuery ? "No documents match your search" : "No documents yet"}
                            </p>
                          </div>
                        </TableCell>
                      </TableRow>
                    ) : (
                      filteredDocuments.map((doc) => (
                        <TableRow key={doc.id || doc._id}>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <code className="text-xs bg-muted px-2 py-1 rounded">
                                {(doc.id || doc._id || "").substring(0, 12)}...
                              </code>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6"
                                onClick={() => copyDocumentId(doc.id || doc._id)}
                              >
                                <Copy className="h-3 w-3" />
                              </Button>
                            </div>
                          </TableCell>
                          <TableCell>
                            <div 
                              className="text-xs text-muted-foreground cursor-pointer hover:text-foreground transition-colors"
                              onClick={() => viewDocument(doc)}
                            >
                              <code className="line-clamp-2">
                                {JSON.stringify(doc.data || doc, null, 2).substring(0, 150)}...
                              </code>
                            </div>
                          </TableCell>
                          <TableCell>
                            <span className="text-sm text-muted-foreground">
                              {doc.created_at ? new Date(doc.created_at).toLocaleDateString() : "N/A"}
                            </span>
                          </TableCell>
                          <TableCell className="text-right">
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon" className="h-8 w-8">
                                  <MoreVertical className="h-4 w-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                <DropdownMenuItem onClick={() => viewDocument(doc)}>
                                  <Code className="h-4 w-4 mr-2" />
                                  View Document
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => {
                                  viewDocument(doc)
                                  setEditingDocument(true)
                                }}>
                                  <Edit className="h-4 w-4 mr-2" />
                                  Edit Document
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => copyDocumentId(doc.id || doc._id)}>
                                  <Copy className="h-4 w-4 mr-2" />
                                  Copy ID
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem 
                                  className="text-destructive"
                                  onClick={() => handleDeleteDocument(doc.id || doc._id)}
                                >
                                  <Trash2 className="h-4 w-4 mr-2" />
                                  Delete
                                </DropdownMenuItem>
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </Card>
              
              {/* Pagination Controls */}
              {totalPages > 1 && (
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">
                      Showing {((currentPage - 1) * pageSize) + 1} to {Math.min(currentPage * pageSize, totalDocuments)} of {totalDocuments} documents
                    </span>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                      <Label htmlFor="page-size" className="text-sm">Page size:</Label>
                      <Select value={pageSize.toString()} onValueChange={(value) => {
                        setPageSize(parseInt(value))
                        setCurrentPage(1) // Reset to first page when changing page size
                      }}>
                        <SelectTrigger id="page-size" className="w-[70px]">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="10">10</SelectItem>
                          <SelectItem value="20">20</SelectItem>
                          <SelectItem value="50">50</SelectItem>
                          <SelectItem value="100">100</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                        disabled={currentPage === 1}
                      >
                        Previous
                      </Button>
                      <span className="text-sm">
                        Page {currentPage} of {totalPages}
                      </span>
                      <Button 
                        variant="outline" 
                        size="sm"
                        onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                        disabled={currentPage === totalPages}
                      >
                        Next
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </TabsContent>

          <TabsContent value="schema" className="flex-1 mt-6">
            <div className="space-y-4">
              {/* Schema Toolbar */}
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-semibold">Collection Schema</h3>
                  <p className="text-sm text-muted-foreground">
                    Define the structure and validation rules for documents using OpenAPI format
                  </p>
                </div>
                {!schemaEditMode ? (
                  <Button onClick={() => setSchemaEditMode(true)}>
                    <Edit2 className="h-4 w-4 mr-2" />
                    Edit Schema
                  </Button>
                ) : (
                  <div className="flex items-center gap-2">
                    <Button variant="outline" onClick={() => {
                      setEditedSchema(collection?.schema || { type: 'object', properties: {}, required: [] })
                      setSchemaEditMode(false)
                    }}>
                      <X className="h-4 w-4 mr-2" />
                      Cancel
                    </Button>
                    <Button onClick={saveSchema}>
                      <Save className="h-4 w-4 mr-2" />
                      Save Schema
                    </Button>
                  </div>
                )}
              </div>

              {/* Schema Editor */}
              <SchemaEditor
                schema={schemaEditMode ? editedSchema : (collection?.schema || { type: 'object', properties: {}, required: [] })}
                onChange={(newSchema) => {
                  if (schemaEditMode) {
                    setEditedSchema(newSchema);
                  }
                }}
                readOnly={!schemaEditMode}
              />
            </div>
          </TabsContent>

          <TabsContent value="indexes" className="flex-1 mt-6">
            <div className="space-y-4">
              {/* Indexes Toolbar */}
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-semibold">Collection Indexes</h3>
                  <p className="text-sm text-muted-foreground">
                    Optimize query performance with strategic indexes
                  </p>
                </div>
                <Button onClick={() => setIndexDialogOpen(true)}>
                  <Plus className="h-4 w-4 mr-2" />
                  Create Index
                </Button>
              </div>

              {/* Indexes Table */}
              <Card>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Index Name</TableHead>
                      <TableHead>Fields</TableHead>
                      <TableHead>Properties</TableHead>
                      <TableHead className="w-[100px] text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {indexes.length === 0 ? (
                      <TableRow>
                        <TableCell colSpan={4} className="h-32 text-center">
                          <div className="flex flex-col items-center justify-center">
                            <Key className="h-8 w-8 text-muted-foreground mb-2" />
                            <p className="text-sm text-muted-foreground">
                              No custom indexes defined
                            </p>
                          </div>
                        </TableCell>
                      </TableRow>
                    ) : (
                      indexes.map((index: any) => (
                        <TableRow key={index.name}>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <code className="text-sm font-medium">{index.name}</code>
                              {index.name === "_id_" && (
                                <Badge variant="secondary" className="text-xs">Primary</Badge>
                              )}
                            </div>
                          </TableCell>
                          <TableCell>
                            <code className="text-xs bg-muted px-2 py-1 rounded">
                              {Object.entries(index.key || {}).map(([field, direction]) => 
                                `${field}${direction === -1 ? ' (desc)' : ''}`
                              ).join(", ")}
                            </code>
                          </TableCell>
                          <TableCell>
                            <div className="flex gap-2">
                              {index.unique && <Badge variant="outline">Unique</Badge>}
                              {index.sparse && <Badge variant="outline">Sparse</Badge>}
                              {!index.unique && !index.sparse && index.name !== "_id_" && (
                                <span className="text-sm text-muted-foreground">Standard</span>
                              )}
                            </div>
                          </TableCell>
                          <TableCell className="text-right">
                            {index.name !== "_id_" ? (
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-8 w-8"
                                onClick={() => handleDeleteIndex(index.name)}
                              >
                                <Trash2 className="h-4 w-4 text-destructive" />
                              </Button>
                            ) : (
                              <span className="text-xs text-muted-foreground">System</span>
                            )}
                          </TableCell>
                        </TableRow>
                      ))
                    )}
                  </TableBody>
                </Table>
              </Card>

              {/* Index Info Card */}
              <Card className="bg-muted/50">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-medium flex items-center gap-2">
                    <Info className="h-4 w-4" />
                    Index Best Practices
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <ul className="text-sm text-muted-foreground space-y-1">
                    <li>• Create indexes on fields used frequently in queries</li>
                    <li>• Use compound indexes for queries with multiple fields</li>
                    <li>• Unique indexes enforce data integrity constraints</li>
                    <li>• Sparse indexes exclude documents without the indexed field</li>
                  </ul>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="settings" className="flex-1 mt-6">
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-semibold mb-1">Collection Settings</h3>
                <p className="text-sm text-muted-foreground">
                  Configure how your collection behaves
                </p>
              </div>

              <Card>
                <CardContent className="pt-6 space-y-6">
                  {editMode ? (
                    <>
                      <div className="space-y-2">
                        <Label htmlFor="description">Description</Label>
                        <Textarea
                          id="description"
                          value={editedCollection.description || ""}
                          onChange={(e) => setEditedCollection({
                            ...editedCollection,
                            description: e.target.value
                          })}
                          placeholder="Describe what this collection stores..."
                          className="resize-none"
                        />
                      </div>

                      <Separator />

                      <div className="space-y-4">
                        <h4 className="text-sm font-medium">Features</h4>
                        
                        <div className="flex items-center justify-between">
                          <div className="space-y-0.5">
                            <Label htmlFor="versioning">Document Versioning</Label>
                            <p className="text-sm text-muted-foreground">
                              Keep history of all document changes
                            </p>
                          </div>
                          <Switch
                            id="versioning"
                            checked={editedCollection.settings?.versioning || false}
                            onCheckedChange={(checked) => setEditedCollection({
                              ...editedCollection,
                              settings: { ...editedCollection.settings, versioning: checked }
                            })}
                          />
                        </div>

                        <div className="flex items-center justify-between">
                          <div className="space-y-0.5">
                            <Label htmlFor="soft_delete">Soft Delete</Label>
                            <p className="text-sm text-muted-foreground">
                              Mark documents as deleted instead of removing them
                            </p>
                          </div>
                          <Switch
                            id="soft_delete"
                            checked={editedCollection.settings?.soft_delete || false}
                            onCheckedChange={(checked) => setEditedCollection({
                              ...editedCollection,
                              settings: { ...editedCollection.settings, soft_delete: checked }
                            })}
                          />
                        </div>

                        <div className="flex items-center justify-between">
                          <div className="space-y-0.5">
                            <Label htmlFor="auditing">Audit Logging</Label>
                            <p className="text-sm text-muted-foreground">
                              Track all operations performed on documents
                            </p>
                          </div>
                          <Switch
                            id="auditing"
                            checked={editedCollection.settings?.auditing || false}
                            onCheckedChange={(checked) => setEditedCollection({
                              ...editedCollection,
                              settings: { ...editedCollection.settings, auditing: checked }
                            })}
                          />
                        </div>
                      </div>
                    </>
                  ) : (
                    <>
                      <div>
                        <Label className="text-muted-foreground">Description</Label>
                        <p className="mt-1">
                          {collection.description || "No description provided"}
                        </p>
                      </div>

                      <Separator />

                      <div className="space-y-4">
                        <h4 className="text-sm font-medium">Features</h4>
                        
                        <div className="grid gap-4">
                          <div className="flex items-center justify-between">
                            <div className="space-y-0.5">
                              <p className="text-sm font-medium">Document Versioning</p>
                              <p className="text-sm text-muted-foreground">
                                Keep history of all document changes
                              </p>
                            </div>
                            <Badge variant={collection.settings?.versioning ? "default" : "secondary"}>
                              {collection.settings?.versioning ? "Enabled" : "Disabled"}
                            </Badge>
                          </div>

                          <div className="flex items-center justify-between">
                            <div className="space-y-0.5">
                              <p className="text-sm font-medium">Soft Delete</p>
                              <p className="text-sm text-muted-foreground">
                                Mark documents as deleted instead of removing them
                              </p>
                            </div>
                            <Badge variant={collection.settings?.soft_delete ? "default" : "secondary"}>
                              {collection.settings?.soft_delete ? "Enabled" : "Disabled"}
                            </Badge>
                          </div>

                          <div className="flex items-center justify-between">
                            <div className="space-y-0.5">
                              <p className="text-sm font-medium">Audit Logging</p>
                              <p className="text-sm text-muted-foreground">
                                Track all operations performed on documents
                              </p>
                            </div>
                            <Badge variant={collection.settings?.auditing ? "default" : "secondary"}>
                              {collection.settings?.auditing ? "Enabled" : "Disabled"}
                            </Badge>
                          </div>
                        </div>
                      </div>

                      <Separator />

                      <div className="space-y-4">
                        <h4 className="text-sm font-medium">Collection Information</h4>
                        
                        <div className="grid gap-4 text-sm">
                          <div className="flex justify-between">
                            <span className="text-muted-foreground">Collection Name</span>
                            <code className="font-medium">{collection.name}</code>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-muted-foreground">Created</span>
                            <span className="font-medium">
                              {collection.created_at ? new Date(collection.created_at).toLocaleDateString() : "N/A"}
                            </span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-muted-foreground">Last Updated</span>
                            <span className="font-medium">
                              {collection.updated_at ? new Date(collection.updated_at).toLocaleDateString() : "N/A"}
                            </span>
                          </div>
                        </div>
                      </div>
                    </>
                  )}
                </CardContent>
              </Card>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      {/* Create Document Dialog */}
      <Dialog open={documentDialogOpen} onOpenChange={setDocumentDialogOpen}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>Create Document</DialogTitle>
            <DialogDescription>
              Enter document data in JSON format
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Document Data</Label>
              <Textarea
                placeholder='{"field": "value", "number": 123}'
                value={newDocument}
                onChange={(e) => {
                  setNewDocument(e.target.value)
                  // Real-time validation
                  try {
                    const documentData = JSON.parse(e.target.value)
                    const errors = validateDocument(documentData, collection?.schema)
                    setDocumentValidationErrors(errors)
                  } catch {
                    setDocumentValidationErrors([])
                  }
                }}
                className="min-h-[300px] font-mono text-sm"
              />
              <p className="text-xs text-muted-foreground">
                The document ID will be automatically generated if not provided
              </p>
              
              {/* Validation Errors */}
              {documentValidationErrors.length > 0 && (
                <div className="mt-2 p-3 bg-destructive/10 border border-destructive/20 rounded-md">
                  <p className="text-sm font-medium text-destructive mb-1">Validation Errors:</p>
                  <ul className="space-y-1">
                    {documentValidationErrors.map((error, index) => (
                      <li key={index} className="text-sm text-destructive/90">
                        • {error.message}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setDocumentDialogOpen(false)
              setDocumentValidationErrors([])
            }}>
              Cancel
            </Button>
            <Button onClick={handleCreateDocument}>
              Create Document
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* View/Edit Document Dialog */}
      <Dialog open={documentViewDialogOpen} onOpenChange={(open) => {
        if (!open) {
          setEditingDocument(false)
          setEditValidationErrors([])
        }
        setDocumentViewDialogOpen(open)
      }}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>
              {editingDocument ? "Edit Document" : "Document Details"}
            </DialogTitle>
            <DialogDescription>
              {editingDocument ? "Modify the document data in JSON format" : "Full document data in JSON format"}
            </DialogDescription>
          </DialogHeader>
          
          {editingDocument ? (
            <div className="space-y-2">
              <Textarea
                value={editedDocumentContent}
                onChange={(e) => {
                  setEditedDocumentContent(e.target.value)
                  // Real-time validation
                  try {
                    const documentData = JSON.parse(e.target.value)
                    const errors = validateDocument(documentData, collection?.schema)
                    setEditValidationErrors(errors)
                  } catch {
                    setEditValidationErrors([])
                  }
                }}
                className="h-[400px] font-mono text-sm"
                placeholder="Enter valid JSON..."
              />
              
              {/* Validation Errors */}
              {editValidationErrors.length > 0 && (
                <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
                  <p className="text-sm font-medium text-destructive mb-1">Validation Errors:</p>
                  <ul className="space-y-1">
                    {editValidationErrors.map((error, index) => (
                      <li key={index} className="text-sm text-destructive/90">
                        • {error.message}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          ) : (
            <ScrollArea className="h-[400px] w-full rounded-md border p-4">
              <pre className="text-sm">
                <code>{(() => {
                  // Show full document in view mode including metadata
                  const { 
                    _id, id, collection, created_by, updated_by, 
                    created_at, updated_at, version, _version,
                    _created_at, _updated_at, _created_by, _updated_by,
                    ...dataOnly 
                  } = selectedDocument || {}
                  
                  // Format for display with metadata at the bottom
                  const displayDoc = {
                    ...dataOnly,
                    _metadata: {
                      id: _id || id,
                      collection,
                      version: version || _version,
                      created_at: created_at || _created_at,
                      updated_at: updated_at || _updated_at,
                      created_by: created_by || _created_by,
                      updated_by: updated_by || _updated_by,
                    }
                  }
                  
                  return JSON.stringify(displayDoc, null, 2)
                })()}</code>
              </pre>
            </ScrollArea>
          )}
          
          <DialogFooter>
            {editingDocument ? (
              <>
                <Button variant="outline" onClick={() => setEditingDocument(false)}>
                  Cancel
                </Button>
                <Button onClick={handleUpdateDocument}>
                  Save Changes
                </Button>
              </>
            ) : (
              <>
                <Button 
                  variant="outline" 
                  onClick={() => {
                    // Copy clean document structure
                    const { 
                      _id, id, collection, created_by, updated_by, 
                      created_at, updated_at, version, _version,
                      _created_at, _updated_at, _created_by, _updated_by,
                      ...dataOnly 
                    } = selectedDocument || {}
                    
                    const displayDoc = {
                      ...dataOnly,
                      _metadata: {
                        id: _id || id,
                        collection,
                        version: version || _version,
                        created_at: created_at || _created_at,
                        updated_at: updated_at || _updated_at,
                        created_by: created_by || _created_by,
                        updated_by: updated_by || _updated_by,
                      }
                    }
                    
                    navigator.clipboard.writeText(JSON.stringify(displayDoc, null, 2))
                    toast({
                      title: "Copied",
                      description: "Document copied to clipboard",
                    })
                  }}
                >
                  <Copy className="h-4 w-4 mr-2" />
                  Copy
                </Button>
                <Button 
                  variant="outline"
                  onClick={() => setEditingDocument(true)}
                >
                  <Edit className="h-4 w-4 mr-2" />
                  Edit
                </Button>
                <Button onClick={() => setDocumentViewDialogOpen(false)}>
                  Close
                </Button>
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create Index Dialog */}
      <Dialog open={indexDialogOpen} onOpenChange={setIndexDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Index</DialogTitle>
            <DialogDescription>
              Add a new index to improve query performance
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="indexName">Index Name</Label>
              <Input
                id="indexName"
                placeholder="idx_user_email"
                value={newIndex.name}
                onChange={(e) => setNewIndex({ ...newIndex, name: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="fields">Fields</Label>
              <Input
                id="fields"
                placeholder="email:1, createdAt:-1"
                value={newIndex.fields}
                onChange={(e) => setNewIndex({ ...newIndex, fields: e.target.value })}
              />
              <p className="text-xs text-muted-foreground">
                Format: fieldName:direction (1 for ascending, -1 for descending). Separate multiple fields with commas.
              </p>
            </div>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="unique">Unique Index</Label>
                  <p className="text-xs text-muted-foreground">
                    Ensures all values are unique
                  </p>
                </div>
                <Switch
                  id="unique"
                  checked={newIndex.unique}
                  onCheckedChange={(checked) => setNewIndex({ ...newIndex, unique: checked })}
                />
              </div>
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="sparse">Sparse Index</Label>
                  <p className="text-xs text-muted-foreground">
                    Only includes documents with the indexed field
                  </p>
                </div>
                <Switch
                  id="sparse"
                  checked={newIndex.sparse}
                  onCheckedChange={(checked) => setNewIndex({ ...newIndex, sparse: checked })}
                />
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIndexDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateIndex}>
              Create Index
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}