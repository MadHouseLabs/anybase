'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { 
  Plus, 
  Trash2, 
  Brain, 
  Key, 
  Loader2,
  ChevronRight,
  ArrowLeft,
  CheckCircle,
  AlertCircle,
} from 'lucide-react';
import { useToast } from '@/hooks/use-toast';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { aiProvidersApi } from '@/lib/api';
import Link from 'next/link';

interface AIProvider {
  id: string;
  name: string;
  type: string;
  base_url?: string;
  rate_limits?: {
    requests_per_minute: number;
    tokens_per_minute: number;
  };
  active: boolean;
  created_at: string;
  updated_at: string;
}

const PROVIDER_TYPES = [
  { value: 'openai', label: 'OpenAI', requiresKey: true, defaultUrl: '' },
  { value: 'anthropic', label: 'Anthropic', requiresKey: true, defaultUrl: '' },
  { value: 'cohere', label: 'Cohere', requiresKey: true, defaultUrl: '' },
  { value: 'huggingface', label: 'HuggingFace', requiresKey: true, defaultUrl: '' },
  { value: 'ollama', label: 'Ollama (Local)', requiresKey: false, defaultUrl: 'http://localhost:11434' },
  { value: 'custom', label: 'Custom', requiresKey: false, defaultUrl: '' },
];

export default function AIProvidersPage() {
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAddDialog, setShowAddDialog] = useState(false);
  const { toast } = useToast();

  // Form state for new provider
  const [newProvider, setNewProvider] = useState({
    name: '',
    type: 'openai',
    api_key: '',
    base_url: '',
    requests_per_minute: 100,
  });

  useEffect(() => {
    fetchProviders();
  }, []);

  const fetchProviders = async () => {
    try {
      const data = await aiProvidersApi.list();
      setProviders(data.providers || []);
    } catch (error) {
      console.error('Failed to fetch providers:', error);
      // Don't show error toast on initial load
    } finally {
      setLoading(false);
    }
  };

  const handleAddProvider = async () => {
    try {
      if (!newProvider.name) {
        toast({
          title: 'Error',
          description: 'Provider name is required',
          variant: 'destructive',
        });
        return;
      }

      const providerConfig = PROVIDER_TYPES.find(p => p.value === newProvider.type);
      if (providerConfig?.requiresKey && !newProvider.api_key) {
        toast({
          title: 'Error',
          description: 'API key is required for this provider type',
          variant: 'destructive',
        });
        return;
      }

      const payload = {
        name: newProvider.name,
        type: newProvider.type,
        api_key: newProvider.api_key,
        base_url: newProvider.base_url || undefined,
        rate_limits: {
          requests_per_minute: newProvider.requests_per_minute,
        },
        active: true,
      };

      await aiProvidersApi.create(payload);

      toast({
        title: 'Success',
        description: 'AI provider added successfully',
      });

      setShowAddDialog(false);
      setNewProvider({
        name: '',
        type: 'openai',
        api_key: '',
        base_url: '',
        requests_per_minute: 100,
      });
      fetchProviders();
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to add AI provider',
        variant: 'destructive',
      });
    }
  };

  const handleDeleteProvider = async (id: string) => {
    try {
      await aiProvidersApi.delete(id);

      toast({
        title: 'Success',
        description: 'AI provider deleted successfully',
      });

      fetchProviders();
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to delete AI provider',
        variant: 'destructive',
      });
    }
  };

  const getProviderTypeConfig = (type: string) => {
    return PROVIDER_TYPES.find(p => p.value === type) || PROVIDER_TYPES[0];
  };

  const getProviderIcon = (type: string) => {
    switch (type) {
      case 'openai': return '🤖';
      case 'anthropic': return '🧠';
      case 'cohere': return '🔮';
      case 'huggingface': return '🤗';
      case 'ollama': return '🦙';
      default: return '⚡';
    }
  };

  const getProviderBadgeColor = (type: string) => {
    switch (type) {
      case 'openai': return 'bg-green-500/10 text-green-600 dark:text-green-400 border-green-500/20';
      case 'anthropic': return 'bg-orange-500/10 text-orange-600 dark:text-orange-400 border-orange-500/20';
      case 'cohere': return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border-purple-500/20';
      case 'huggingface': return 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400 border-yellow-500/20';
      case 'ollama': return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/20';
      default: return 'bg-gray-500/10 text-gray-600 dark:text-gray-400 border-gray-500/20';
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto py-6">
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-6">
      {/* Breadcrumb Navigation */}
      <nav className="flex items-center space-x-2 text-sm text-muted-foreground mb-6">
        <Link href="/dashboard" className="hover:text-foreground transition-colors">
          Dashboard
        </Link>
        <ChevronRight className="h-4 w-4" />
        <Link href="/integrations" className="hover:text-foreground transition-colors">
          Integrations
        </Link>
        <ChevronRight className="h-4 w-4" />
        <span className="text-foreground">AI Providers</span>
      </nav>

      {/* Page Header */}
      <div className="mb-8">
        <div className="flex justify-between items-start">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <Link 
                href="/integrations" 
                className="p-2 hover:bg-muted rounded-lg transition-colors"
                title="Back to Integrations"
              >
                <ArrowLeft className="h-5 w-5" />
              </Link>
              <div className="p-2 bg-primary/10 rounded-lg">
                <Brain className="h-6 w-6 text-primary" />
              </div>
              <h1 className="text-3xl font-bold">AI Providers</h1>
            </div>
            <p className="text-muted-foreground ml-14">
              Configure AI providers for embeddings, RAG, and intelligent search capabilities
            </p>
          </div>
          <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Add Provider
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-lg">
              <DialogHeader>
                <DialogTitle>Add AI Provider</DialogTitle>
                <DialogDescription>
                  Configure credentials for accessing AI services
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="space-y-2">
                  <Label htmlFor="name">Provider Name</Label>
                  <Input
                    id="name"
                    value={newProvider.name}
                    onChange={(e) => setNewProvider({ ...newProvider, name: e.target.value })}
                    placeholder="e.g., Production OpenAI"
                  />
                  <p className="text-xs text-muted-foreground">
                    A friendly name to identify this provider
                  </p>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="type">Provider Type</Label>
                  <Select
                    value={newProvider.type}
                    onValueChange={(value) => {
                      const config = PROVIDER_TYPES.find(p => p.value === value);
                      setNewProvider({
                        ...newProvider,
                        type: value,
                        base_url: config?.defaultUrl || '',
                      });
                    }}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {PROVIDER_TYPES.map((type) => (
                        <SelectItem key={type.value} value={type.value}>
                          {type.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {getProviderTypeConfig(newProvider.type).requiresKey && (
                  <div className="space-y-2">
                    <Label htmlFor="api_key">API Key</Label>
                    <Input
                      id="api_key"
                      type="password"
                      value={newProvider.api_key}
                      onChange={(e) => setNewProvider({ ...newProvider, api_key: e.target.value })}
                      placeholder="sk-..."
                    />
                    <p className="text-xs text-muted-foreground">
                      Your API key will be encrypted and stored securely
                    </p>
                  </div>
                )}

                {(newProvider.type === 'ollama' || newProvider.type === 'custom') && (
                  <div className="space-y-2">
                    <Label htmlFor="base_url">Base URL</Label>
                    <Input
                      id="base_url"
                      value={newProvider.base_url}
                      onChange={(e) => setNewProvider({ ...newProvider, base_url: e.target.value })}
                      placeholder="http://localhost:11434"
                    />
                    <p className="text-xs text-muted-foreground">
                      The endpoint URL for the AI service
                    </p>
                  </div>
                )}

                <div className="space-y-2">
                  <Label htmlFor="rate_limit">Rate Limit (requests/minute)</Label>
                  <Input
                    id="rate_limit"
                    type="number"
                    value={newProvider.requests_per_minute}
                    onChange={(e) => setNewProvider({ ...newProvider, requests_per_minute: parseInt(e.target.value) })}
                    placeholder="100"
                  />
                  <p className="text-xs text-muted-foreground">
                    Limit API requests to prevent rate limiting errors
                  </p>
                </div>
              </div>
              <DialogFooter>
                <Button variant="outline" onClick={() => setShowAddDialog(false)}>
                  Cancel
                </Button>
                <Button onClick={handleAddProvider}>
                  Add Provider
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Info Alert */}
      <Card className="mb-6 border-blue-500/20 bg-blue-500/5">
        <CardHeader className="pb-3">
          <div className="flex items-start gap-3">
            <AlertCircle className="h-5 w-5 text-blue-500 mt-0.5" />
            <div>
              <CardTitle className="text-base">How it works</CardTitle>
              <CardDescription className="mt-1">
                Configure your AI provider credentials here. When creating vector fields on collections, 
                you'll select which provider and model to use for generating embeddings. 
                This allows flexibility to use different models for different use cases.
              </CardDescription>
            </div>
          </div>
        </CardHeader>
      </Card>

      {/* Providers List */}
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Configured Providers</h2>
          {providers.length > 0 && (
            <Badge variant="secondary">
              {providers.length} Provider{providers.length !== 1 ? 's' : ''}
            </Badge>
          )}
        </div>

        {providers.length === 0 ? (
          <Card className="border-dashed">
            <CardContent className="text-center py-12">
              <div className="mx-auto w-fit p-3 bg-muted rounded-lg mb-4">
                <Brain className="h-8 w-8 text-muted-foreground" />
              </div>
              <h3 className="font-semibold mb-2">No providers configured</h3>
              <p className="text-sm text-muted-foreground mb-4 max-w-sm mx-auto">
                Add your first AI provider to enable embeddings and RAG capabilities
              </p>
              <Button onClick={() => setShowAddDialog(true)}>
                <Plus className="h-4 w-4 mr-2" />
                Add Your First Provider
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {providers.map((provider) => (
              <Card key={provider.id} className="group hover:shadow-md transition-shadow">
                <CardHeader>
                  <div className="flex justify-between items-start">
                    <div className="flex items-start gap-3 flex-1">
                      <div className="text-2xl mt-1">
                        {getProviderIcon(provider.type)}
                      </div>
                      <div className="flex-1 min-w-0">
                        <CardTitle className="text-base truncate">{provider.name}</CardTitle>
                        <div className="flex items-center gap-2 mt-1">
                          <Badge 
                            variant="outline" 
                            className={`text-xs ${getProviderBadgeColor(provider.type)}`}
                          >
                            {provider.type}
                          </Badge>
                          {provider.active && (
                            <Badge variant="outline" className="text-xs border-green-500/50">
                              <CheckCircle className="h-3 w-3 mr-1 text-green-500" />
                              Active
                            </Badge>
                          )}
                        </div>
                      </div>
                    </div>
                    <Button
                      size="icon"
                      variant="ghost"
                      className="h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity hover:text-destructive"
                      onClick={() => handleDeleteProvider(provider.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="space-y-2 text-sm">
                    {provider.base_url && (
                      <div className="flex items-center gap-2">
                        <span className="text-muted-foreground">URL:</span>
                        <span className="font-mono text-xs truncate">{provider.base_url}</span>
                      </div>
                    )}
                    {provider.rate_limits && (
                      <div className="flex items-center gap-2">
                        <span className="text-muted-foreground">Rate limit:</span>
                        <span>{provider.rate_limits.requests_per_minute} req/min</span>
                      </div>
                    )}
                    <div className="flex items-center gap-2">
                      <Key className="h-3.5 w-3.5 text-muted-foreground" />
                      <span className="text-muted-foreground">API Key:</span>
                      <Badge variant="outline" className="text-xs">
                        Configured
                      </Badge>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}