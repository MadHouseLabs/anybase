"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { collectionsApi } from "@/lib/api"
import { Plus, Database, Trash2, Edit, Eye, ArrowRight } from "lucide-react"
import { useToast } from "@/hooks/use-toast"

export default function CollectionsPage() {
  const { toast } = useToast()
  const router = useRouter()
  const [collections, setCollections] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [newCollection, setNewCollection] = useState({
    name: "",
    description: "",
  })

  useEffect(() => {
    loadCollections()
  }, [])

  const loadCollections = async () => {
    try {
      const response = await collectionsApi.list()
      setCollections(response.collections || [])
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load collections",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const createCollection = async () => {
    try {
      await collectionsApi.create({
        name: newCollection.name,
        description: newCollection.description,
        settings: {
          versioning: true,
          soft_delete: true,
          auditing: true,
        },
      })
      toast({
        title: "Success",
        description: "Collection created successfully",
      })
      setDialogOpen(false)
      setNewCollection({ name: "", description: "" })
      loadCollections()
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to create collection",
        variant: "destructive",
      })
    }
  }

  const deleteCollection = async (name: string) => {
    if (!confirm(`Are you sure you want to delete collection "${name}"?`)) return

    try {
      await collectionsApi.delete(name)
      toast({
        title: "Success",
        description: "Collection deleted successfully",
      })
      loadCollections()
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to delete collection",
        variant: "destructive",
      })
    }
  }

  return (
    <div className="container mx-auto py-6 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Collections</h1>
          <p className="text-muted-foreground">Manage your database collections</p>
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              New Collection
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Collection</DialogTitle>
              <DialogDescription>
                Add a new collection to your database
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input
                  id="name"
                  placeholder="products"
                  value={newCollection.name}
                  onChange={(e) => setNewCollection({ ...newCollection, name: e.target.value })}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  placeholder="Store product information"
                  value={newCollection.description}
                  onChange={(e) => setNewCollection({ ...newCollection, description: e.target.value })}
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                Cancel
              </Button>
              <Button onClick={createCollection}>Create</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {collections.map((collection) => (
          <Card 
            key={collection.name} 
            className="cursor-pointer hover:shadow-lg transition-shadow"
            onClick={() => router.push(`/collections/${collection.name}`)}
          >
            <CardHeader>
              <div className="flex items-center justify-between">
                <Database className="h-8 w-8 text-muted-foreground" />
                <div className="flex gap-1">
                  <Button 
                    variant="ghost" 
                    size="icon"
                    onClick={(e) => {
                      e.stopPropagation()
                      router.push(`/collections/${collection.name}`)
                    }}
                  >
                    <ArrowRight className="h-4 w-4" />
                  </Button>
                  <Button 
                    variant="ghost" 
                    size="icon"
                    onClick={(e) => {
                      e.stopPropagation()
                      deleteCollection(collection.name)
                    }}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              </div>
              <CardTitle>{collection.name}</CardTitle>
              <CardDescription>{collection.description || "No description"}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <div className="flex flex-wrap gap-2">
                  {collection.settings?.versioning && (
                    <Badge variant="secondary">Versioning</Badge>
                  )}
                  {collection.settings?.soft_delete && (
                    <Badge variant="secondary">Soft Delete</Badge>
                  )}
                  {collection.settings?.auditing && (
                    <Badge variant="secondary">Auditing</Badge>
                  )}
                </div>
                {collection.document_count !== undefined && (
                  <p className="text-sm text-muted-foreground">
                    {collection.document_count} documents
                  </p>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {collections.length === 0 && !loading && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Database className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-lg font-medium">No collections yet</p>
            <p className="text-sm text-muted-foreground">Create your first collection to get started</p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}