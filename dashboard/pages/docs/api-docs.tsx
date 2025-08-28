import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { 
  Key, 
  Code, 
  Database, 
  Shield,
  FileJson,
  ChevronRight,
  Copy,
  CheckCircle,
  Lock,
  Globe,
  Terminal,
  BookOpen,
  AlertCircle
} from "lucide-react"
import { useToast } from "@/hooks/use-toast"

export function ApiDocsPage() {
  const { toast } = useToast()
  const [copiedCode, setCopiedCode] = useState<string | null>(null)
  
  const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

  const copyToClipboard = (code: string, id: string) => {
    navigator.clipboard.writeText(code)
    setCopiedCode(id)
    toast({
      title: "Copied to clipboard",
      description: "Code snippet copied successfully",
    })
    setTimeout(() => setCopiedCode(null), 2000)
  }

  const CodeBlock = ({ code, language = "bash", id }: { code: string; language?: string; id: string }) => (
    <div className="relative group">
      <pre className={`language-${language} bg-slate-950 text-slate-50 p-4 rounded-lg overflow-x-auto text-sm`}>
        <code>{code}</code>
      </pre>
      <Button
        size="sm"
        variant="ghost"
        className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity"
        onClick={() => copyToClipboard(code, id)}
      >
        {copiedCode === id ? (
          <CheckCircle className="h-4 w-4 text-green-500" />
        ) : (
          <Copy className="h-4 w-4" />
        )}
      </Button>
    </div>
  )

  return (
    <div className="container mx-auto py-6 px-4 lg:px-8">
      <div className="max-w-7xl mx-auto space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Access Key API Documentation</h1>
          <p className="text-muted-foreground mt-2">
            Guide to using Anybase API with Bearer token authentication (Access Keys)
          </p>
        </div>

      <Alert className="bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-900">
        <div className="flex items-start gap-2">
          <Globe className="h-4 w-4 text-blue-600 mt-0.5" />
          <AlertDescription className="text-blue-900 dark:text-blue-100">
            <strong>Base URL:</strong> {API_BASE}/api/v1
          </AlertDescription>
        </div>
      </Alert>

      <Alert className="border-amber-200 bg-amber-50 dark:bg-amber-950/20 dark:border-amber-900">
        <div className="flex items-start gap-2">
          <AlertCircle className="h-4 w-4 text-amber-600 mt-0.5" />
          <AlertDescription className="text-amber-900 dark:text-amber-100">
            <strong>Important:</strong> Use <code className="bg-amber-100 dark:bg-amber-900 px-1 rounded">Authorization: Bearer &lt;access_key&gt;</code> header format, not <code>X-API-Key</code>
          </AlertDescription>
        </div>
      </Alert>

      <Tabs defaultValue="quickstart" className="space-y-4">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="quickstart">Quick Start</TabsTrigger>
          <TabsTrigger value="collections">Collections</TabsTrigger>
          <TabsTrigger value="views">Views</TabsTrigger>
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
        </TabsList>

        {/* Quick Start Tab */}
        <TabsContent value="quickstart" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Key className="h-5 w-5" />
                Getting Started with Access Keys
              </CardTitle>
              <CardDescription>
                Access keys provide programmatic API access with specific permissions
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">Step 1: Create an Access Key</h3>
                <p className="text-sm text-muted-foreground">
                  Access keys are created through the dashboard by administrators or developers. Each key can have specific permissions.
                </p>
                <Alert>
                  <Terminal className="h-4 w-4" />
                  <AlertDescription>
                    Access keys are created in the dashboard under "Access Keys" section. You cannot create access keys via API.
                  </AlertDescription>
                </Alert>
                
                <h3 className="text-lg font-semibold">Step 2: Use the Access Key</h3>
                <CodeBlock
                  id="basic-usage"
                  code={`# Example: List documents from a collection
curl -X GET "${API_BASE}/api/v1/data/products" \\
  -H "Authorization: Bearer ak_your_access_key_here"

# Example: Execute a view
curl -X GET "${API_BASE}/api/v1/views/active_products/query?limit=10" \\
  -H "Authorization: Bearer ak_your_access_key_here"`}
                />

                <h3 className="text-lg font-semibold">Response Format</h3>
                <CodeBlock
                  id="response-format"
                  language="json"
                  code={`// Successful response
{
  "data": [...],
  "pagination": {
    "limit": 10,
    "skip": 0,
    "count": 10
  }
}

// Error response
{
  "error": "Access denied: missing permission collection:products:read"
}`}
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Collections Tab */}
        <TabsContent value="collections" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Database className="h-5 w-5" />
                Collections API with Access Keys
              </CardTitle>
              <CardDescription>
                Access keys can perform CRUD operations on collections based on their permissions
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <Alert className="mb-4">
                <Lock className="h-4 w-4" />
                <AlertDescription>
                  Required permissions: <code>collection:&lt;name&gt;:read</code>, <code>collection:&lt;name&gt;:write</code>, <code>collection:&lt;name&gt;:delete</code>
                </AlertDescription>
              </Alert>

              <div className="space-y-4">
                <h3 className="text-lg font-semibold">Query Documents</h3>
                <p className="text-sm text-muted-foreground">Requires: <code>collection:products:read</code></p>
                <CodeBlock
                  id="query-collection"
                  code={`# Simple query
curl -X GET "${API_BASE}/api/v1/data/products?limit=20&skip=0" \\
  -H "Authorization: Bearer ak_your_access_key"

# With filter (URL encoded)
curl -X GET "${API_BASE}/api/v1/data/products?filter=%7B%22category%22%3A%22electronics%22%7D&limit=10" \\
  -H "Authorization: Bearer ak_your_access_key"

# Decoded filter: {"category":"electronics"}`}
                />

                <h3 className="text-lg font-semibold">Insert Document</h3>
                <p className="text-sm text-muted-foreground">Requires: <code>collection:products:write</code></p>
                <CodeBlock
                  id="insert-document"
                  code={`curl -X POST "${API_BASE}/api/v1/data/products" \\
  -H "Authorization: Bearer ak_your_access_key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "name": "New Product",
    "price": 99.99,
    "category": "electronics",
    "inStock": true
  }'`}
                />

                <h3 className="text-lg font-semibold">Update Document</h3>
                <p className="text-sm text-muted-foreground">Requires: <code>collection:products:write</code></p>
                <CodeBlock
                  id="update-document"
                  code={`curl -X PUT "${API_BASE}/api/v1/data/products/507f1f77bcf86cd799439011" \\
  -H "Authorization: Bearer ak_your_access_key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "price": 89.99,
    "inStock": false
  }'`}
                />

                <h3 className="text-lg font-semibold">Delete Document</h3>
                <p className="text-sm text-muted-foreground">Requires: <code>collection:products:delete</code></p>
                <CodeBlock
                  id="delete-document"
                  code={`curl -X DELETE "${API_BASE}/api/v1/data/products/507f1f77bcf86cd799439011" \\
  -H "Authorization: Bearer ak_your_access_key"`}
                />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Access Key Limitations</CardTitle>
            </CardHeader>
            <CardContent>
              <Alert className="border-red-200 bg-red-50 dark:bg-red-950/20 dark:border-red-900">
                <AlertCircle className="h-4 w-4 text-red-600" />
                <AlertDescription className="text-red-900 dark:text-red-100">
                  <strong>Access keys CANNOT:</strong>
                  <ul className="list-disc list-inside mt-2 space-y-1">
                    <li>Create, update, or delete collections</li>
                    <li>Modify collection schemas</li>
                    <li>View collection metadata</li>
                    <li>List all collections</li>
                  </ul>
                  These operations require JWT authentication (user login).
                </AlertDescription>
              </Alert>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Views Tab */}
        <TabsContent value="views" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Code className="h-5 w-5" />
                Views API with Access Keys
              </CardTitle>
              <CardDescription>
                Execute pre-defined views with runtime parameters
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <Alert className="mb-4">
                <Lock className="h-4 w-4" />
                <AlertDescription>
                  Required permission: <code>view:&lt;view_name&gt;:execute</code> or <code>view:*:execute</code>
                </AlertDescription>
              </Alert>

              <Alert className="border-red-200 bg-red-50 dark:bg-red-950/20 dark:border-red-900 mb-4">
                <AlertCircle className="h-4 w-4 text-red-600" />
                <AlertDescription className="text-red-900 dark:text-red-100">
                  <strong>Important:</strong> Access keys can ONLY execute views. They cannot create, list, update, or delete view definitions.
                </AlertDescription>
              </Alert>

              <div className="space-y-4">
                <h3 className="text-lg font-semibold">Execute View</h3>
                <p className="text-sm text-muted-foreground mb-2">
                  Views are pre-defined queries with base filters and field projections. You can add runtime parameters.
                </p>
                
                <h4 className="text-sm font-medium">Basic Execution</h4>
                <CodeBlock
                  id="execute-view-basic"
                  code={`curl -X GET "${API_BASE}/api/v1/views/active_products/query" \\
  -H "Authorization: Bearer ak_your_access_key"`}
                />

                <h4 className="text-sm font-medium">With Pagination</h4>
                <CodeBlock
                  id="execute-view-pagination"
                  code={`curl -X GET "${API_BASE}/api/v1/views/active_products/query?limit=10&skip=20" \\
  -H "Authorization: Bearer ak_your_access_key"`}
                />

                <h4 className="text-sm font-medium">With Runtime Filters</h4>
                <p className="text-xs text-muted-foreground">Runtime filters are applied ON TOP of the view's base filter</p>
                <CodeBlock
                  id="execute-view-filter"
                  code={`# URL encoded filter: {"category":"electronics"}
curl -X GET "${API_BASE}/api/v1/views/active_products/query?filter=%7B%22category%22%3A%22electronics%22%7D" \\
  -H "Authorization: Bearer ak_your_access_key"`}
                />

                <h4 className="text-sm font-medium">With Sort</h4>
                <CodeBlock
                  id="execute-view-sort"
                  code={`# URL encoded sort: {"price":-1} (descending)
curl -X GET "${API_BASE}/api/v1/views/active_products/query?sort=%7B%22price%22%3A-1%7D" \\
  -H "Authorization: Bearer ak_your_access_key"`}
                />

                <h4 className="text-sm font-medium">Combined Parameters</h4>
                <CodeBlock
                  id="execute-view-combined"
                  code={`curl -X GET "${API_BASE}/api/v1/views/active_products/query?limit=5&skip=0&filter=%7B%22category%22%3A%22electronics%22%7D&sort=%7B%22price%22%3A-1%7D" \\
  -H "Authorization: Bearer ak_your_access_key"`}
                />

                <h3 className="text-lg font-semibold">Response Format</h3>
                <CodeBlock
                  id="view-response-format"
                  language="json"
                  code={`{
  "data": [
    {
      "_id": "507f1f77bcf86cd799439011",
      "name": "Product Name",
      "price": 199.99,
      "category": "electronics"
      // Only fields specified in view's projection
    }
  ],
  "pagination": {
    "limit": 5,
    "skip": 0,
    "count": 5  // Number of records returned
  }
}`}
                />

                <h3 className="text-lg font-semibold">Query Parameters Reference</h3>
                <div className="border rounded-lg p-4">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b">
                        <th className="text-left py-2">Parameter</th>
                        <th className="text-left py-2">Type</th>
                        <th className="text-left py-2">Description</th>
                        <th className="text-left py-2">Example</th>
                      </tr>
                    </thead>
                    <tbody className="space-y-2">
                      <tr className="border-b">
                        <td className="py-2"><code>limit</code></td>
                        <td className="py-2">number</td>
                        <td className="py-2">Max records to return (default: 100)</td>
                        <td className="py-2"><code>limit=10</code></td>
                      </tr>
                      <tr className="border-b">
                        <td className="py-2"><code>skip</code></td>
                        <td className="py-2">number</td>
                        <td className="py-2">Records to skip for pagination</td>
                        <td className="py-2"><code>skip=20</code></td>
                      </tr>
                      <tr className="border-b">
                        <td className="py-2"><code>filter</code></td>
                        <td className="py-2">JSON (URL encoded)</td>
                        <td className="py-2">Additional filters on top of view filter</td>
                        <td className="py-2"><code>filter=%7B%22status%22%3A%22active%22%7D</code></td>
                      </tr>
                      <tr>
                        <td className="py-2"><code>sort</code></td>
                        <td className="py-2">JSON (URL encoded)</td>
                        <td className="py-2">Sort order (1: asc, -1: desc)</td>
                        <td className="py-2"><code>sort=%7B%22price%22%3A-1%7D</code></td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Permissions Tab */}
        <TabsContent value="permissions" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Shield className="h-5 w-5" />
                Access Key Permissions
              </CardTitle>
              <CardDescription>
                Understanding the permission system for access keys
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <h3 className="text-lg font-semibold">Permission Format</h3>
                <div className="p-4 bg-muted rounded-lg">
                  <code className="text-lg">resource:identifier:action</code>
                  <div className="mt-2 text-sm text-muted-foreground">
                    <div><strong>resource:</strong> The type of resource (collection, view)</div>
                    <div><strong>identifier:</strong> Specific resource name or * for all</div>
                    <div><strong>action:</strong> The operation (read, write, delete, execute)</div>
                  </div>
                </div>

                <h3 className="text-lg font-semibold">Collection Permissions</h3>
                <div className="space-y-2">
                  <div className="p-3 border rounded-lg">
                    <code className="font-mono">collection:products:read</code>
                    <p className="text-sm text-muted-foreground mt-1">Read documents from the products collection</p>
                  </div>
                  <div className="p-3 border rounded-lg">
                    <code className="font-mono">collection:products:write</code>
                    <p className="text-sm text-muted-foreground mt-1">Create or update documents in products collection</p>
                  </div>
                  <div className="p-3 border rounded-lg">
                    <code className="font-mono">collection:products:delete</code>
                    <p className="text-sm text-muted-foreground mt-1">Delete documents from products collection</p>
                  </div>
                  <div className="p-3 border rounded-lg">
                    <code className="font-mono">collection:*:read</code>
                    <p className="text-sm text-muted-foreground mt-1">Read from any collection (use with caution)</p>
                  </div>
                </div>

                <h3 className="text-lg font-semibold">View Permissions</h3>
                <div className="space-y-2">
                  <div className="p-3 border rounded-lg">
                    <code className="font-mono">view:sales_report:execute</code>
                    <p className="text-sm text-muted-foreground mt-1">Execute the sales_report view</p>
                  </div>
                  <div className="p-3 border rounded-lg">
                    <code className="font-mono">view:*:execute</code>
                    <p className="text-sm text-muted-foreground mt-1">Execute any view (use with caution)</p>
                  </div>
                </div>

                <h3 className="text-lg font-semibold">Wildcard Permission</h3>
                <div className="p-3 border border-red-300 bg-red-50 dark:bg-red-950/20 rounded-lg">
                  <code className="font-mono">*:*:*</code>
                  <p className="text-sm text-red-700 dark:text-red-300 mt-1">
                    Full access to all resources. Use only for testing or admin tools.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Example Permission Sets</CardTitle>
              <CardDescription>Common access key configurations</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-4">
                <div>
                  <h4 className="font-medium mb-2">Read-Only Analytics</h4>
                  <CodeBlock
                    id="perm-analytics"
                    language="json"
                    code={`[
  "collection:events:read",
  "collection:users:read",
  "view:daily_stats:execute",
  "view:user_analytics:execute"
]`}
                  />
                </div>

                <div>
                  <h4 className="font-medium mb-2">Product Management Service</h4>
                  <CodeBlock
                    id="perm-products"
                    language="json"
                    code={`[
  "collection:products:read",
  "collection:products:write",
  "collection:inventory:read",
  "collection:inventory:write",
  "view:low_stock_alert:execute"
]`}
                  />
                </div>

                <div>
                  <h4 className="font-medium mb-2">Reporting Dashboard</h4>
                  <CodeBlock
                    id="perm-reporting"
                    language="json"
                    code={`[
  "view:sales_summary:execute",
  "view:inventory_status:execute",
  "view:customer_metrics:execute",
  "view:revenue_report:execute"
]`}
                  />
                </div>

                <div>
                  <h4 className="font-medium mb-2">Data Import Tool</h4>
                  <CodeBlock
                    id="perm-import"
                    language="json"
                    code={`[
  "collection:products:write",
  "collection:customers:write",
  "collection:orders:write"
]`}
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Security Best Practices</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="flex items-start gap-3">
                  <Badge className="mt-1">1</Badge>
                  <div>
                    <p className="font-medium">Principle of Least Privilege</p>
                    <p className="text-sm text-muted-foreground">
                      Grant only the minimum permissions required. Avoid wildcards unless necessary.
                    </p>
                  </div>
                </div>
                
                <div className="flex items-start gap-3">
                  <Badge className="mt-1">2</Badge>
                  <div>
                    <p className="font-medium">Use Environment Variables</p>
                    <p className="text-sm text-muted-foreground">
                      Never hardcode access keys in source code. Use environment variables or secret management systems.
                    </p>
                  </div>
                </div>
                
                <div className="flex items-start gap-3">
                  <Badge className="mt-1">3</Badge>
                  <div>
                    <p className="font-medium">Set Expiration Dates</p>
                    <p className="text-sm text-muted-foreground">
                      Always set an expiration date for access keys and rotate them regularly.
                    </p>
                  </div>
                </div>
                
                <div className="flex items-start gap-3">
                  <Badge className="mt-1">4</Badge>
                  <div>
                    <p className="font-medium">Monitor Usage</p>
                    <p className="text-sm text-muted-foreground">
                      Regularly review access logs and deactivate unused keys immediately.
                    </p>
                  </div>
                </div>
                
                <div className="flex items-start gap-3">
                  <Badge className="mt-1">5</Badge>
                  <div>
                    <p className="font-medium">Use Views for Complex Queries</p>
                    <p className="text-sm text-muted-foreground">
                      Instead of giving broad collection access, create specific views with controlled filters and projections.
                    </p>
                  </div>
                </div>

                <div className="flex items-start gap-3">
                  <Badge className="mt-1">6</Badge>
                  <div>
                    <p className="font-medium">Separate Keys by Environment</p>
                    <p className="text-sm text-muted-foreground">
                      Use different access keys for development, staging, and production environments.
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
      </div>
    </div>
  )
}