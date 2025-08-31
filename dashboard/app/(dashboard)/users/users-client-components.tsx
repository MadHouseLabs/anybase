"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { createUser, updateUser, deleteUser } from "@/app/actions/users-actions";
import { Plus, Edit, Trash2, Loader2, Search, Filter, Activity } from "lucide-react";

interface User {
  _id?: string;
  id?: string;
  email: string;
  first_name?: string;
  last_name?: string;
  role: string;
  active: boolean;
  last_login?: string;
}

export function CreateUserButton({ canEditUsers }: { canEditUsers: boolean }) {
  const { toast } = useToast();
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [newUser, setNewUser] = useState({
    first_name: "",
    last_name: "",
    email: "",
    password: "",
    role: "developer",
    active: true
  });

  if (!canEditUsers) return null;

  const handleCreateUser = async () => {
    if (!newUser.email || !newUser.password) {
      toast({
        title: "Error",
        description: "Email and password are required",
        variant: "destructive",
      });
      return;
    }
    
    setIsCreating(true);
    const result = await createUser(newUser);
    
    if (result.success) {
      toast({
        title: "Success",
        description: "User created successfully",
      });
      setCreateDialogOpen(false);
      setNewUser({
        first_name: "",
        last_name: "",
        email: "",
        password: "",
        role: "developer",
        active: true
      });
    } else {
      toast({
        title: "Error",
        description: result.error || "Failed to create user",
        variant: "destructive",
      });
    }
    setIsCreating(false);
  };

  return (
    <>
      <Button onClick={() => setCreateDialogOpen(true)} size="lg">
        <Plus className="h-5 w-5 mr-2" />
        Add User
      </Button>

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
                  disabled={isCreating}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="last_name">Last Name</Label>
                <Input
                  id="last_name"
                  value={newUser.last_name}
                  onChange={(e) => setNewUser({...newUser, last_name: e.target.value})}
                  placeholder="Doe"
                  disabled={isCreating}
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
                disabled={isCreating}
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
                disabled={isCreating}
              />
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="new_role">Role</Label>
              <Select
                value={newUser.role}
                onValueChange={(value) => setNewUser({...newUser, role: value})}
                disabled={isCreating}
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
                disabled={isCreating}
              />
              <Label htmlFor="new_active">Active Account</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setCreateDialogOpen(false);
              setNewUser({
                first_name: "",
                last_name: "",
                email: "",
                password: "",
                role: "developer",
                active: true
              });
            }} disabled={isCreating}>
              Cancel
            </Button>
            <Button onClick={handleCreateUser} disabled={isCreating}>
              {isCreating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating...
                </>
              ) : (
                "Create User"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

export function UserActionsCell({ user, disabled }: { user: User; disabled: boolean }) {
  const { toast } = useToast();
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [editData, setEditData] = useState({
    role: user.role || "developer",
    active: user.active
  });

  const handleUpdateUser = async () => {
    setIsUpdating(true);
    const result = await updateUser(user._id || user.id!, editData);
    
    if (result.success) {
      toast({
        title: "Success",
        description: "User updated successfully",
      });
      setEditDialogOpen(false);
    } else {
      toast({
        title: "Error",
        description: result.error || "Failed to update user",
        variant: "destructive",
      });
    }
    setIsUpdating(false);
  };

  const handleDeleteUser = async () => {
    if (!confirm(`Are you sure you want to delete user ${user.email}?`)) return;

    setIsDeleting(true);
    const result = await deleteUser(user._id || user.id!);
    
    if (result.success) {
      toast({
        title: "Success",
        description: "User deleted successfully",
      });
    } else {
      toast({
        title: "Error",
        description: result.error || "Failed to delete user",
        variant: "destructive",
      });
    }
    setIsDeleting(false);
  };

  return (
    <div className="flex justify-end gap-2">
      <Button
        variant="ghost"
        size="sm"
        onClick={() => {
          setEditData({ role: user.role || "developer", active: user.active });
          setEditDialogOpen(true);
        }}
        className="hover:bg-blue-50"
      >
        <Edit className="h-4 w-4" />
      </Button>
      
      <Button
        variant="ghost"
        size="sm"
        onClick={handleDeleteUser}
        disabled={disabled || isDeleting}
        className="hover:bg-red-50"
      >
        {isDeleting ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <Trash2 className="h-4 w-4 text-destructive" />
        )}
      </Button>

      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit User</DialogTitle>
            <DialogDescription>
              Update user settings and access level
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Email</Label>
              <p className="text-sm text-muted-foreground">{user.email}</p>
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="role">Role</Label>
              <Select
                value={editData.role}
                onValueChange={(value) => setEditData({...editData, role: value})}
                disabled={isUpdating}
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
                checked={editData.active}
                onCheckedChange={(checked) => setEditData({...editData, active: checked})}
                disabled={isUpdating}
              />
              <Label htmlFor="active">Active Account</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)} disabled={isUpdating}>
              Cancel
            </Button>
            <Button onClick={handleUpdateUser} disabled={isUpdating}>
              {isUpdating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                "Save Changes"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export function UsersSearchFilter({ onSearch, onRoleFilter, onStatusFilter }: {
  onSearch: (value: string) => void;
  onRoleFilter: (value: string) => void;
  onStatusFilter: (value: string) => void;
}) {
  return (
    <div className="flex flex-col sm:flex-row gap-4">
      <div className="flex-1">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
          <Input
            placeholder="Search by name or email..."
            onChange={(e) => onSearch(e.target.value)}
            className="pl-10"
          />
        </div>
      </div>
      <div className="flex gap-2">
        <Select onValueChange={onRoleFilter} defaultValue="all">
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
        <Select onValueChange={onStatusFilter} defaultValue="all">
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
  );
}