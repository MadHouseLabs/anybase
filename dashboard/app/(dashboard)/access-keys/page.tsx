"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from "@/components/ui/dialog"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { accessKeysApi } from "@/lib/accesskeys"
import { Key, Plus, Trash2, RefreshCw, Copy, Info, AlertCircle, Clock, X, CheckCircle, Shield, Activity, Calendar, Hash, Search, Filter, Lock, Unlock } from "lucide-react"
import { useToast } from "@/hooks/use-toast"
import { format } from 'date-fns'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import Cookies from "js-cookie"

export default function AccessKeysPage() {
  const { toast } = useToast()
  const [accessKeys, setAccessKeys] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [generatedKeys, setGeneratedKeys] = useState<Record<string, string>>({})
  const [copiedKey, setCopiedKey] = useState<string | null>(null)
  const [customPermissionInput, setCustomPermissionInput] = useState("")
  const [searchQuery, setSearchQuery] = useState("")
  const [statusFilter, setStatusFilter] = useState("all")
  const [newKey, setNewKey] = useState({
    name: "",
    description: "",
    permissions: [] as string[],
    expires_in: 0
  })

  // Check if current user has access (admin or developer)
  const currentUser = JSON.parse(Cookies.get("user") || "{}")
  const hasAccess = currentUser.role === "admin" || currentUser.role === "developer"

  useEffect(() => {
    if (hasAccess) {
      loadAccessKeys()
    } else {
      setLoading(false)
    }
  }, [hasAccess])

  const loadAccessKeys = async () => {
    try {
      const response = await accessKeysApi.list()
      setAccessKeys(response.access_keys || [])
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load access keys",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleCreateKey = async () => {
    if (!newKey.name || newKey.permissions.length === 0) {
      toast({
        title: "Error",
        description: "Key name and at least one permission are required",
        variant: "destructive",
      })
      return
    }

    try {
      const response = await accessKeysApi.create(newKey)
      if (response.key) {
        setGeneratedKeys(prev => ({ ...prev, [response.id]: response.key! }))
      }
      toast({
        title: "Success",
        description: `Access key ${newKey.name} created successfully`,
      })
      setCreateDialogOpen(false)
      setNewKey({ name: "", description: "", permissions: [], expires_in: 0 })
      setCustomPermissionInput("")
      loadAccessKeys()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to create access key",
        variant: "destructive",
      })
    }
  }

  const handleRegenerateKey = async (id: string, name: string) => {
    if (!confirm(`Are you sure you want to regenerate the key for ${name}? The old key will stop working immediately.`)) return

    try {
      const response = await accessKeysApi.regenerate(id)
      setGeneratedKeys(prev => ({ ...prev, [id]: response.key }))
      toast({
        title: "Success",
        description: `Access key ${name} regenerated successfully`,
      })
      loadAccessKeys()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to regenerate key",
        variant: "destructive",
      })
    }
  }

  const handleDeleteKey = async (id: string, name: string) => {
    if (!confirm(`Are you sure you want to delete the access key ${name}? This action cannot be undone.`)) return

    try {
      await accessKeysApi.delete(id)
      toast({
        title: "Success",
        description: `Access key ${name} deleted successfully`,
      })
      // Remove from generated keys if it exists
      setGeneratedKeys(prev => {
        const newKeys = { ...prev }
        delete newKeys[id]
        return newKeys
      })
      loadAccessKeys()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to delete access key",
        variant: "destructive",
      })
    }
  }

  const handleToggleActive = async (id: string, active: boolean, name: string) => {
    try {
      await accessKeysApi.update(id, { active: !active })
      toast({
        title: "Success",
        description: `Access key ${name} ${!active ? 'enabled' : 'disabled'} successfully`,
      })
      loadAccessKeys()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to update access key",
        variant: "destructive",
      })
    }
  }

  const copyToClipboard = (text: string, keyId: string) => {
    navigator.clipboard.writeText(text)
    setCopiedKey(keyId)
    setTimeout(() => setCopiedKey(null), 2000)
    toast({
      title: "Copied",
      description: "Access key copied to clipboard",
    })
  }

  const addCustomPermission = () => {
    const trimmed = customPermissionInput.trim()
    if (trimmed && !newKey.permissions.includes(trimmed)) {
      setNewKey(prev => ({
        ...prev,
        permissions: [...prev.permissions, trimmed]
      }))
      setCustomPermissionInput("")
    }
  }

  const removePermission = (permission: string) => {
    setNewKey(prev => ({
      ...prev,
      permissions: prev.permissions.filter(p => p !== permission)
    }))
  }

  if (!hasAccess) {
    return (
      <div className="container mx-auto py-6">
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Only administrators and developers can manage access keys.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  // Filter keys based on search and status
  const filteredKeys = accessKeys.filter(key => {
    const matchesSearch = searchQuery === "" || 
      key.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      key.description?.toLowerCase().includes(searchQuery.toLowerCase())
    
    const matchesStatus = statusFilter === "all" || 
      (statusFilter === "active" && key.active) ||
      (statusFilter === "inactive" && !key.active) ||
      (statusFilter === "expired" && key.expires_at && new Date(key.expires_at) < new Date())
    
    return matchesSearch && matchesStatus
  })

  // Calculate statistics
  const stats = {
    total: accessKeys.length,
    active: accessKeys.filter(k => k.active).length,
    inactive: accessKeys.filter(k => !k.active).length,
    expiring: accessKeys.filter(k => {
      if (!k.expires_at) return false
      const daysUntilExpiry = Math.ceil((new Date(k.expires_at).getTime() - Date.now()) / (1000 * 60 * 60 * 24))
      return daysUntilExpiry > 0 && daysUntilExpiry <= 30
    }).length
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Key className="h-8 w-8" />
            Access Keys Management
          </h1>
          <p className="text-muted-foreground mt-1">
            Create and manage API access keys for programmatic access to your resources
          </p>
        </div>
        <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button size="lg">
              <Plus className="h-5 w-5 mr-2" />
              Create Access Key
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-2xl">
            <DialogHeader>
              <DialogTitle>Create Access Key</DialogTitle>
              <DialogDescription>
                Create a new API access key with specific permissions
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Key Name</Label>
                <Input
                  id="name"
                  placeholder="e.g., CI/CD Pipeline"
                  value={newKey.name}
                  onChange={(e) => setNewKey(prev => ({ ...prev, name: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="description">Description</Label>
                <Textarea
                  id="description"
                  placeholder="Describe the purpose of this access key"
                  value={newKey.description}
                  onChange={(e) => setNewKey(prev => ({ ...prev, description: e.target.value }))}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="expires">Expiration (hours)</Label>
                <Input
                  id="expires"
                  type="number"
                  placeholder="0 for no expiration"
                  value={newKey.expires_in}
                  onChange={(e) => setNewKey(prev => ({ ...prev, expires_in: parseInt(e.target.value) || 0 }))}
                />
                <p className="text-xs text-muted-foreground">
                  Set to 0 for keys that never expire
                </p>
              </div>
              <div className="space-y-2">
                <Label>Permissions</Label>
                
                {/* Selected Permissions Display */}
                {newKey.permissions.length > 0 && (
                  <div className="mb-3">
                    <Label className="text-xs text-muted-foreground">Selected Permissions</Label>
                    <div className="flex flex-wrap gap-2 mt-2 p-3 bg-muted/30 rounded-md min-h-[60px]">
                      {newKey.permissions.map(perm => (
                        <Badge key={perm} variant="secondary" className="pl-2 pr-1 py-1">
                          <span className="text-xs font-mono">{perm}</span>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-4 w-4 ml-1 hover:bg-transparent"
                            onClick={() => removePermission(perm)}
                          >
                            <X className="h-3 w-3" />
                          </Button>
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}

                {/* Custom Permission Input */}
                <div className="flex gap-2 mb-3">
                  <Input
                    placeholder="Add permission (type:name:action, e.g., collection:products:read)"
                    value={customPermissionInput}
                    onChange={(e) => setCustomPermissionInput(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addCustomPermission())}
                    className="flex-1"
                  />
                  <Button
                    type="button"
                    size="sm"
                    onClick={addCustomPermission}
                    disabled={!customPermissionInput.trim()}
                  >
                    Add
                  </Button>
                </div>

                <div className="space-y-2">
                  <Label className="text-xs text-muted-foreground">Common Permission Patterns</Label>
                  <div className="text-xs text-muted-foreground space-y-1 p-3 bg-muted/30 rounded-md">
                    <div><code>collection:products:read</code> - Read products collection</div>
                    <div><code>collection:*:read</code> - Read all collections</div>
                    <div><code>view:dashboard:execute</code> - Execute dashboard view</div>
                    <div><code>api:users:*</code> - All actions on users API</div>
                    <div><code>*:*:*</code> - Full access (use with caution)</div>
                  </div>
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => {
                setCreateDialogOpen(false)
                setNewKey({ name: "", description: "", permissions: [], expires_in: 0 })
                setCustomPermissionInput("")
              }}>
                Cancel
              </Button>
              <Button onClick={handleCreateKey}>
                Create Access Key
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* Statistics Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Keys</CardTitle>
            <Hash className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total}</div>
            <p className="text-xs text-muted-foreground">API access keys</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active</CardTitle>
            <Unlock className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.active}</div>
            <p className="text-xs text-muted-foreground">Currently enabled</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Inactive</CardTitle>
            <Lock className="h-4 w-4 text-orange-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.inactive}</div>
            <p className="text-xs text-muted-foreground">Temporarily disabled</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Expiring Soon</CardTitle>
            <Clock className="h-4 w-4 text-yellow-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.expiring}</div>
            <p className="text-xs text-muted-foreground">Within 30 days</p>
          </CardContent>
        </Card>
      </div>

      {/* Security Notice */}
      <Alert className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-900">
        <Shield className="h-4 w-4 text-blue-600" />
        <AlertDescription className="text-blue-900 dark:text-blue-100">
          <strong>Security Best Practices:</strong> Never share or expose your API keys. Store them securely in environment variables or secret management systems. Rotate keys regularly and use the minimum required permissions.
        </AlertDescription>
      </Alert>

      {/* Main Content Card */}
      <Card>
        <CardHeader>
          <div className="space-y-4">
            <div>
              <CardTitle className="text-xl">API Keys</CardTitle>
              <CardDescription>
                Manage your API access keys and their permissions
              </CardDescription>
            </div>
            
            {/* Filters */}
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="flex-1">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                  <Input
                    placeholder="Search keys by name or description..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10"
                  />
                </div>
              </div>
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger className="w-[180px]">
                  <Filter className="h-4 w-4 mr-2" />
                  <SelectValue placeholder="Filter by status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Keys</SelectItem>
                  <SelectItem value="active">Active Only</SelectItem>
                  <SelectItem value="inactive">Inactive Only</SelectItem>
                  <SelectItem value="expired">Expired</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {filteredKeys.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12">
                <Key className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-lg font-medium mb-2">
                  {searchQuery || statusFilter !== "all" 
                    ? "No keys found matching your filters" 
                    : "No access keys yet"}
                </p>
                <p className="text-sm text-muted-foreground">
                  {searchQuery || statusFilter !== "all" 
                    ? "Try adjusting your search or filters" 
                    : "Create your first access key to get started"}
                </p>
              </div>
            ) : (
              filteredKeys.map((key) => (
                <div key={key.id} className="border rounded-lg p-6 space-y-4 hover:bg-muted/50 transition-colors">
                  <div className="flex items-start justify-between">
                    <div className="space-y-1">
                      <div className="flex items-center gap-3">
                        <div className="p-2 bg-primary/10 rounded-lg">
                          <Key className="h-5 w-5 text-primary" />
                        </div>
                        <div className="flex-1">
                          <h3 className="text-lg font-semibold">{key.name}</h3>
                        </div>
                        <Badge 
                          variant={key.active ? "default" : "secondary"}
                          className={key.active ? "bg-green-100 text-green-800 hover:bg-green-100" : ""}
                        >
                          {key.active ? (
                            <><Unlock className="h-3 w-3 mr-1" /> Active</>
                          ) : (
                            <><Lock className="h-3 w-3 mr-1" /> Inactive</>
                          )}
                        </Badge>
                        {key.expires_at && (
                          <Badge variant="outline" className="gap-1">
                            <Clock className="h-3 w-3" />
                            Expires {format(new Date(key.expires_at), 'MMM d, yyyy')}
                          </Badge>
                        )}
                      </div>
                      {key.description && (
                        <p className="text-sm text-muted-foreground">{key.description}</p>
                      )}
                      <div className="flex items-center gap-4 text-xs text-muted-foreground pt-2">
                        <span className="flex items-center gap-1">
                          <Calendar className="h-3 w-3" />
                          Created {format(new Date(key.created_at), 'MMM d, yyyy')}
                        </span>
                        {key.last_used && (
                          <span className="flex items-center gap-1">
                            <Activity className="h-3 w-3" />
                            Last used {format(new Date(key.last_used), 'MMM d, yyyy')}
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleToggleActive(key.id, key.active, key.name)}
                        className={key.active ? "hover:bg-orange-50" : "hover:bg-green-50"}
                      >
                        {key.active ? (
                          <><Lock className="h-4 w-4 mr-1" /> Disable</>
                        ) : (
                          <><Unlock className="h-4 w-4 mr-1" /> Enable</>
                        )}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleRegenerateKey(key.id, key.name)}
                        className="hover:bg-blue-50"
                      >
                        <RefreshCw className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDeleteKey(key.id, key.name)}
                        className="hover:bg-red-50"
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </div>

                  {/* Show generated key if it exists */}
                  {generatedKeys[key.id] && (
                    <Alert className="bg-muted/50">
                      <Key className="h-4 w-4" />
                      <AlertDescription>
                        <div className="space-y-2">
                          <p className="font-medium">New Access Key Generated:</p>
                          <div className="flex items-center gap-2">
                            <code className="flex-1 p-2 bg-background rounded text-sm font-mono break-all">
                              {generatedKeys[key.id]}
                            </code>
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => copyToClipboard(generatedKeys[key.id], key.id)}
                            >
                              {copiedKey === key.id ? (
                                <>
                                  <CheckCircle className="h-4 w-4 mr-2" />
                                  Copied
                                </>
                              ) : (
                                <>
                                  <Copy className="h-4 w-4 mr-2" />
                                  Copy
                                </>
                              )}
                            </Button>
                          </div>
                          <p className="text-xs text-muted-foreground">
                            Save this key securely. You won't be able to see it again.
                          </p>
                        </div>
                      </AlertDescription>
                    </Alert>
                  )}

                  {/* Permissions */}
                  <div>
                    <Label className="text-xs text-muted-foreground">Permissions</Label>
                    <div className="flex flex-wrap gap-1 mt-2">
                      {key.permissions?.slice(0, 5).map((perm: string) => (
                        <Badge key={perm} variant="outline" className="text-xs font-mono">
                          <Shield className="h-3 w-3 mr-1" />
                          {perm}
                        </Badge>
                      ))}
                      {key.permissions?.length > 5 && (
                        <Badge variant="outline" className="text-xs">
                          +{key.permissions.length - 5} more
                        </Badge>
                      )}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
          {filteredKeys.length > 0 && (
            <div className="px-6 py-4 border-t text-sm text-muted-foreground">
              Showing {filteredKeys.length} of {accessKeys.length} keys
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}