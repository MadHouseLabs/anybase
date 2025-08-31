import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Shield, UserCheck, UserX, Clock, Calendar, User } from "lucide-react";
import { UserActionsCell } from "./users-client-components";
import { format } from 'date-fns';

interface UserType {
  _id?: string;
  id?: string;
  email: string;
  first_name?: string;
  last_name?: string;
  role: string;
  active: boolean;
  last_login?: string;
  created_at?: string;
}

interface UsersTableProps {
  users: UserType[];
  canEditUsers: boolean;
  currentUserEmail: string;
}

function getInitials(user: UserType) {
  const firstName = user.first_name || "";
  const lastName = user.last_name || "";
  const email = user.email || "";
  
  if (firstName && lastName) {
    return `${firstName[0]}${lastName[0]}`.toUpperCase();
  } else if (firstName) {
    return firstName.substring(0, 2).toUpperCase();
  } else if (email) {
    return email.substring(0, 2).toUpperCase();
  }
  return "U";
}

function getRoleBadgeVariant(role: string) {
  switch(role) {
    case "admin": return "destructive";
    case "developer": return "default";
    default: return "outline";
  }
}

function getRoleIcon(role: string) {
  switch(role) {
    case "admin": return Shield;
    case "developer": return User;
    default: return User;
  }
}

export function UsersTable({ users, canEditUsers, currentUserEmail }: UsersTableProps) {
  return (
    <div className="border">
      <Table>
        <TableHeader>
          <TableRow className="bg-muted/50">
            <TableHead className="font-semibold">User</TableHead>
            <TableHead className="font-semibold">Role</TableHead>
            <TableHead className="font-semibold">Status</TableHead>
            <TableHead className="font-semibold">Last Activity</TableHead>
            <TableHead className="font-semibold">Joined</TableHead>
            {canEditUsers && <TableHead className="text-right font-semibold">Actions</TableHead>}
          </TableRow>
        </TableHeader>
        <TableBody>
          {users.length === 0 ? (
            <TableRow>
              <TableCell colSpan={canEditUsers ? 6 : 5} className="text-center py-12">
                <div className="flex flex-col items-center justify-center space-y-2">
                  <div className="p-3 bg-muted">
                    <UserX className="h-8 w-8 text-muted-foreground" />
                  </div>
                  <p className="text-lg font-medium text-muted-foreground">No users found</p>
                  <p className="text-sm text-muted-foreground">Start by creating your first user</p>
                </div>
              </TableCell>
            </TableRow>
          ) : (
            users.map((user) => {
              const RoleIcon = getRoleIcon(user.role);
              const isCurrentUser = user.email === currentUserEmail;
              
              return (
                <TableRow key={user._id || user.id} className="hover:bg-muted/30 transition-colors">
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <Avatar className="h-10 w-10 border-2 rounded-none">
                        <AvatarImage src={`https://api.dicebear.com/7.x/initials/svg?seed=${user.email}&backgroundColor=3b82f6,10b981,8b5cf6,f59e0b`} />
                        <AvatarFallback className="font-semibold bg-primary/10 rounded-none">
                          {getInitials(user)}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium flex items-center gap-2">
                          {user.first_name || user.last_name ? 
                            `${user.first_name || ''} ${user.last_name || ''}`.trim() : 
                            'Unnamed User'
                          }
                          {isCurrentUser && (
                            <Badge variant="outline" className="text-xs">You</Badge>
                          )}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          {user.email}
                        </div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge 
                      variant={getRoleBadgeVariant(user.role)} 
                      className="font-medium px-3 py-1"
                    >
                      <RoleIcon className="h-3 w-3 mr-1.5" />
                      {user.role ? user.role.charAt(0).toUpperCase() + user.role.slice(1) : 'User'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge 
                      variant={user.active ? "default" : "secondary"}
                      className={user.active ? 
                        "bg-green-50 text-green-700 hover:bg-green-100 border-green-200" : 
                        "bg-gray-100 text-gray-700 hover:bg-gray-200"
                      }
                    >
                      {user.active ? (
                        <>
                          <UserCheck className="h-3 w-3 mr-1.5" />
                          Active
                        </>
                      ) : (
                        <>
                          <UserX className="h-3 w-3 mr-1.5" />
                          Inactive
                        </>
                      )}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                      <Clock className="h-3 w-3" />
                      {user.last_login ? 
                        format(new Date(user.last_login), 'MMM d, yyyy') : 
                        "Never logged in"
                      }
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                      <Calendar className="h-3 w-3" />
                      {user.created_at ? 
                        format(new Date(user.created_at), 'MMM d, yyyy') : 
                        "-"
                      }
                    </div>
                  </TableCell>
                  {canEditUsers && (
                    <TableCell className="text-right">
                      <UserActionsCell 
                        user={user}
                        disabled={isCurrentUser}
                      />
                    </TableCell>
                  )}
                </TableRow>
              );
            })
          )}
        </TableBody>
      </Table>
    </div>
  );
}