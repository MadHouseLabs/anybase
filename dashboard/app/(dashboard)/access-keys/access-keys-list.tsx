"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Search, Filter, Key } from "lucide-react";
import { AccessKeyCardDisplay } from "./access-key-card-server";
import { ToggleAccessKeyButton, RegenerateKeyButton, DeleteAccessKeyButton } from "./access-keys-client-components";
import { AccessKeysEmptyState } from "./access-keys-list";

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
              <div key={key.id}>
                <AccessKeyCardDisplay accessKey={key}>
                  <ToggleAccessKeyButton accessKey={key} />
                  <RegenerateKeyButton accessKey={key} />
                  <DeleteAccessKeyButton accessKey={key} />
                </AccessKeyCardDisplay>
              </div>
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