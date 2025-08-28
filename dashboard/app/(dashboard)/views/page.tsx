"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { useToast } from "@/hooks/use-toast"
import { viewsApi } from "@/lib/views"
import { collectionsApi } from "@/lib/api"
import { Eye, Plus, Database, Filter, SortAsc, Play, Edit, Trash2, Code, Search, AlertCircle, ChevronRight, Copy, CheckCircle, FileJson, Calendar, Shield } from "lucide-react"
import { format } from 'date-fns'
import Cookies from "js-cookie"

export default function ViewsPage() {
  const { toast } = useToast()
  const [views, setViews] = useState<any[]>([])
  const [collections, setCollections] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [executeDialogOpen, setExecuteDialogOpen] = useState(false)
  const [selectedView, setSelectedView] = useState<any>(null)
  const [queryResults, setQueryResults] = useState<any[]>([])
  const [searchQuery, setSearchQuery] = useState("")
  const [copiedQuery, setCopiedQuery] = useState<string | null>(null)
  
  const [newView, setNewView] = useState({
    name: "",
    description: "",
    collection: "",
    filter: "{}",
    fields: ""
  })
  
  const [editingView, setEditingView] = useState({
    name: "",
    description: "",
    collection: "",
    filter: "{}",
    fields: ""
  })
  
  // Runtime query options
  const [queryOptions, setQueryOptions] = useState({
    filter: "{}",
    sort: "{}",
    limit: "10",
    skip: "0"
  })

  // Check if current user has access (admin or developer)
  const currentUser = JSON.parse(Cookies.get("user") || "{}")
  const hasAccess = currentUser.role === "admin" || currentUser.role === "developer"

  useEffect(() => {
    if (hasAccess) {
      loadData()
    } else {
      setLoading(false)
    }
  }, [hasAccess])

  const loadData = async () => {
    try {
      const [viewsRes, collectionsRes] = await Promise.all([
        viewsApi.list(),
        collectionsApi.list()
      ])
      setViews(viewsRes.views || [])
      setCollections(collectionsRes.collections || [])
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load views",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleCreateView = async () => {
    if (!newView.name || !newView.collection) {
      toast({
        title: "Error",
        description: "View name and collection are required",
        variant: "destructive",
      })
      return
    }

    try {
      // Parse JSON fields
      let filter = {}
      let fields: string[] = []

      try {
        if (newView.filter.trim()) {
          filter = JSON.parse(newView.filter)
        }
      } catch {
        toast({
          title: "Error",
          description: "Invalid filter JSON",
          variant: "destructive",
        })
        return
      }

      if (newView.fields.trim()) {
        fields = newView.fields.split(',').map(f => f.trim()).filter(f => f)
      }

      await viewsApi.create({
        name: newView.name,
        description: newView.description,
        collection: newView.collection,
        filter: Object.keys(filter).length > 0 ? filter : undefined,
        fields: fields.length > 0 ? fields : undefined,
      })

      toast({
        title: "Success",
        description: `View "${newView.name}" created successfully`,
      })
      setCreateDialogOpen(false)
      setNewView({
        name: "",
        description: "",
        collection: "",
        filter: "{}",
        fields: ""
      })
      loadData()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to create view",
        variant: "destructive",
      })
    }
  }

  const handleEditView = (view: any) => {
    setEditingView({
      name: view.name,
      description: view.description || "",
      collection: view.collection,
      filter: JSON.stringify(view.filter || {}),
      fields: view.fields ? view.fields.join(", ") : ""
    })
    setEditDialogOpen(true)
  }

  const handleUpdateView = async () => {
    if (!editingView.name || !editingView.collection) {
      toast({
        title: "Error",
        description: "Name and collection are required",
        variant: "destructive",
      })
      return
    }

    try {
      let filter = {}
      let fields: string[] = []
      
      if (editingView.filter.trim() && editingView.filter !== "{}") {
        filter = JSON.parse(editingView.filter)
      }
      
      if (editingView.fields.trim()) {
        fields = editingView.fields.split(',').map(f => f.trim()).filter(f => f)
      }
      
      await viewsApi.update(editingView.name, {
        description: editingView.description,
        collection: editingView.collection,
        filter,
        fields: fields.length > 0 ? fields : undefined
      })
      
      toast({
        title: "Success",
        description: `View "${editingView.name}" updated successfully`,
      })
      setEditDialogOpen(false)
      loadData()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to update view",
        variant: "destructive",
      })
    }
  }

  const handleExecuteView = async (view: any, applyOptions = false) => {
    try {
      setSelectedView(view)
      setQueryResults([])
      if (!applyOptions) {
        // Reset query options when opening dialog
        setQueryOptions({
          filter: "{}",
          sort: "{}",
          limit: "10",
          skip: "0"
        })
        setExecuteDialogOpen(true)
        return
      }
      
      // Parse and apply runtime query options
      let extraFilter = {}
      let sort = {}
      let limit = 10
      let skip = 0
      
      try {
        if (queryOptions.filter.trim() && queryOptions.filter !== "{}") {
          extraFilter = JSON.parse(queryOptions.filter)
        }
        if (queryOptions.sort.trim() && queryOptions.sort !== "{}") {
          sort = JSON.parse(queryOptions.sort)
        }
        limit = parseInt(queryOptions.limit) || 10
        skip = parseInt(queryOptions.skip) || 0
      } catch (e) {
        toast({
          title: "Error",
          description: "Invalid JSON in filter or sort",
          variant: "destructive",
        })
        return
      }
      
      const results = await viewsApi.query(view.name, { 
        limit,
        skip,
        filter: Object.keys(extraFilter).length > 0 ? extraFilter : undefined,
        sort: Object.keys(sort).length > 0 ? sort : undefined
      })
      setQueryResults(results.documents || results.data || [])
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to execute view",
        variant: "destructive",
      })
    }
  }

  const handleDeleteView = async (name: string) => {
    if (!confirm(`Are you sure you want to delete the view "${name}"?`)) return

    try {
      await viewsApi.delete(name)
      toast({
        title: "Success",
        description: `View "${name}" deleted successfully`,
      })
      loadData()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to delete view",
        variant: "destructive",
      })
    }
  }

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text)
    setCopiedQuery(id)
    setTimeout(() => setCopiedQuery(null), 2000)
    toast({
      title: "Copied",
      description: "Query copied to clipboard",
    })
  }

  const generateQueryString = (view: any) => {
    const parts = []
    if (view.filter) parts.push(`filter: ${JSON.stringify(view.filter)}`)
    if (view.fields?.length) parts.push(`fields: [${view.fields.join(', ')}]`)
    if (view.sort) parts.push(`sort: ${JSON.stringify(view.sort)}`)
    return parts.join('\n')
  }

  if (!hasAccess) {
    return (
      <div className="container mx-auto py-6">
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Only administrators and developers can manage views.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  const filteredViews = views.filter(view => 
    searchQuery === "" || 
    view.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    view.collection?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    view.description?.toLowerCase().includes(searchQuery.toLowerCase())
  )

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Eye className="h-8 w-8" />
            Views Management
          </h1>
          <p className="text-muted-foreground mt-1">
            Create and manage saved queries for your collections
          </p>
        </div>
        <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button size="lg">
              <Plus className="h-5 w-5 mr-2" />
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
                    value={newView.name}
                    onChange={(e) => setNewView(prev => ({ ...prev, name: e.target.value }))}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="collection">Collection</Label>
                  <Select
                    value={newView.collection}
                    onValueChange={(value) => setNewView(prev => ({ ...prev, collection: value }))}
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
                  value={newView.description}
                  onChange={(e) => setNewView(prev => ({ ...prev, description: e.target.value }))}
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
                    value={newView.filter}
                    onChange={(e) => setNewView(prev => ({ ...prev, filter: e.target.value }))}
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
                    value={newView.fields}
                    onChange={(e) => setNewView(prev => ({ ...prev, fields: e.target.value }))}
                  />
                  <p className="text-xs text-muted-foreground">
                    Comma-separated list of fields to include in results (leave empty for all fields)
                  </p>
                </TabsContent>
                
              </Tabs>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => {
                setCreateDialogOpen(false)
                setNewView({
                  name: "",
                  description: "",
                  collection: "",
                  filter: "{}",
                  fields: ""
                })
              }}>
                Cancel
              </Button>
              <Button onClick={handleCreateView}>
                Create View
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Edit View Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit View: {editingView.name}</DialogTitle>
            <DialogDescription>
              Update the view configuration
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>View Name</Label>
                <Input
                  value={editingView.name}
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
                  value={editingView.collection}
                  onValueChange={(value) => setEditingView(prev => ({ ...prev, collection: value }))}
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
                value={editingView.description}
                onChange={(e) => setEditingView(prev => ({ ...prev, description: e.target.value }))}
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
                  value={editingView.filter}
                  onChange={(e) => setEditingView(prev => ({ ...prev, filter: e.target.value }))}
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
                  value={editingView.fields}
                  onChange={(e) => setEditingView(prev => ({ ...prev, fields: e.target.value }))}
                />
                <p className="text-xs text-muted-foreground">
                  Comma-separated list of fields to include in results (leave empty for all fields)
                </p>
              </TabsContent>
            </Tabs>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleUpdateView}>
              Update View
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Info Alert */}
      <Alert className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-900">
        <Eye className="h-4 w-4 text-blue-600" />
        <AlertDescription className="text-blue-900 dark:text-blue-100">
          <strong>About Views:</strong> Views are saved queries that allow you to filter and project data from a collection. 
          They provide a consistent way to access frequently used data subsets without writing the same queries repeatedly.
        </AlertDescription>
      </Alert>

      {/* Main Content */}
      <Card>
        <CardHeader>
          <div className="space-y-4">
            <div>
              <CardTitle className="text-xl">Saved Views</CardTitle>
              <CardDescription>
                Manage your collection views and saved queries
              </CardDescription>
            </div>
            
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
              <Input
                placeholder="Search views by name, collection, or description..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {filteredViews.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12">
                <Eye className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-lg font-medium mb-2">
                  {searchQuery ? "No views found" : "No views yet"}
                </p>
                <p className="text-sm text-muted-foreground">
                  {searchQuery ? "Try adjusting your search" : "Create your first view to get started"}
                </p>
              </div>
            ) : (
              filteredViews.map((view) => (
                <div key={view.id || view.name} className="border rounded-lg p-6 space-y-4 hover:bg-muted/50 transition-colors">
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
                      <Button
                        variant="default"
                        size="sm"
                        onClick={() => handleExecuteView(view)}
                      >
                        <Play className="h-4 w-4 mr-1" />
                        Execute
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleEditView(view)}
                      >
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyToClipboard(generateQueryString(view), view.id || view.name)}
                      >
                        {copiedQuery === (view.id || view.name) ? (
                          <CheckCircle className="h-4 w-4" />
                        ) : (
                          <Copy className="h-4 w-4" />
                        )}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDeleteView(view.name)}
                        className="hover:bg-red-50"
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </div>

                  {/* Query Details */}
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
              ))
            )}
          </div>
        </CardContent>
      </Card>

      {/* Execute View Dialog */}
      <Dialog open={executeDialogOpen} onOpenChange={setExecuteDialogOpen}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Execute View: {selectedView?.name}</DialogTitle>
            <DialogDescription>
              Query {selectedView?.collection} collection with runtime filters and pagination
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            {/* Query Options */}
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
                onClick={() => handleExecuteView(selectedView, true)}
                className="w-full"
              >
                <Play className="h-4 w-4 mr-2" />
                Execute Query
              </Button>
            </div>
            
            {/* Results */}
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
    </div>
  )
}