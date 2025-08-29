"use client";

import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Search, Filter, Key, CheckCircle, Copy } from "lucide-react";
import { AccessKeyCardDisplay } from "./access-key-card-server";
import { ToggleAccessKeyButton, RegenerateKeyButton, DeleteAccessKeyButton, ViewKeyButton } from "./access-keys-client-components";
import { AccessKeysEmptyState } from "./access-keys-list";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";

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

interface AccessKeysListWithSearchProps {
  accessKeys: AccessKey[];
}

export function AccessKeysListWithSearch({ accessKeys }: AccessKeysListWithSearchProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [storedKeyIds, setStoredKeyIds] = useState<Set<string>>(new Set());

  useEffect(() => {
    // Check which keys are stored locally
    const checkStoredKeys = () => {
      const storedKeys = localStorage.getItem("anybase_access_keys");
      if (storedKeys) {
        const keys = JSON.parse(storedKeys);
        setStoredKeyIds(new Set(Object.keys(keys)));
      } else {
        setStoredKeyIds(new Set());
      }
    };

    checkStoredKeys();

    // Listen for storage changes (including from ViewKeyButton)
    const handleStorageChange = () => {
      checkStoredKeys();
    };

    window.addEventListener("storage", handleStorageChange);
    // Custom event for same-window updates
    window.addEventListener("anybase-keys-updated", handleStorageChange);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
      window.removeEventListener("anybase-keys-updated", handleStorageChange);
    };
  }, []);

  const filteredKeys = accessKeys.filter(key => {
    const matchesSearch = searchQuery === "" || 
      key.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      key.description?.toLowerCase().includes(searchQuery.toLowerCase());
    
    const matchesStatus = statusFilter === "all" || 
      (statusFilter === "active" && key.active) ||
      (statusFilter === "inactive" && !key.active) ||
      (statusFilter === "expired" && key.expires_at && new Date(key.expires_at) < new Date());
    
    return matchesSearch && matchesStatus;
  });

  return (
    <>
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
      <div className="space-y-4 mt-4">
        {filteredKeys.length === 0 ? (
          <AccessKeysEmptyState hasSearch={accessKeys.length > 0 && filteredKeys.length === 0} />
        ) : (
          <>
            {filteredKeys.map((key) => (
              <AccessKeyCard key={key.id} accessKey={key} hasStoredKey={storedKeyIds.has(key.id)} />
            ))}
            {filteredKeys.length > 0 && (
              <div className="px-6 py-4 border-t text-sm text-muted-foreground">
                Showing {filteredKeys.length} of {accessKeys.length} keys
              </div>
            )}
          </>
        )}
      </div>
    </>
  );
}

function AccessKeyCard({ accessKey, hasStoredKey }: { accessKey: AccessKey; hasStoredKey: boolean }) {
  const [regeneratedKey, setRegeneratedKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const { toast } = useToast();

  const handleCopy = () => {
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
    <div className="space-y-2">
      <div className="relative">
        {hasStoredKey && (
          <Badge 
            className="absolute -top-2 -right-2 z-10 bg-green-100 text-green-800 border-green-200"
            variant="outline"
          >
            <CheckCircle className="h-3 w-3 mr-1" />
            Key Saved
          </Badge>
        )}
        <AccessKeyCardDisplay accessKey={accessKey}>
          <ViewKeyButton accessKey={accessKey} />
          <ToggleAccessKeyButton accessKey={accessKey} />
          <RegenerateKeyButton accessKey={accessKey} onRegenerate={setRegeneratedKey} />
          <DeleteAccessKeyButton accessKey={accessKey} />
        </AccessKeyCardDisplay>
      </div>
      
      {regeneratedKey && (
        <Alert className="bg-green-50 border-green-200">
          <Key className="h-4 w-4 text-green-600" />
          <AlertDescription>
            <div className="space-y-3">
              <p className="font-medium text-green-800">New Access Key Generated:</p>
              <div className="flex items-start gap-2">
                <code className="flex-1 p-3 bg-white border border-green-200 rounded text-xs font-mono break-all select-all">
                  {regeneratedKey}
                </code>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleCopy}
                  className="shrink-0"
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
    </div>
  );
}

export function AccessKeysEmptyState({ hasSearch }: { hasSearch: boolean }) {
  return (
    <div className="flex flex-col items-center justify-center py-12">
      <Key className="h-12 w-12 text-muted-foreground mb-4" />
      <p className="text-lg font-medium mb-2">
        {hasSearch ? "No keys found matching your filters" : "No access keys yet"}
      </p>
      <p className="text-sm text-muted-foreground">
        {hasSearch ? "Try adjusting your search or filters" : "Create your first access key to get started"}
      </p>
    </div>
  );
}