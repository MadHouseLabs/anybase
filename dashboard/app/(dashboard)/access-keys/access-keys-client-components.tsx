"use client";

import { useState, useTransition } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from "@/components/ui/dialog";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { useToast } from "@/hooks/use-toast";
import { createAccessKey, updateAccessKey, deleteAccessKey } from "@/app/actions/access-keys-actions";
import { Plus, X, RefreshCw, Trash2, Lock, Unlock, Key, Copy, CheckCircle } from "lucide-react";
import { accessKeysApi } from "@/lib/accesskeys";

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

export function CreateAccessKeyButton() {
  const { toast } = useToast();
  const [open, setOpen] = useState(false);
  const [isPending, startTransition] = useTransition();
  const [customPermissionInput, setCustomPermissionInput] = useState("");
  const [formData, setFormData] = useState({
    name: "",
    description: "",
    permissions: [] as string[],
    expires_in: 0
  });

  const handleSubmit = () => {
    if (!formData.name || formData.permissions.length === 0) {
      toast({
        title: "Error",
        description: "Key name and at least one permission are required",
        variant: "destructive",
      });
      return;
    }

    startTransition(async () => {
      try {
        // Use client API to get the generated key
        const response = await accessKeysApi.create(formData);
        if (response.key) {
          toast({
            title: "Success",
            description: (
              <div className="space-y-2">
                <p>Access key created successfully!</p>
                <div className="p-2 bg-muted rounded">
                  <code className="text-xs break-all">{response.key}</code>
                </div>
                <p className="text-xs">Save this key securely. You won't see it again.</p>
              </div>
            ),
          });
        }
        setOpen(false);
        setFormData({ name: "", description: "", permissions: [], expires_in: 0 });
        setCustomPermissionInput("");
      } catch (error) {
        toast({
          title: "Error",
          description: "Failed to create access key",
          variant: "destructive",
        });
      }
    });
  };

  const addCustomPermission = () => {
    const trimmed = customPermissionInput.trim();
    if (trimmed && !formData.permissions.includes(trimmed)) {
      setFormData(prev => ({
        ...prev,
        permissions: [...prev.permissions, trimmed]
      }));
      setCustomPermissionInput("");
    }
  };

  const removePermission = (permission: string) => {
    setFormData(prev => ({
      ...prev,
      permissions: prev.permissions.filter(p => p !== permission)
    }));
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
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
              value={formData.name}
              onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Describe the purpose of this access key"
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="expires">Expiration (hours)</Label>
            <Input
              id="expires"
              type="number"
              placeholder="0 for no expiration"
              value={formData.expires_in}
              onChange={(e) => setFormData(prev => ({ ...prev, expires_in: parseInt(e.target.value) || 0 }))}
            />
            <p className="text-xs text-muted-foreground">
              Set to 0 for keys that never expire
            </p>
          </div>
          <div className="space-y-2">
            <Label>Permissions</Label>
            
            {formData.permissions.length > 0 && (
              <div className="mb-3">
                <Label className="text-xs text-muted-foreground">Selected Permissions</Label>
                <div className="flex flex-wrap gap-2 mt-2 p-3 bg-muted/30 rounded-md min-h-[60px]">
                  {formData.permissions.map(perm => (
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
          <Button variant="outline" onClick={() => setOpen(false)} disabled={isPending}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isPending}>
            {isPending ? "Creating..." : "Create Access Key"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function ToggleAccessKeyButton({ accessKey }: { accessKey: AccessKey }) {
  const { toast } = useToast();
  const [isPending, startTransition] = useTransition();

  const handleToggle = () => {
    startTransition(async () => {
      const result = await updateAccessKey(accessKey.id, { 
        permissions: accessKey.permissions 
      });
      
      if (result.success) {
        toast({
          title: "Success",
          description: `Access key ${accessKey.name} ${!accessKey.active ? 'enabled' : 'disabled'} successfully`,
        });
      } else {
        toast({
          title: "Error",
          description: result.error || "Failed to update access key",
          variant: "destructive",
        });
      }
    });
  };

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={handleToggle}
      disabled={isPending}
      className={accessKey.active ? "hover:bg-orange-50" : "hover:bg-green-50"}
    >
      {accessKey.active ? (
        <><Lock className="h-4 w-4 mr-1" /> Disable</>
      ) : (
        <><Unlock className="h-4 w-4 mr-1" /> Enable</>
      )}
    </Button>
  );
}

export function RegenerateKeyButton({ accessKey }: { accessKey: AccessKey }) {
  const { toast } = useToast();
  const [regeneratedKey, setRegeneratedKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleRegenerate = async () => {
    if (!confirm(`Are you sure you want to regenerate the key for ${accessKey.name}? The old key will stop working immediately.`)) return;

    try {
      const response = await accessKeysApi.regenerate(accessKey.id);
      setRegeneratedKey(response.key);
      toast({
        title: "Success",
        description: `Access key ${accessKey.name} regenerated successfully`,
      });
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to regenerate key",
        variant: "destructive",
      });
    }
  };

  const copyToClipboard = () => {
    if (regeneratedKey) {
      navigator.clipboard.writeText(regeneratedKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
      toast({
        title: "Copied",
        description: "Access key copied to clipboard",
      });
    }
  };

  return (
    <>
      <Button
        variant="ghost"
        size="sm"
        onClick={handleRegenerate}
        className="hover:bg-blue-50"
      >
        <RefreshCw className="h-4 w-4" />
      </Button>
      
      {regeneratedKey && (
        <Alert className="bg-muted/50 mt-4">
          <Key className="h-4 w-4" />
          <AlertDescription>
            <div className="space-y-2">
              <p className="font-medium">New Access Key Generated:</p>
              <div className="flex items-center gap-2">
                <code className="flex-1 p-2 bg-background rounded text-sm font-mono break-all">
                  {regeneratedKey}
                </code>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={copyToClipboard}
                >
                  {copied ? (
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
    </>
  );
}

export function DeleteAccessKeyButton({ accessKey }: { accessKey: AccessKey }) {
  const { toast } = useToast();
  const [isPending, startTransition] = useTransition();

  const handleDelete = () => {
    if (!confirm(`Are you sure you want to delete the access key ${accessKey.name}? This action cannot be undone.`)) return;

    startTransition(async () => {
      const result = await deleteAccessKey(accessKey.id);
      
      if (result.success) {
        toast({
          title: "Success",
          description: `Access key ${accessKey.name} deleted successfully`,
        });
      } else {
        toast({
          title: "Error",
          description: result.error || "Failed to delete access key",
          variant: "destructive",
        });
      }
    });
  };

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={handleDelete}
      disabled={isPending}
      className="hover:bg-red-50"
    >
      <Trash2 className="h-4 w-4 text-destructive" />
    </Button>
  );
}