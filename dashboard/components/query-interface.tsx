"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { 
  Search, Code, Layers, Database, Play, Copy, Download,
  ChevronDown, ChevronUp, AlertCircle, Sparkles, Filter
} from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { ScrollArea } from "@/components/ui/scroll-area";
import { collectionsApi, dataApi } from "@/lib/api";
import { formatLocalDateTime } from "@/lib/date-utils";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import { Slider } from "@/components/ui/slider";

interface QueryInterfaceProps {
  collectionName: string;
}

interface QueryResult {
  results: any[];
  count: number;
  executionTime?: number;
  error?: string;
}

interface VectorField {
  name: string;
  dimensions: number;
  metric?: string;
}

export function QueryInterface({ collectionName }: QueryInterfaceProps) {
  const { toast } = useToast();
  const [activeTab, setActiveTab] = useState("standard");
  const [isLoading, setIsLoading] = useState(false);
  const [results, setResults] = useState<QueryResult | null>(null);
  const [vectorFields, setVectorFields] = useState<VectorField[]>([]);
  const [showAdvanced, setShowAdvanced] = useState(false);
  
  // Standard query state
  const [filter, setFilter] = useState("{}");
  const [projection, setProjection] = useState("");
  const [sort, setSort] = useState("");
  const [limit, setLimit] = useState("20");
  const [skip, setSkip] = useState("0");
  
  // Vector search state
  const [selectedVectorField, setSelectedVectorField] = useState("");
  const [vectorQuery, setVectorQuery] = useState("");
  const [topK, setTopK] = useState("10");
  const [vectorFilter, setVectorFilter] = useState("{}");
  
  // Hybrid search state
  const [textQuery, setTextQuery] = useState("");
  const [hybridVectorField, setHybridVectorField] = useState("");
  const [hybridVector, setHybridVector] = useState("");
  const [hybridTopK, setHybridTopK] = useState("10");
  const [alpha, setAlpha] = useState([0.5]);
  const [hybridFilter, setHybridFilter] = useState("{}");

  // Fetch vector fields
  useEffect(() => {
    const fetchVectorFields = async () => {
      try {
        const data = await collectionsApi.getVectorFields(collectionName);
        setVectorFields(data.fields || []);
        if (data.fields?.length > 0) {
          setSelectedVectorField(data.fields[0].name);
          setHybridVectorField(data.fields[0].name);
        }
      } catch (error) {
        console.error("Error fetching vector fields:", error);
      }
    };
    fetchVectorFields();
  }, [collectionName]);

  // Execute standard query
  const executeStandardQuery = async () => {
    try {
      setIsLoading(true);
      const startTime = Date.now();
      
      const params: any = {
        limit: parseInt(limit) || 20,
        skip: parseInt(skip) || 0,
      };
      
      if (filter && filter !== "{}") {
        try {
          params.filter = JSON.parse(filter);
        } catch (e) {
          throw new Error("Invalid filter JSON");
        }
      }
      
      if (projection) {
        params.projection = projection.split(',').map(p => p.trim());
      }
      
      if (sort) {
        try {
          params.sort = JSON.parse(sort);
        } catch (e) {
          throw new Error("Invalid sort JSON");
        }
      }
      
      const response = await dataApi.query(collectionName, params);
      const executionTime = Date.now() - startTime;
      
      setResults({
        results: response.documents || response.data || [],
        count: response.count || response.documents?.length || 0,
        executionTime,
      });
    } catch (error: any) {
      setResults({
        results: [],
        count: 0,
        error: error.message || "Query failed",
      });
      toast({
        title: "Query Error",
        description: error.message || "Failed to execute query",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Execute vector search
  const executeVectorSearch = async () => {
    if (!selectedVectorField) {
      toast({
        title: "Error",
        description: "Please select a vector field",
        variant: "destructive",
      });
      return;
    }
    
    if (!vectorQuery) {
      toast({
        title: "Error",
        description: "Please enter a vector query",
        variant: "destructive",
      });
      return;
    }
    
    try {
      setIsLoading(true);
      const startTime = Date.now();
      
      let queryVector: number[];
      try {
        queryVector = JSON.parse(vectorQuery);
        if (!Array.isArray(queryVector)) {
          throw new Error("Vector must be an array");
        }
      } catch (e) {
        throw new Error("Invalid vector format. Please provide a JSON array of numbers");
      }
      
      const searchParams: any = {
        vector_field: selectedVectorField,
        query_vector: queryVector,
        top_k: parseInt(topK) || 10,
      };
      
      if (vectorFilter && vectorFilter !== "{}") {
        try {
          searchParams.filter = JSON.parse(vectorFilter);
        } catch (e) {
          throw new Error("Invalid filter JSON");
        }
      }
      
      const response = await dataApi.vectorSearch(collectionName, searchParams);
      const executionTime = Date.now() - startTime;
      
      setResults({
        results: response.results || [],
        count: response.count || response.results?.length || 0,
        executionTime,
      });
    } catch (error: any) {
      setResults({
        results: [],
        count: 0,
        error: error.message || "Vector search failed",
      });
      toast({
        title: "Vector Search Error",
        description: error.message || "Failed to execute vector search",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Execute hybrid search
  const executeHybridSearch = async () => {
    if (!textQuery) {
      toast({
        title: "Error",
        description: "Please enter a text query",
        variant: "destructive",
      });
      return;
    }
    
    if (!hybridVectorField || !hybridVector) {
      toast({
        title: "Error",
        description: "Please select a vector field and provide a vector",
        variant: "destructive",
      });
      return;
    }
    
    try {
      setIsLoading(true);
      const startTime = Date.now();
      
      let queryVector: number[];
      try {
        queryVector = JSON.parse(hybridVector);
        if (!Array.isArray(queryVector)) {
          throw new Error("Vector must be an array");
        }
      } catch (e) {
        throw new Error("Invalid vector format. Please provide a JSON array of numbers");
      }
      
      const searchParams: any = {
        text_query: textQuery,
        vector_field: hybridVectorField,
        query_vector: queryVector,
        top_k: parseInt(hybridTopK) || 10,
        alpha: alpha[0],
      };
      
      if (hybridFilter && hybridFilter !== "{}") {
        try {
          searchParams.filter = JSON.parse(hybridFilter);
        } catch (e) {
          throw new Error("Invalid filter JSON");
        }
      }
      
      const response = await dataApi.hybridSearch(collectionName, searchParams);
      const executionTime = Date.now() - startTime;
      
      setResults({
        results: response.results || [],
        count: response.count || response.results?.length || 0,
        executionTime,
      });
    } catch (error: any) {
      setResults({
        results: [],
        count: 0,
        error: error.message || "Hybrid search failed",
      });
      toast({
        title: "Hybrid Search Error",
        description: error.message || "Failed to execute hybrid search",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const executeQuery = () => {
    switch (activeTab) {
      case "standard":
        executeStandardQuery();
        break;
      case "vector":
        executeVectorSearch();
        break;
      case "hybrid":
        executeHybridSearch();
        break;
    }
  };

  const copyResults = () => {
    if (results?.results) {
      navigator.clipboard.writeText(JSON.stringify(results.results, null, 2));
      toast({
        title: "Copied",
        description: "Results copied to clipboard",
      });
    }
  };

  const downloadResults = () => {
    if (results?.results) {
      const blob = new Blob([JSON.stringify(results.results, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${collectionName}_query_results_${Date.now()}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Query Interface</CardTitle>
          <CardDescription>
            Execute queries, vector searches, and hybrid searches on your collection
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="standard" className="gap-2">
                <Database className="h-4 w-4" />
                Standard Query
              </TabsTrigger>
              <TabsTrigger value="vector" className="gap-2" disabled={vectorFields.length === 0}>
                <Layers className="h-4 w-4" />
                Vector Search
              </TabsTrigger>
              <TabsTrigger value="hybrid" className="gap-2" disabled={vectorFields.length === 0}>
                <Sparkles className="h-4 w-4" />
                Hybrid Search
              </TabsTrigger>
            </TabsList>

            <TabsContent value="standard" className="space-y-4 mt-6">
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="filter">Filter (MongoDB-style JSON)</Label>
                  <Textarea
                    id="filter"
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                    placeholder='{"status": "active", "price": {"$gte": 100}}'
                    className="font-mono text-sm"
                    rows={3}
                  />
                </div>

                <Collapsible open={showAdvanced} onOpenChange={setShowAdvanced}>
                  <CollapsibleTrigger asChild>
                    <Button variant="outline" className="gap-2">
                      {showAdvanced ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                      Advanced Options
                    </Button>
                  </CollapsibleTrigger>
                  <CollapsibleContent className="space-y-4 mt-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="projection">Projection (comma-separated fields)</Label>
                        <Input
                          id="projection"
                          value={projection}
                          onChange={(e) => setProjection(e.target.value)}
                          placeholder="name, price, status"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="sort">Sort (JSON)</Label>
                        <Input
                          id="sort"
                          value={sort}
                          onChange={(e) => setSort(e.target.value)}
                          placeholder='{"created_at": -1}'
                          className="font-mono"
                        />
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="limit">Limit</Label>
                        <Input
                          id="limit"
                          type="number"
                          value={limit}
                          onChange={(e) => setLimit(e.target.value)}
                          placeholder="20"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="skip">Skip</Label>
                        <Input
                          id="skip"
                          type="number"
                          value={skip}
                          onChange={(e) => setSkip(e.target.value)}
                          placeholder="0"
                        />
                      </div>
                    </div>
                  </CollapsibleContent>
                </Collapsible>
              </div>
            </TabsContent>

            <TabsContent value="vector" className="space-y-4 mt-6">
              {vectorFields.length === 0 ? (
                <Alert>
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>
                    No vector fields configured. Add vector fields to enable vector search.
                  </AlertDescription>
                </Alert>
              ) : (
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="vector-field">Vector Field</Label>
                    <Select value={selectedVectorField} onValueChange={setSelectedVectorField}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {vectorFields.map((field) => (
                          <SelectItem key={field.name} value={field.name}>
                            <div className="flex items-center justify-between w-full">
                              <span>{field.name}</span>
                              <Badge variant="secondary" className="ml-2">
                                {field.dimensions}D
                              </Badge>
                            </div>
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="vector-query">Query Vector (JSON array)</Label>
                    <Textarea
                      id="vector-query"
                      value={vectorQuery}
                      onChange={(e) => setVectorQuery(e.target.value)}
                      placeholder="[0.1, -0.2, 0.3, ...]"
                      className="font-mono text-sm"
                      rows={4}
                    />
                    <p className="text-xs text-muted-foreground">
                      Paste your embedding vector as a JSON array of numbers
                    </p>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="top-k">Top K Results</Label>
                      <Input
                        id="top-k"
                        type="number"
                        value={topK}
                        onChange={(e) => setTopK(e.target.value)}
                        placeholder="10"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="vector-filter">Pre-filter (Optional)</Label>
                      <Input
                        id="vector-filter"
                        value={vectorFilter}
                        onChange={(e) => setVectorFilter(e.target.value)}
                        placeholder='{"category": "electronics"}'
                        className="font-mono"
                      />
                    </div>
                  </div>
                </div>
              )}
            </TabsContent>

            <TabsContent value="hybrid" className="space-y-4 mt-6">
              {vectorFields.length === 0 ? (
                <Alert>
                  <AlertCircle className="h-4 w-4" />
                  <AlertDescription>
                    No vector fields configured. Add vector fields to enable hybrid search.
                  </AlertDescription>
                </Alert>
              ) : (
                <div className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="text-query">Text Query</Label>
                    <Input
                      id="text-query"
                      value={textQuery}
                      onChange={(e) => setTextQuery(e.target.value)}
                      placeholder="Search for products with wireless connectivity..."
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="hybrid-vector-field">Vector Field</Label>
                    <Select value={hybridVectorField} onValueChange={setHybridVectorField}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {vectorFields.map((field) => (
                          <SelectItem key={field.name} value={field.name}>
                            <div className="flex items-center justify-between w-full">
                              <span>{field.name}</span>
                              <Badge variant="secondary" className="ml-2">
                                {field.dimensions}D
                              </Badge>
                            </div>
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="hybrid-vector">Query Vector (JSON array)</Label>
                    <Textarea
                      id="hybrid-vector"
                      value={hybridVector}
                      onChange={(e) => setHybridVector(e.target.value)}
                      placeholder="[0.1, -0.2, 0.3, ...]"
                      className="font-mono text-sm"
                      rows={3}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="alpha">
                      Weight Balance (Alpha: {alpha[0].toFixed(2)})
                    </Label>
                    <div className="flex items-center gap-4">
                      <span className="text-sm text-muted-foreground">Text</span>
                      <Slider
                        value={alpha}
                        onValueChange={setAlpha}
                        min={0}
                        max={1}
                        step={0.1}
                        className="flex-1"
                      />
                      <span className="text-sm text-muted-foreground">Vector</span>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      0 = Text only, 1 = Vector only, 0.5 = Equal weight
                    </p>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="hybrid-top-k">Top K Results</Label>
                      <Input
                        id="hybrid-top-k"
                        type="number"
                        value={hybridTopK}
                        onChange={(e) => setHybridTopK(e.target.value)}
                        placeholder="10"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="hybrid-filter">Pre-filter (Optional)</Label>
                      <Input
                        id="hybrid-filter"
                        value={hybridFilter}
                        onChange={(e) => setHybridFilter(e.target.value)}
                        placeholder='{"status": "active"}'
                        className="font-mono"
                      />
                    </div>
                  </div>
                </div>
              )}
            </TabsContent>
          </Tabs>

          <div className="flex justify-end mt-6">
            <Button 
              onClick={executeQuery} 
              disabled={isLoading}
              className="gap-2"
            >
              {isLoading ? (
                <>
                  <div className="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent" />
                  Executing...
                </>
              ) : (
                <>
                  <Play className="h-4 w-4" />
                  Execute Query
                </>
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {results && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Query Results</CardTitle>
                <CardDescription>
                  {results.error ? (
                    <span className="text-destructive">Error: {results.error}</span>
                  ) : (
                    <span>
                      Found {results.count} results
                      {results.executionTime && ` in ${results.executionTime}ms`}
                    </span>
                  )}
                </CardDescription>
              </div>
              {!results.error && results.results.length > 0 && (
                <div className="flex gap-2">
                  <Button variant="outline" size="sm" onClick={copyResults}>
                    <Copy className="h-4 w-4 mr-2" />
                    Copy
                  </Button>
                  <Button variant="outline" size="sm" onClick={downloadResults}>
                    <Download className="h-4 w-4 mr-2" />
                    Download
                  </Button>
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent>
            {results.error ? (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{results.error}</AlertDescription>
              </Alert>
            ) : results.results.length === 0 ? (
              <Alert>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>No results found</AlertDescription>
              </Alert>
            ) : (
              <ScrollArea className="h-[600px] w-full rounded-md border p-4">
                <pre className="text-sm">
                  {JSON.stringify(results.results, null, 2)}
                </pre>
              </ScrollArea>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}