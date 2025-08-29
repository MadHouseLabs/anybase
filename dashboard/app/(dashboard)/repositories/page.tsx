"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { 
  Plus, Search, GitBranch, GitCommit, GitPullRequest,
  Github, Gitlab, Cloud, Link2, CheckCircle, AlertCircle,
  RefreshCw, MoreVertical, Settings, Trash2, ExternalLink,
  FolderGit2
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

export default function RepositoriesPage() {
  const [searchQuery, setSearchQuery] = useState("")

  // Mock data - replace with actual API calls
  const repositories = [
    {
      id: "repo-1",
      name: "payment-functions",
      displayName: "Payment Functions",
      description: "Payment processing and validation functions",
      url: "https://github.com/company/payment-functions",
      provider: "github",
      branch: "main",
      status: "connected",
      lastSync: "2024-01-15T10:30:00Z",
      functions: 12,
      language: "TypeScript"
    },
    {
      id: "repo-2",
      name: "notification-service",
      displayName: "Notification Service",
      description: "Email and SMS notification handlers",
      url: "https://gitlab.com/company/notification-service",
      provider: "gitlab",
      branch: "master",
      status: "connected",
      lastSync: "2024-01-15T09:45:00Z",
      functions: 8,
      language: "Python"
    },
    {
      id: "repo-3",
      name: "data-transformers",
      displayName: "Data Transformers",
      description: "ETL and data transformation functions",
      url: "https://github.com/company/data-transformers",
      provider: "github",
      branch: "main",
      status: "syncing",
      lastSync: "2024-01-15T11:00:00Z",
      functions: 24,
      language: "Go"
    },
    {
      id: "repo-4",
      name: "analytics-processors",
      displayName: "Analytics Processors",
      description: "Real-time analytics and aggregation functions",
      url: "https://bitbucket.org/company/analytics",
      provider: "bitbucket",
      branch: "develop",
      status: "error",
      lastSync: "2024-01-14T15:20:00Z",
      functions: 6,
      language: "JavaScript",
      error: "Authentication failed"
    },
    {
      id: "repo-5",
      name: "ml-models",
      displayName: "ML Models",
      description: "Machine learning inference functions",
      url: "https://github.com/company/ml-models",
      provider: "github",
      branch: "production",
      status: "connected",
      lastSync: "2024-01-15T08:00:00Z",
      functions: 15,
      language: "Python"
    }
  ]

  const filteredRepos = repositories.filter(repo => {
    if (!searchQuery) return true
    const searchLower = searchQuery.toLowerCase()
    return repo.displayName.toLowerCase().includes(searchLower) ||
           repo.description.toLowerCase().includes(searchLower) ||
           repo.language.toLowerCase().includes(searchLower)
  })

  const connectedCount = repositories.filter(r => r.status === "connected").length
  const totalFunctions = repositories.reduce((sum, r) => sum + r.functions, 0)
  const uniqueLanguages = [...new Set(repositories.map(r => r.language))].length

  const getProviderIcon = (provider: string) => {
    switch(provider) {
      case 'github': return <Github className="h-4 w-4" />
      case 'gitlab': return <Gitlab className="h-4 w-4" />
      case 'bitbucket': return <Cloud className="h-4 w-4" />
      default: return <FolderGit2 className="h-4 w-4" />
    }
  }

  const getStatusBadge = (status: string, error?: string) => {
    switch(status) {
      case 'connected':
        return (
          <Badge variant="default" className="font-normal">
            <CheckCircle className="h-3 w-3 mr-1.5" />
            Connected
          </Badge>
        )
      case 'syncing':
        return (
          <Badge variant="secondary" className="font-normal">
            <RefreshCw className="h-3 w-3 mr-1.5 animate-spin" />
            Syncing
          </Badge>
        )
      case 'error':
        return (
          <Badge variant="destructive" className="font-normal" title={error}>
            <AlertCircle className="h-3 w-3 mr-1.5" />
            Error
          </Badge>
        )
      default:
        return null
    }
  }

  const formatLastSync = (dateStr: string) => {
    const date = new Date(dateStr)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(minutes / 60)
    const days = Math.floor(hours / 24)

    if (days > 0) return `${days}d ago`
    if (hours > 0) return `${hours}h ago`
    if (minutes > 0) return `${minutes}m ago`
    return 'Just now'
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
                <BreadcrumbPage>Repositories</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-semibold">Code Repositories</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Connect and manage source code repositories for functions
              </p>
            </div>
            <Button size="sm">
              <Plus className="h-4 w-4 mr-2" />
              Connect Repository
            </Button>
          </div>

          {/* Metrics */}
          <div className="flex items-center gap-6 pt-4 border-t">
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{repositories.length}</span>
              <span className="text-sm text-muted-foreground">Repositories</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{connectedCount}</span>
              <span className="text-sm text-muted-foreground">Connected</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{totalFunctions}</span>
              <span className="text-sm text-muted-foreground">Functions</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{uniqueLanguages}</span>
              <span className="text-sm text-muted-foreground">Languages</span>
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
              placeholder="Search repositories..." 
              className="pl-10"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>

        {/* Repositories Table */}
        <div className="border">
        <table className="w-full">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="text-left p-4 font-medium text-sm">Repository</th>
              <th className="text-left p-4 font-medium text-sm">Provider</th>
              <th className="text-left p-4 font-medium text-sm">Branch</th>
              <th className="text-left p-4 font-medium text-sm">Status</th>
              <th className="text-right p-4 font-medium text-sm">Functions</th>
              <th className="text-right p-4 font-medium text-sm">Last Sync</th>
              <th className="w-10"></th>
            </tr>
          </thead>
          <tbody>
            {filteredRepos.map((repo, index) => (
              <tr 
                key={repo.id} 
                className={`border-b hover:bg-muted/30 transition-colors cursor-pointer ${
                  index === filteredRepos.length - 1 ? 'border-b-0' : ''
                }`}
                onClick={() => window.location.href = `/repositories/${repo.name}`}
              >
                <td className="p-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10">
                      <FolderGit2 className="h-4 w-4 text-primary" />
                    </div>
                    <div>
                      <p className="font-medium">{repo.displayName}</p>
                      <p className="text-sm text-muted-foreground">{repo.description}</p>
                    </div>
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-2">
                    {getProviderIcon(repo.provider)}
                    <span className="text-sm capitalize">{repo.provider}</span>
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-1">
                    <GitBranch className="h-3 w-3 text-muted-foreground" />
                    <code className="text-sm">{repo.branch}</code>
                  </div>
                </td>
                <td className="p-4">
                  {getStatusBadge(repo.status, repo.error)}
                </td>
                <td className="p-4 text-right">
                  <div className="flex items-center justify-end gap-2">
                    <Badge variant="outline" className="font-normal">
                      {repo.language}
                    </Badge>
                    <span className="font-mono font-medium">{repo.functions}</span>
                  </div>
                </td>
                <td className="p-4 text-right">
                  <span className="text-sm text-muted-foreground">
                    {formatLastSync(repo.lastSync)}
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
                        <FolderGit2 className="h-4 w-4 mr-2" />
                        View Functions
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <RefreshCw className="h-4 w-4 mr-2" />
                        Sync Now
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <ExternalLink className="h-4 w-4 mr-2" />
                        Open in {repo.provider}
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <Settings className="h-4 w-4 mr-2" />
                        Settings
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem className="text-red-600">
                        <Trash2 className="h-4 w-4 mr-2" />
                        Disconnect
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
        {filteredRepos.length === 0 && (
          <div className="text-center py-12 border">
            <p className="text-muted-foreground">No repositories found</p>
          </div>
        )}
      </div>
    </div>
  )
}