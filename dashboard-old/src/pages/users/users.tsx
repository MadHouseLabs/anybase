import { useState, useEffect } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useToast } from "@/hooks/use-toast"
import { usersApi } from "@/lib/api"
import { Shield, Trash2, Edit, UserCheck, UserX, AlertCircle, Plus, Users, Mail, Clock, ShieldCheck, Activity, Search, Filter } from "lucide-react"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"

export function UsersPage() {
  const { toast } = useToast()
  const [users, setUsers] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedUser, setSelectedUser] = useState<any>(null)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState("")
  const [roleFilter, setRoleFilter] = useState("all")
  const [statusFilter, setStatusFilter] = useState("all")
  const [newUser, setNewUser] = useState({
    first_name: "",
    last_name: "",
    email: "",
    password: "",
    role: "developer",
    active: true
  })
  
  // Check if current user is admin or developer
  const currentUser = JSON.parse(localStorage.getItem("user") || "{}")
  const isAdmin = currentUser.role === "admin"
  const isDeveloper = currentUser.role === "developer"
  const canViewUsers = isAdmin || isDeveloper
  const canEditUsers = isAdmin // Only admins can edit

  useEffect(() => {
    if (canViewUsers) {
      loadData()
    }
  }, [canViewUsers])

  const loadData = async () => {
    try {
      const usersRes = await usersApi.list()
      setUsers(usersRes.users || [])
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to load users",
        variant: "destructive",
      })
    } finally {
      setLoading(false)
    }
  }

  const handleCreateUser = async () => {
    if (!newUser.email || !newUser.password) {
      toast({
        title: "Error",
        description: "Email and password are required",
        variant: "destructive",
      })
      return
    }
    
    try {
      await usersApi.create(newUser)
      toast({
        title: "Success",
        description: "User created successfully",
      })
      setCreateDialogOpen(false)
      setNewUser({
        first_name: "",
        last_name: "",
        email: "",
        password: "",
        role: "developer",
        active: true
      })
      loadData()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to create user",
        variant: "destructive",
      })
    }
  }

  const handleUpdateUser = async () => {
    if (!selectedUser) return
    
    try {
      await usersApi.update(selectedUser._id || selectedUser.id, {
        active: selectedUser.active,
        role: selectedUser.role
      })
      toast({
        title: "Success",
        description: "User updated successfully",
      })
      setEditDialogOpen(false)
      setSelectedUser(null)
      loadData()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to update user",
        variant: "destructive",
      })
    }
  }

  const handleDeleteUser = async (userId: string, userName: string) => {
    if (!confirm(`Are you sure you want to delete user ${userName}?`)) return

    try {
      await usersApi.delete(userId)
      toast({
        title: "Success",
        description: "User deleted successfully",
      })
      loadData()
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || "Failed to delete user",
        variant: "destructive",
      })
    }
  }

  const getRoleBadgeVariant = (role: string) => {
    switch(role) {
      case "admin": return "destructive"
      case "developer": return "default"
      default: return "outline"
    }
  }

  const getInitials = (user: any) => {
    const firstName = user.first_name || ""
    const lastName = user.last_name || ""
    const email = user.email || ""
    
    if (firstName && lastName) {
      return `${firstName[0]}${lastName[0]}`.toUpperCase()
    } else if (firstName) {
      return firstName.substring(0, 2).toUpperCase()
    } else if (email) {
      return email.substring(0, 2).toUpperCase()
    }
    return "U"
  }

  const filteredUsers = users.filter(user => {
    const matchesSearch = searchQuery === "" || 
      user.email?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.first_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.last_name?.toLowerCase().includes(searchQuery.toLowerCase())
    
    const matchesRole = roleFilter === "all" || user.role === roleFilter
    const matchesStatus = statusFilter === "all" || 
      (statusFilter === "active" && user.active) ||
      (statusFilter === "inactive" && !user.active)
    
    return matchesSearch && matchesRole && matchesStatus
  })

  const stats = {
    total: users.length,
    admins: users.filter(u => u.role === "admin").length,
    developers: users.filter(u => u.role === "developer").length,
    active: users.filter(u => u.active).length,
    inactive: users.filter(u => !u.active).length
  }

  if (!canViewUsers) {
    return (
      <div className="container mx-auto py-6">
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            You don't have permission to view users.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Users className="h-8 w-8" />
            User Management
          </h1>
          <p className="text-muted-foreground mt-1">
            {canEditUsers ? "Manage system users and their access permissions" : "View system users (read-only access)"}
          </p>
        </div>
        {canEditUsers && (
          <Button onClick={() => setCreateDialogOpen(true)} size="lg">
            <Plus className="h-5 w-5 mr-2" />
            Add User
          </Button>
        )}
      </div>

      {/* Statistics Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total}</div>
            <p className="text-xs text-muted-foreground">Registered accounts</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Administrators</CardTitle>
            <ShieldCheck className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.admins}</div>
            <p className="text-xs text-muted-foreground">Full access users</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Developers</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.developers}</div>
            <p className="text-xs text-muted-foreground">Standard users</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active</CardTitle>
            <UserCheck className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.active}</div>
            <p className="text-xs text-muted-foreground">Can access system</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Inactive</CardTitle>
            <UserX className="h-4 w-4 text-red-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.inactive}</div>
            <p className="text-xs text-muted-foreground">Access disabled</p>
          </CardContent>
        </Card>
      </div>

      {/* Users Table */}
      <Card>
        <CardHeader>
          <div className="space-y-4">
            <div>
              <CardTitle className="text-xl">All Users</CardTitle>
              <CardDescription>
                Complete list of registered users and their access levels
              </CardDescription>
            </div>
            
            {/* Filters */}
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="flex-1">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                  <Input
                    placeholder="Search by name or email..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10"
                  />
                </div>
              </div>
              <div className="flex gap-2">
                <Select value={roleFilter} onValueChange={setRoleFilter}>
                  <SelectTrigger className="w-[140px]">
                    <Filter className="h-4 w-4 mr-2" />
                    <SelectValue placeholder="Role" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Roles</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                    <SelectItem value="developer">Developer</SelectItem>
                  </SelectContent>
                </Select>
                <Select value={statusFilter} onValueChange={setStatusFilter}>
                  <SelectTrigger className="w-[140px]">
                    <Activity className="h-4 w-4 mr-2" />
                    <SelectValue placeholder="Status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Status</SelectItem>
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="inactive">Inactive</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>User</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Last Login</TableHead>
                {canEditUsers && <TableHead className="text-right">Actions</TableHead>}
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredUsers.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={canEditUsers ? 6 : 5} className="text-center py-8 text-muted-foreground">
                    {searchQuery || roleFilter !== "all" || statusFilter !== "all" 
                      ? "No users found matching your filters" 
                      : "No users registered yet"}
                  </TableCell>
                </TableRow>
              ) : (
                filteredUsers.map((user) => (
                  <TableRow key={user._id || user.id} className="hover:bg-muted/50">
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <Avatar className="h-10 w-10">
                          <AvatarImage src={`https://api.dicebear.com/7.x/initials/svg?seed=${user.email}`} />
                          <AvatarFallback>{getInitials(user)}</AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="font-medium">
                            {user.first_name || user.last_name ? 
                              `${user.first_name || ''} ${user.last_name || ''}`.trim() : 
                              user.username || 'Unnamed User'
                            }
                          </div>
                          <div className="text-sm text-muted-foreground flex items-center gap-1">
                            <Mail className="h-3 w-3" />
                            {user.email}
                          </div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{user.email}</TableCell>
                    <TableCell>
                      <Badge variant={getRoleBadgeVariant(user.role)} className="font-semibold">
                        <Shield className="h-3 w-3 mr-1" />
                        {user.role ? user.role.toUpperCase() : 'DEVELOPER'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge 
                        variant={user.active ? "default" : "secondary"}
                        className={user.active ? "bg-green-100 text-green-800 hover:bg-green-100" : ""}
                      >
                        {user.active ? <UserCheck className="h-3 w-3 mr-1" /> : <UserX className="h-3 w-3 mr-1" />}
                        {user.active ? "Active" : "Inactive"}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1 text-sm">
                        <Clock className="h-3 w-3 text-muted-foreground" />
                        {user.last_login ? 
                          new Date(user.last_login).toLocaleDateString('en-US', { 
                            month: 'short', 
                            day: 'numeric', 
                            year: 'numeric' 
                          }) : 
                          "Never"
                        }
                      </div>
                    </TableCell>
                  {canEditUsers && (
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => {
                            setSelectedUser(user)
                            setEditDialogOpen(true)
                          }}
                          className="hover:bg-blue-50"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDeleteUser(user._id || user.id, user.email)}
                          disabled={user.email === currentUser.email}
                          className="hover:bg-red-50"
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  )}
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
          {filteredUsers.length > 0 && (
            <div className="px-6 py-4 border-t text-sm text-muted-foreground">
              Showing {filteredUsers.length} of {users.length} users
            </div>
          )}
        </CardContent>
      </Card>

      {/* Edit User Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit User</DialogTitle>
            <DialogDescription>
              Update user settings and access level
            </DialogDescription>
          </DialogHeader>
          {selectedUser && (
            <div className="space-y-4">
              <div className="space-y-2">
                <Label>Email</Label>
                <p className="text-sm text-muted-foreground">{selectedUser.email}</p>
              </div>
              
              <div className="space-y-2">
                <Label htmlFor="role">Role</Label>
                <Select
                  value={selectedUser.role || "developer"}
                  onValueChange={(value) => setSelectedUser({...selectedUser, role: value})}
                >
                  <SelectTrigger id="role">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="admin">Admin</SelectItem>
                    <SelectItem value="developer">Developer</SelectItem>
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">
                  Admin: Full system access | Developer: Manage collections and views
                </p>
              </div>

              <div className="flex items-center space-x-2">
                <Switch
                  id="active"
                  checked={selectedUser.active}
                  onCheckedChange={(checked) => setSelectedUser({...selectedUser, active: checked})}
                />
                <Label htmlFor="active">Active Account</Label>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setEditDialogOpen(false)
              setSelectedUser(null)
            }}>
              Cancel
            </Button>
            <Button onClick={handleUpdateUser}>
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Create User Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create User</DialogTitle>
            <DialogDescription>
              Add a new user to the system
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="first_name">First Name</Label>
                <Input
                  id="first_name"
                  value={newUser.first_name}
                  onChange={(e) => setNewUser({...newUser, first_name: e.target.value})}
                  placeholder="John"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="last_name">Last Name</Label>
                <Input
                  id="last_name"
                  value={newUser.last_name}
                  onChange={(e) => setNewUser({...newUser, last_name: e.target.value})}
                  placeholder="Doe"
                />
              </div>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="email">Email *</Label>
              <Input
                id="email"
                type="email"
                value={newUser.email}
                onChange={(e) => setNewUser({...newUser, email: e.target.value})}
                placeholder="john.doe@example.com"
                required
              />
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="password">Password *</Label>
              <Input
                id="password"
                type="password"
                value={newUser.password}
                onChange={(e) => setNewUser({...newUser, password: e.target.value})}
                placeholder="Minimum 6 characters"
                required
              />
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="new_role">Role</Label>
              <Select
                value={newUser.role}
                onValueChange={(value) => setNewUser({...newUser, role: value})}
              >
                <SelectTrigger id="new_role">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="developer">Developer</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Admin: Full system access | Developer: Manage collections and views
              </p>
            </div>

            <div className="flex items-center space-x-2">
              <Switch
                id="new_active"
                checked={newUser.active}
                onCheckedChange={(checked) => setNewUser({...newUser, active: checked})}
              />
              <Label htmlFor="new_active">Active Account</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setCreateDialogOpen(false)
              setNewUser({
                first_name: "",
                last_name: "",
                email: "",
                password: "",
                role: "developer",
                active: true
              })
            }}>
              Cancel
            </Button>
            <Button onClick={handleCreateUser}>
              Create User
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}