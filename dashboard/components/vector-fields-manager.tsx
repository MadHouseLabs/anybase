"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { 
  Plus, Trash2, Layers, Hash, Cpu, Search, Info,
  AlertCircle, CheckCircle, XCircle, Bot, Loader2
} from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { collectionsApi } from "@/lib/api";
import { Switch } from "@/components/ui/switch";
import { Checkbox } from "@/components/ui/checkbox";

interface VectorField {
  name: string;
  dimensions: number;
  model?: string;
  metric?: string;
  index_type?: string;
  list_size?: number;
  m?: number;
  ef_construct?: number;
  enable_auto_rag?: boolean;
  provider_id?: string;
  source_fields?: string[];
  auto_embed?: boolean;
}

interface AIProvider {
  id: string;
  name: string;
  type: string;
}

interface AIModel {
  id: string;
  name: string;
  description?: string;
  type: string;
  dimensions?: number;
}

interface VectorFieldsManagerProps {
  collectionName: string;
  onFieldsUpdate?: () => void;
}

export function VectorFieldsManager({ collectionName, onFieldsUpdate }: VectorFieldsManagerProps) {
  const { toast } = useToast();
  const [vectorFields, setVectorFields] = useState<VectorField[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState<string | null>(null);
  const [providers, setProviders] = useState<AIProvider[]>([]);
  const [availableModels, setAvailableModels] = useState<AIModel[]>([]);
  const [isLoadingModels, setIsLoadingModels] = useState(false);
  const [collectionFields, setCollectionFields] = useState<string[]>([]);
  
  // New field form state
  const [newField, setNewField] = useState<VectorField>({
    name: "",
    dimensions: 1536, // Default for OpenAI text-embedding-ada-002
    metric: "cosine",
    index_type: "ivfflat",
    list_size: 100,
    enable_auto_rag: false,
    provider_id: "",
    source_fields: [],
    auto_embed: true,
    model: "text-embedding-ada-002", // Default embedding model
  });

  // Fetch vector fields
  const fetchVectorFields = async () => {
    try {
      setIsLoading(true);
      const data = await collectionsApi.getVectorFields(collectionName);
      setVectorFields(data.fields || []);
    } catch (error) {
      console.error('Error fetching vector fields:', error);
      toast({
        title: "Error",
        description: "Failed to load vector fields",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Fetch AI providers
  const fetchProviders = async () => {
    try {
      const { aiProvidersApi } = await import('@/lib/api');
      const data = await aiProvidersApi.list();
      setProviders(data.providers || []);
    } catch (error) {
      console.error('Error fetching providers:', error);
    }
  };

  // Fetch collection schema to get field names
  const fetchCollectionSchema = async () => {
    try {
      const collection = await collectionsApi.get(collectionName);
      if (collection?.schema?.properties) {
        setCollectionFields(Object.keys(collection.schema.properties));
      }
    } catch (error) {
      console.error('Error fetching collection schema:', error);
    }
  };

  // Fetch available models for a provider
  const fetchProviderModels = async (providerId: string) => {
    if (!providerId) {
      setAvailableModels([]);
      return;
    }

    try {
      setIsLoadingModels(true);
      const { aiProvidersApi } = await import('@/lib/api');
      const data = await aiProvidersApi.getModels(providerId);
      setAvailableModels(data.models || []);
    } catch (error) {
      console.error('Error fetching provider models:', error);
      setAvailableModels([]);
      toast({
        title: "Warning",
        description: "Unable to fetch available models from provider",
        variant: "default",
      });
    } finally {
      setIsLoadingModels(false);
    }
  };

  useEffect(() => {
    fetchVectorFields();
    fetchProviders();
    fetchCollectionSchema();
  }, [collectionName]);

  // Add vector field
  const handleAddField = async () => {
    if (!newField.name || newField.dimensions <= 0) {
      toast({
        title: "Validation Error",
        description: "Please provide a valid field name and dimensions",
        variant: "destructive",
      });
      return;
    }

    try {
      await collectionsApi.addVectorField(collectionName, newField);

      toast({
        title: "Success",
        description: `Vector field "${newField.name}" added successfully`,
      });

      setIsAddDialogOpen(false);
      setNewField({
        name: "",
        dimensions: 1536,
        metric: "cosine",
        index_type: "ivfflat",
        list_size: 100,
        enable_auto_rag: false,
        provider_id: "",
        source_fields: [],
        auto_embed: true,
        model: "",
      });
      setAvailableModels([]); // Clear models when dialog closes
      
      await fetchVectorFields();
      onFieldsUpdate?.();
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || error.message || "Failed to add vector field",
        variant: "destructive",
      });
    }
  };

  // Remove vector field
  const handleRemoveField = async (fieldName: string) => {
    try {
      setIsDeleting(fieldName);
      
      await collectionsApi.removeVectorField(collectionName, fieldName);

      toast({
        title: "Success",
        description: `Vector field "${fieldName}" removed successfully`,
      });

      await fetchVectorFields();
      onFieldsUpdate?.();
    } catch (error: any) {
      toast({
        title: "Error",
        description: error.response?.data?.error || error.message || "Failed to remove vector field",
        variant: "destructive",
      });
    } finally {
      setIsDeleting(null);
    }
  };

  const getMetricIcon = (metric: string) => {
    switch (metric) {
      case "cosine":
        return <Search className="h-4 w-4" />;
      case "l2":
        return <Hash className="h-4 w-4" />;
      case "inner_product":
      case "ip":
        return <Cpu className="h-4 w-4" />;
      default:
        return <Layers className="h-4 w-4" />;
    }
  };

  const getIndexTypeBadge = (indexType: string) => {
    switch (indexType) {
      case "ivfflat":
        return <Badge variant="secondary">IVFFlat</Badge>;
      case "hnsw":
        return <Badge variant="default">HNSW</Badge>;
      default:
        return <Badge variant="outline">{indexType}</Badge>;
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Vector Fields</CardTitle>
          <CardDescription>Loading vector fields...</CardDescription>
        </CardHeader>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Vector Fields</CardTitle>
              <CardDescription>
                Manage vector embeddings for similarity search and AI features
              </CardDescription>
            </div>
            <Button onClick={() => setIsAddDialogOpen(true)} className="gap-2">
              <Plus className="h-4 w-4" />
              Add Vector Field
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {vectorFields.length === 0 ? (
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                <p className="font-medium">No vector fields configured</p>
                <p className="text-sm text-muted-foreground mt-1">
                  Add vector fields to enable semantic search and AI-powered features
                </p>
              </AlertDescription>
            </Alert>
          ) : (
            <div className="space-y-4">
              {vectorFields.map((field) => (
                <div
                  key={field.name}
                  className="flex items-center justify-between p-4 border rounded-lg bg-muted/30"
                >
                  <div className="flex items-start gap-4">
                    <div className="p-2 bg-primary/10 rounded-lg">
                      <Layers className="h-5 w-5 text-primary" />
                    </div>
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <p className="font-medium">{field.name}</p>
                        {getIndexTypeBadge(field.index_type || "ivfflat")}
                      </div>
                      <div className="flex items-center gap-4 text-sm text-muted-foreground">
                        <span className="flex items-center gap-1">
                          <Hash className="h-3 w-3" />
                          {field.dimensions} dimensions
                        </span>
                        <span className="flex items-center gap-1">
                          {getMetricIcon(field.metric || "cosine")}
                          {field.metric || "cosine"} distance
                        </span>
                        {field.model && (
                          <span className="flex items-center gap-1">
                            <Cpu className="h-3 w-3" />
                            {field.model}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleRemoveField(field.name)}
                    disabled={isDeleting === field.name}
                    title="Remove vector field"
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </div>
              ))}
            </div>
          )}

          <div className="mt-6 p-4 bg-muted/50 rounded-lg">
            <div className="flex items-start gap-2">
              <Info className="h-4 w-4 text-muted-foreground mt-0.5" />
              <div className="text-sm text-muted-foreground">
                <p className="font-medium mb-1">About Vector Fields</p>
                <ul className="space-y-1">
                  <li>• Vector fields store high-dimensional embeddings for similarity search</li>
                  <li>• Use cosine distance for text embeddings (recommended)</li>
                  <li>• IVFFlat index is suitable for most use cases</li>
                  <li>• HNSW provides better recall but uses more memory</li>
                </ul>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Add Vector Field Dialog */}
      <Dialog open={isAddDialogOpen} onOpenChange={setIsAddDialogOpen}>
        <DialogContent className="sm:max-w-[650px] max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="text-xl font-semibold">Configure Vector Field</DialogTitle>
            <DialogDescription className="text-sm text-muted-foreground mt-1.5">
              Set up a vector field to enable semantic search and AI-powered features for your collection.
            </DialogDescription>
          </DialogHeader>
          
          <div className="space-y-5 py-4">
            {/* Field Name Section */}
            <div className="space-y-2">
              <Label htmlFor="field-name" className="text-sm font-medium">
                Field Name <span className="text-red-500">*</span>
              </Label>
              <Input
                id="field-name"
                value={newField.name}
                onChange={(e) => setNewField({ ...newField, name: e.target.value })}
                placeholder="e.g., embeddings, content_vector"
                className="w-full"
              />
              <p className="text-xs text-muted-foreground">
                Choose a descriptive name for your vector field
              </p>
            </div>
            
            {/* Dimensions Section */}
            <div className="space-y-2">
              <Label htmlFor="dimensions" className="text-sm font-medium">
                Vector Dimensions <span className="text-red-500">*</span>
              </Label>
              <Input
                id="dimensions"
                type="number"
                value={newField.dimensions}
                onChange={(e) => setNewField({ ...newField, dimensions: parseInt(e.target.value) || 0 })}
                className="w-full"
                placeholder="Enter number of dimensions"
              />
              <div className="flex items-start gap-1">
                <Info className="h-3 w-3 text-muted-foreground mt-0.5" />
                <p className="text-xs text-muted-foreground">
                  Common dimensions: OpenAI (1536), Cohere (768), BERT (768), Custom models vary
                </p>
              </div>
            </div>

            {/* Model Section - Only shown when NOT using Auto-RAG */}
            {!newField.enable_auto_rag && (
              <div className="space-y-2">
                <Label htmlFor="model" className="text-sm font-medium">
                  Embedding Model <span className="text-xs text-muted-foreground">(Optional)</span>
                </Label>
                <Input
                  id="model"
                  value={newField.model || ""}
                  onChange={(e) => setNewField({ ...newField, model: e.target.value })}
                  className="w-full"
                  placeholder="e.g., text-embedding-ada-002, all-MiniLM-L6-v2"
                />
                <p className="text-xs text-muted-foreground">
                  Document which model was used for generating embeddings
                </p>
              </div>
            )}

            {/* Auto-RAG Configuration Section */}
            <div className="space-y-4 border-t pt-4">
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label htmlFor="enable-rag" className="text-sm font-medium flex items-center gap-2">
                    <Bot className="h-4 w-4" />
                    Enable Auto-RAG
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Automatically generate embeddings using AI providers
                  </p>
                </div>
                <Switch
                  id="enable-rag"
                  checked={newField.enable_auto_rag}
                  onCheckedChange={(checked) => setNewField({ ...newField, enable_auto_rag: checked })}
                />
              </div>

              {newField.enable_auto_rag && (
                <div className="space-y-4 pl-4 border-l-2 border-muted ml-2">
                  {/* AI Provider Selection */}
                  <div className="space-y-2">
                    <Label htmlFor="provider" className="text-sm font-medium">
                      AI Provider <span className="text-red-500">*</span>
                    </Label>
                    <Select
                      value={newField.provider_id}
                      onValueChange={async (value) => {
                        setNewField({ 
                          ...newField, 
                          provider_id: value,
                          model: "" // Reset model selection when provider changes
                        });
                        
                        // Fetch available models for the selected provider
                        await fetchProviderModels(value);
                      }}
                    >
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="Select an AI provider" />
                      </SelectTrigger>
                      <SelectContent>
                        {providers.length === 0 ? (
                          <div className="px-2 py-1.5 text-sm text-muted-foreground">
                            No providers configured
                          </div>
                        ) : (
                          providers.map((provider) => (
                            <SelectItem key={provider.id} value={provider.id}>
                              {provider.name} ({provider.type})
                            </SelectItem>
                          ))
                        )}
                      </SelectContent>
                    </Select>
                    {providers.length === 0 && (
                      <p className="text-xs text-yellow-600">
                        No AI providers configured. Configure providers in Integrations → AI Providers.
                      </p>
                    )}
                  </div>

                  {/* Embedding Model Selection */}
                  <div className="space-y-2">
                    <Label htmlFor="embedding-model" className="text-sm font-medium">
                      Embedding Model <span className="text-red-500">*</span>
                    </Label>
                    {newField.provider_id ? (
                      <Select
                        value={newField.model || ""}
                        onValueChange={(value) => {
                          const selectedModel = availableModels.find(m => m.id === value);
                          setNewField({ 
                            ...newField, 
                            model: value,
                            dimensions: selectedModel?.dimensions || newField.dimensions
                          });
                        }}
                        disabled={isLoadingModels}
                      >
                        <SelectTrigger className="w-full">
                          <SelectValue placeholder={
                            isLoadingModels 
                              ? "Loading models..." 
                              : availableModels.length > 0 
                                ? "Select an embedding model" 
                                : "No models available"
                          }>
                            {newField.model && (() => {
                              const selectedModel = availableModels.find(m => m.id === newField.model);
                              if (!selectedModel) return newField.model;
                              return (
                                <div className="flex items-center justify-between w-full gap-4">
                                  <div className="flex-1 min-w-0">
                                    <span className="font-medium text-sm truncate">{selectedModel.name}</span>
                                    {selectedModel.dimensions && (
                                      <span className="ml-2 text-xs text-muted-foreground font-mono">
                                        ({selectedModel.dimensions}d)
                                      </span>
                                    )}
                                  </div>
                                </div>
                              );
                            })()}
                          </SelectValue>
                          {isLoadingModels && (
                            <Loader2 className="h-4 w-4 animate-spin ml-2" />
                          )}
                        </SelectTrigger>
                        <SelectContent>
                          {availableModels.length === 0 ? (
                            <div className="px-2 py-1.5 text-sm text-muted-foreground">
                              {isLoadingModels ? "Loading models..." : "No models available"}
                            </div>
                          ) : (
                            availableModels.map((model) => (
                              <SelectItem key={model.id} value={model.id}>
                                <div className="space-y-1">
                                  <div className="font-medium text-sm">
                                    {model.name}
                                    {model.dimensions && (
                                      <span className="ml-2 text-xs text-muted-foreground font-mono">
                                        ({model.dimensions}d)
                                      </span>
                                    )}
                                  </div>
                                  {model.description && (
                                    <div className="text-xs text-muted-foreground">
                                      {model.description}
                                    </div>
                                  )}
                                </div>
                              </SelectItem>
                            ))
                          )}
                        </SelectContent>
                      </Select>
                    ) : (
                      <Input
                        id="embedding-model"
                        value={newField.model || ""}
                        onChange={(e) => {
                          const modelValue = e.target.value;
                          // Update dimensions based on known models
                          let dimensions = newField.dimensions;
                          if (modelValue.includes('ada-002')) dimensions = 1536;
                          else if (modelValue.includes('embed-english')) dimensions = 1024;
                          else if (modelValue.includes('MiniLM')) dimensions = 384;
                          else if (modelValue.includes('nomic-embed')) dimensions = 768;
                          
                          setNewField({ ...newField, model: modelValue, dimensions });
                        }}
                        placeholder="e.g., text-embedding-ada-002"
                        className="w-full"
                        disabled
                      />
                    )}
                    <p className="text-xs text-muted-foreground">
                      {newField.provider_id 
                        ? "Select from available models provided by your chosen AI provider" 
                        : "Please select an AI provider first to see available models"}
                    </p>
                  </div>

                  {/* Source Fields Selection */}
                  <div className="space-y-2">
                    <Label className="text-sm font-medium">
                      Source Fields for Embedding <span className="text-red-500">*</span>
                    </Label>
                    <p className="text-xs text-muted-foreground mb-2">
                      Select which document fields to use for generating embeddings
                    </p>
                    <div className="space-y-2 max-h-32 overflow-y-auto border rounded-md p-2">
                      {collectionFields.length === 0 ? (
                        <p className="text-xs text-muted-foreground">No fields available</p>
                      ) : (
                        collectionFields.map((field) => (
                          <div key={field} className="flex items-center space-x-2">
                            <Checkbox
                              id={`field-${field}`}
                              checked={newField.source_fields?.includes(field)}
                              onCheckedChange={(checked) => {
                                const currentFields = newField.source_fields || [];
                                if (checked) {
                                  setNewField({ 
                                    ...newField, 
                                    source_fields: [...currentFields, field]
                                  });
                                } else {
                                  setNewField({ 
                                    ...newField, 
                                    source_fields: currentFields.filter(f => f !== field)
                                  });
                                }
                              }}
                            />
                            <Label 
                              htmlFor={`field-${field}`}
                              className="text-sm font-normal cursor-pointer"
                            >
                              {field}
                            </Label>
                          </div>
                        ))
                      )}
                    </div>
                  </div>

                  {/* Auto-Embed Option */}
                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label htmlFor="auto-embed" className="text-sm font-medium">
                        Auto-Generate on Document Changes
                      </Label>
                      <p className="text-xs text-muted-foreground">
                        Automatically update embeddings when documents are created or modified
                      </p>
                    </div>
                    <Switch
                      id="auto-embed"
                      checked={newField.auto_embed}
                      onCheckedChange={(checked) => setNewField({ ...newField, auto_embed: checked })}
                    />
                  </div>
                </div>
              )}
            </div>

            {/* Distance Metric Section */}
            <div className="space-y-2">
              <Label htmlFor="metric" className="text-sm font-medium">
                Distance Metric <span className="text-red-500">*</span>
              </Label>
              <Select
                value={newField.metric}
                onValueChange={(value) => setNewField({ ...newField, metric: value })}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="cosine">Cosine Similarity</SelectItem>
                  <SelectItem value="l2">Euclidean Distance (L2)</SelectItem>
                  <SelectItem value="inner_product">Inner Product</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                Cosine: Best for text embeddings | L2: Good for dense vectors | Inner Product: Fast computation
              </p>
            </div>

            {/* Index Type Section */}
            <div className="space-y-2">
              <Label htmlFor="index-type" className="text-sm font-medium">
                Index Type <span className="text-red-500">*</span>
              </Label>
              <Select
                value={newField.index_type}
                onValueChange={(value) => setNewField({ ...newField, index_type: value })}
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="ivfflat">IVFFlat</SelectItem>
                  <SelectItem value="hnsw">HNSW</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                IVFFlat: Balanced speed & accuracy | HNSW: Higher recall, faster queries, more memory
              </p>
            </div>

            {/* Advanced Index Parameters */}
            {newField.index_type === "ivfflat" && (
              <div className="space-y-3 p-4 bg-muted/20 rounded-lg border border-muted">
                <h4 className="text-sm font-semibold">IVFFlat Parameters</h4>
                <div className="space-y-2">
                  <Label htmlFor="list-size" className="text-sm font-medium">
                    List Size
                  </Label>
                  <Input
                    id="list-size"
                    type="number"
                    value={newField.list_size || 100}
                    onChange={(e) => setNewField({ ...newField, list_size: parseInt(e.target.value) || 100 })}
                    className="w-full"
                    placeholder="100"
                  />
                  <p className="text-xs text-muted-foreground">
                    Number of clusters. Higher values = better recall but slower build time
                  </p>
                </div>
              </div>
            )}

            {newField.index_type === "hnsw" && (
              <div className="space-y-4 p-4 bg-muted/20 rounded-lg border border-muted">
                <h4 className="text-sm font-semibold">HNSW Parameters</h4>
                <div className="space-y-2">
                  <Label htmlFor="m" className="text-sm font-medium">
                    M (Max Connections)
                  </Label>
                  <Input
                    id="m"
                    type="number"
                    value={newField.m || 16}
                    onChange={(e) => setNewField({ ...newField, m: parseInt(e.target.value) || 16 })}
                    className="w-full"
                    placeholder="16"
                  />
                  <p className="text-xs text-muted-foreground">
                    Number of bi-directional links per node (16-64 recommended)
                  </p>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="ef-construct" className="text-sm font-medium">
                    EF Construction
                  </Label>
                  <Input
                    id="ef-construct"
                    type="number"
                    value={newField.ef_construct || 200}
                    onChange={(e) => setNewField({ ...newField, ef_construct: parseInt(e.target.value) || 200 })}
                    className="w-full"
                    placeholder="200"
                  />
                  <p className="text-xs text-muted-foreground">
                    Size of dynamic candidate list (higher = better quality, slower build)
                  </p>
                </div>
              </div>
            )}
          </div>

          <DialogFooter className="flex-row justify-end space-x-2 pt-4 border-t">
            <Button 
              variant="outline" 
              onClick={() => {
                setIsAddDialogOpen(false);
                setNewField({
                  name: "",
                  dimensions: 1536,
                  metric: "cosine",
                  index_type: "ivfflat",
                  list_size: 100,
                  enable_auto_rag: false,
                  provider_id: "",
                  source_fields: [],
                  auto_embed: true,
                  model: "",
                });
                setAvailableModels([]); // Clear models when dialog closes
              }}
            >
              Cancel
            </Button>
            <Button 
              onClick={handleAddField}
              disabled={!newField.name || newField.dimensions <= 0}
            >
              <Plus className="h-4 w-4 mr-2" />
              Create Vector Field
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}