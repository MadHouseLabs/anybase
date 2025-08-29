"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { 
  Plus, Search, Globe, Lock, Unlock, Zap,
  Activity, TrendingUp, AlertCircle, CheckCircle,
  MoreVertical, Settings, Trash2, Copy, ExternalLink,
  Shield, Code2, FileJson, Key
} from "lucide-react"
import Link from "next/link"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export default function APIsPage() {
  const [searchQuery, setSearchQuery] = useState("")

  // Mock data - replace with actual API calls
  const apis = [
    {
      id: "api-1",
      name: "users-api",
      displayName: "Users API",
      description: "User authentication and management endpoints",
      path: "/api/v1/users",
      version: "v1",
      method: ["GET", "POST", "PUT", "DELETE"],
      status: "active",
      auth: "api-key",
      rateLimit: "1000/hour",
      stats: {
        requests: 45234,
        avgLatency: 124,
        errorRate: 0.2,
        lastHour: 1543
      },
      functions: ["validateUser", "createUser", "updateProfile"],
      repository: "user-service"
    },
    {
      id: "api-2",
      name: "payments-api",
      displayName: "Payments API",
      description: "Payment processing and transaction management",
      path: "/api/v1/payments",
      version: "v1",
      method: ["POST", "GET"],
      status: "active",
      auth: "oauth2",
      rateLimit: "100/minute",
      stats: {
        requests: 23456,
        avgLatency: 345,
        errorRate: 0.5,
        lastHour: 892
      },
      functions: ["processPayment", "validateCard", "detectFraud"],
      repository: "payment-functions"
    },
    {
      id: "api-3",
      name: "products-api",
      displayName: "Products API",
      description: "Product catalog and inventory management",
      path: "/api/v2/products",
      version: "v2",
      method: ["GET", "POST", "PATCH"],
      status: "active",
      auth: "public",
      rateLimit: "5000/hour",
      stats: {
        requests: 123456,
        avgLatency: 56,
        errorRate: 0.1,
        lastHour: 5678
      },
      functions: ["searchProducts", "updateInventory"],
      repository: "product-service"
    },
    {
      id: "api-4",
      name: "analytics-api",
      displayName: "Analytics API",
      description: "Real-time analytics and reporting endpoints",
      path: "/api/v1/analytics",
      version: "v1",
      method: ["POST", "GET"],
      status: "deprecated",
      auth: "jwt",
      rateLimit: "500/hour",
      stats: {
        requests: 8765,
        avgLatency: 234,
        errorRate: 1.2,
        lastHour: 234
      },
      functions: ["trackEvent", "generateReport"],
      repository: "analytics-processors"
    },
    {
      id: "api-5",
      name: "webhooks-api",
      displayName: "Webhooks API",
      description: "Webhook registration and event delivery",
      path: "/api/v1/webhooks",
      version: "v1",
      method: ["POST", "DELETE"],
      status: "active",
      auth: "signature",
      rateLimit: "100/hour",
      stats: {
        requests: 12345,
        avgLatency: 567,
        errorRate: 0.8,
        lastHour: 456
      },
      functions: ["registerWebhook", "verifySignature", "deliverEvent"],
      repository: "webhook-service"
    }
  ]

  const filteredAPIs = apis.filter(api => {
    if (!searchQuery) return true
    const searchLower = searchQuery.toLowerCase()
    return api.displayName.toLowerCase().includes(searchLower) ||
           api.description.toLowerCase().includes(searchLower) ||
           api.path.toLowerCase().includes(searchLower)
  })

  const activeCount = apis.filter(a => a.status === "active").length
  const deprecatedCount = apis.filter(a => a.status === "deprecated").length
  const totalRequests = apis.reduce((sum, a) => sum + a.stats.lastHour, 0)
  const avgLatency = Math.round(apis.reduce((sum, a) => sum + a.stats.avgLatency, 0) / apis.length)

  const getAuthIcon = (auth: string) => {
    switch(auth) {
      case 'api-key': return <Key className="h-3 w-3" />
      case 'oauth2': return <Shield className="h-3 w-3" />
      case 'jwt': return <Lock className="h-3 w-3" />
      case 'signature': return <Shield className="h-3 w-3" />
      case 'public': return <Unlock className="h-3 w-3" />
      default: return <Lock className="h-3 w-3" />
    }
  }

  const getStatusBadge = (status: string) => {
    switch(status) {
      case 'active':
        return (
          <Badge variant="default" className="font-normal">
            <CheckCircle className="h-3 w-3 mr-1.5" />
            Active
          </Badge>
        )
      case 'deprecated':
        return (
          <Badge variant="secondary" className="font-normal">
            <AlertCircle className="h-3 w-3 mr-1.5" />
            Deprecated
          </Badge>
        )
      default:
        return null
    }
  }

  return (
    <div className="flex flex-col min-h-screen">
      {/* Header */}
      <div className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container mx-auto px-6 py-4 max-w-7xl">
          {/* Breadcrumbs */}
          <Breadcrumb className="mb-4">
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbLink href="/">Dashboard</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>APIs</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-semibold">API Endpoints</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Manage and monitor your API endpoints
              </p>
            </div>
            <Button size="sm">
              <Plus className="h-4 w-4 mr-2" />
              Create API
            </Button>
          </div>

          {/* Metrics */}
          <div className="flex items-center gap-6 pt-4 border-t">
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{activeCount}</span>
              <span className="text-sm text-muted-foreground">Active</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{totalRequests.toLocaleString()}</span>
              <span className="text-sm text-muted-foreground">Requests/hr</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{avgLatency}ms</span>
              <span className="text-sm text-muted-foreground">Avg Latency</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">99.8%</span>
              <span className="text-sm text-muted-foreground">Uptime</span>
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="container mx-auto px-6 py-6 max-w-7xl">

        {/* Search */}
        <div className="mb-6">
          <div className="relative max-w-md">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <Input 
              placeholder="Search APIs..." 
              className="pl-10"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>

        {/* APIs Table */}
        <div className="border">
        <table className="w-full">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="text-left p-4 font-medium text-sm">API Endpoint</th>
              <th className="text-left p-4 font-medium text-sm">Methods</th>
              <th className="text-left p-4 font-medium text-sm">Auth</th>
              <th className="text-left p-4 font-medium text-sm">Status</th>
              <th className="text-right p-4 font-medium text-sm">Requests/hr</th>
              <th className="text-right p-4 font-medium text-sm">Avg Latency</th>
              <th className="text-right p-4 font-medium text-sm">Error Rate</th>
              <th className="w-10"></th>
            </tr>
          </thead>
          <tbody>
            {filteredAPIs.map((api, index) => (
              <tr 
                key={api.id} 
                className={`border-b hover:bg-muted/30 transition-colors cursor-pointer ${
                  index === filteredAPIs.length - 1 ? 'border-b-0' : ''
                }`}
                onClick={() => window.location.href = `/apis/${api.name}`}
              >
                <td className="p-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10">
                      <Globe className="h-4 w-4 text-primary" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <p className="font-medium">{api.displayName}</p>
                        <Badge variant="outline" className="text-xs font-normal">
                          {api.version}
                        </Badge>
                      </div>
                      <code className="text-xs text-muted-foreground">{api.path}</code>
                    </div>
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex gap-1 flex-wrap">
                    {api.method.map(method => (
                      <Badge key={method} variant="secondary" className="text-xs font-normal">
                        {method}
                      </Badge>
                    ))}
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-1.5">
                    {getAuthIcon(api.auth)}
                    <span className="text-sm">{api.auth}</span>
                  </div>
                </td>
                <td className="p-4">
                  {getStatusBadge(api.status)}
                </td>
                <td className="p-4 text-right">
                  <div className="flex items-center justify-end gap-1">
                    <Activity className="h-3 w-3 text-muted-foreground" />
                    <span className="font-mono font-medium">{api.stats.lastHour.toLocaleString()}</span>
                  </div>
                </td>
                <td className="p-4 text-right">
                  <span className="font-mono font-medium">{api.stats.avgLatency}ms</span>
                </td>
                <td className="p-4 text-right">
                  <span className={`font-mono font-medium ${
                    api.stats.errorRate > 1 ? 'text-orange-600' : ''
                  }`}>
                    {api.stats.errorRate}%
                  </span>
                </td>
                <td className="p-4">
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                      <Button variant="ghost" size="icon" className="h-8 w-8">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem>
                        <FileJson className="h-4 w-4 mr-2" />
                        View OpenAPI Spec
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <Code2 className="h-4 w-4 mr-2" />
                        View Functions
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <Activity className="h-4 w-4 mr-2" />
                        View Metrics
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <ExternalLink className="h-4 w-4 mr-2" />
                        Test Endpoint
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <Settings className="h-4 w-4 mr-2" />
                        Settings
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      {api.status === 'active' && (
                        <DropdownMenuItem>
                          <AlertCircle className="h-4 w-4 mr-2" />
                          Deprecate API
                        </DropdownMenuItem>
                      )}
                      <DropdownMenuItem className="text-red-600">
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete API
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>

        {/* Empty State */}
        {filteredAPIs.length === 0 && (
          <div className="text-center py-12 border">
            <p className="text-muted-foreground">No APIs found</p>
          </div>
        )}
      </div>
    </div>
  )
}