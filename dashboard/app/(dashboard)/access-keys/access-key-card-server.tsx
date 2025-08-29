import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { Key, Unlock, Lock, Clock, Calendar, Activity, Shield } from "lucide-react";
import { format } from 'date-fns';

interface AccessKey {
  id: string;
  name: string;
  description?: string;
  active: boolean;
  permissions?: string[];
  expires_at?: string;
  created_at: string;
  last_used?: string;
}

interface AccessKeyCardDisplayProps {
  accessKey: AccessKey;
  children?: React.ReactNode;
}

export function AccessKeyCardDisplay({ accessKey, children }: AccessKeyCardDisplayProps) {
  return (
    <div className="border p-6 space-y-4 hover:bg-muted/50 transition-colors">
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-primary/10">
              <Key className="h-5 w-5 text-primary" />
            </div>
            <div className="flex-1">
              <h3 className="text-lg font-semibold">{accessKey.name}</h3>
            </div>
            <Badge 
              variant={accessKey.active ? "default" : "secondary"}
              className={accessKey.active ? "bg-green-100 text-green-800 hover:bg-green-100" : ""}
            >
              {accessKey.active ? (
                <><Unlock className="h-3 w-3 mr-1" /> Active</>
              ) : (
                <><Lock className="h-3 w-3 mr-1" /> Inactive</>
              )}
            </Badge>
            {accessKey.expires_at && (
              <Badge variant="outline" className="gap-1">
                <Clock className="h-3 w-3" />
                Expires {format(new Date(accessKey.expires_at), 'MMM d, yyyy')}
              </Badge>
            )}
          </div>
          {accessKey.description && (
            <p className="text-sm text-muted-foreground">{accessKey.description}</p>
          )}
          <div className="flex items-center gap-4 text-xs text-muted-foreground pt-2">
            <span className="flex items-center gap-1">
              <Calendar className="h-3 w-3" />
              Created {format(new Date(accessKey.created_at), 'MMM d, yyyy')}
            </span>
            {accessKey.last_used && (
              <span className="flex items-center gap-1">
                <Activity className="h-3 w-3" />
                Last used {format(new Date(accessKey.last_used), 'MMM d, yyyy')}
              </span>
            )}
          </div>
        </div>
        <div className="flex gap-2">
          {children}
        </div>
      </div>

      <div>
        <Label className="text-xs text-muted-foreground">Permissions</Label>
        <div className="flex flex-wrap gap-1 mt-2">
          {accessKey.permissions?.slice(0, 5).map((perm: string) => (
            <Badge key={perm} variant="outline" className="text-xs font-mono">
              <Shield className="h-3 w-3 mr-1" />
              {perm}
            </Badge>
          ))}
          {accessKey.permissions && accessKey.permissions.length > 5 && (
            <Badge variant="outline" className="text-xs">
              +{accessKey.permissions.length - 5} more
            </Badge>
          )}
        </div>
      </div>
    </div>
  );
}