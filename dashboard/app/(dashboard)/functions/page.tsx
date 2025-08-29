"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { 
  Plus, Search, Code2, Play, Pause, Clock, Activity,
  Zap, Terminal, Edit, Trash2, MoreVertical, Settings
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

export default function FunctionsPage() {
  const [searchQuery, setSearchQuery] = useState("")

  // Mock data - replace with actual API calls
  const functions = [
    {
      id: "func-1",
      name: "process-payment",
      displayName: "Payment Processing",
      description: "Handles payment transactions and webhook callbacks",
      runtime: "Node.js 18",
      status: "active",
      stats: {
        invocations: 1543,
        avgDuration: 245,
        errorRate: 0.2
      }
    },
    {
      id: "func-2",
      name: "send-notifications",
      displayName: "Send Notifications",
      description: "Sends email and push notifications to users",
      runtime: "Python 3.11",
      status: "active",
      stats: {
        invocations: 3421,
        avgDuration: 180,
        errorRate: 0.1
      }
    },
    {
      id: "func-3",
      name: "generate-reports",
      displayName: "Generate Reports",
      description: "Creates PDF reports from aggregated data",
      runtime: "Node.js 18",
      status: "paused",
      stats: {
        invocations: 234,
        avgDuration: 1200,
        errorRate: 1.5
      }
    },
    {
      id: "func-4",
      name: "image-optimization",
      displayName: "Image Optimization",
      description: "Resizes and optimizes uploaded images",
      runtime: "Go 1.21",
      status: "active",
      stats: {
        invocations: 5678,
        avgDuration: 450,
        errorRate: 0.3
      }
    },
    {
      id: "func-5",
      name: "data-sync",
      displayName: "Data Sync",
      description: "Synchronizes data between services",
      runtime: "Python 3.11",
      status: "active",
      stats: {
        invocations: 892,
        avgDuration: 320,
        errorRate: 0.5
      }
    }
  ]

  const filteredFunctions = functions.filter(func => {
    if (!searchQuery) return true
    const searchLower = searchQuery.toLowerCase()
    return func.displayName.toLowerCase().includes(searchLower) ||
           func.description.toLowerCase().includes(searchLower)
  })

  const activeCount = functions.filter(f => f.status === "active").length
  const pausedCount = functions.filter(f => f.status === "paused").length
  const totalInvocations = functions.reduce((sum, f) => sum + f.stats.invocations, 0)
  const avgDuration = Math.round(functions.reduce((sum, f) => sum + f.stats.avgDuration, 0) / functions.length)

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
                <BreadcrumbPage>Functions</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-semibold">Functions</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Deploy and manage serverless functions
              </p>
            </div>
            <Button size="sm">
              <Plus className="h-4 w-4 mr-2" />
              New Function
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
              <span className="text-2xl font-semibold">{pausedCount}</span>
              <span className="text-sm text-muted-foreground">Paused</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{totalInvocations.toLocaleString()}</span>
              <span className="text-sm text-muted-foreground">Invocations</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{avgDuration}</span>
              <span className="text-sm text-muted-foreground">ms avg</span>
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
              placeholder="Search functions..." 
              className="pl-10"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>

        {/* Functions Table */}
        <div className="border">
        <table className="w-full">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="text-left p-4 font-medium text-sm">Function</th>
              <th className="text-left p-4 font-medium text-sm">Runtime</th>
              <th className="text-left p-4 font-medium text-sm">Status</th>
              <th className="text-right p-4 font-medium text-sm">Invocations</th>
              <th className="text-right p-4 font-medium text-sm">Avg Duration</th>
              <th className="text-right p-4 font-medium text-sm">Error Rate</th>
              <th className="w-10"></th>
            </tr>
          </thead>
          <tbody>
            {filteredFunctions.map((func, index) => (
              <tr 
                key={func.id} 
                className={`border-b hover:bg-muted/30 transition-colors cursor-pointer ${
                  index === filteredFunctions.length - 1 ? 'border-b-0' : ''
                }`}
                onClick={() => window.location.href = `/functions/${func.name}`}
              >
                <td className="p-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10">
                      <Terminal className="h-4 w-4 text-primary" />
                    </div>
                    <div>
                      <p className="font-medium">{func.displayName}</p>
                      <p className="text-sm text-muted-foreground">{func.description}</p>
                    </div>
                  </div>
                </td>
                <td className="p-4">
                  <Badge variant="outline" className="font-normal">
                    {func.runtime}
                  </Badge>
                </td>
                <td className="p-4">
                  <Badge 
                    variant={func.status === "active" ? "default" : "secondary"} 
                    className="font-normal"
                  >
                    {func.status === "active" ? 
                      <Play className="h-3 w-3 mr-1.5" /> : 
                      <Pause className="h-3 w-3 mr-1.5" />
                    }
                    {func.status}
                  </Badge>
                </td>
                <td className="p-4 text-right">
                  <div className="flex items-center justify-end gap-1">
                    <Activity className="h-3 w-3 text-muted-foreground" />
                    <span className="font-mono font-medium">{func.stats.invocations.toLocaleString()}</span>
                  </div>
                </td>
                <td className="p-4 text-right">
                  <div className="flex items-center justify-end gap-1">
                    <Clock className="h-3 w-3 text-muted-foreground" />
                    <span className="font-mono font-medium">{func.stats.avgDuration}</span>
                    <span className="text-xs text-muted-foreground">ms</span>
                  </div>
                </td>
                <td className="p-4 text-right">
                  <div className="flex items-center justify-end gap-1">
                    <Zap className="h-3 w-3 text-muted-foreground" />
                    <span className={`font-mono font-medium ${
                      func.stats.errorRate > 1 ? 'text-orange-600' : ''
                    }`}>
                      {func.stats.errorRate}%
                    </span>
                  </div>
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
                        <Code2 className="h-4 w-4 mr-2" />
                        View Code
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <Edit className="h-4 w-4 mr-2" />
                        Edit Function
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <Settings className="h-4 w-4 mr-2" />
                        Settings
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      {func.status === "active" ? (
                        <DropdownMenuItem>
                          <Pause className="h-4 w-4 mr-2" />
                          Pause Function
                        </DropdownMenuItem>
                      ) : (
                        <DropdownMenuItem>
                          <Play className="h-4 w-4 mr-2" />
                          Resume Function
                        </DropdownMenuItem>
                      )}
                      <DropdownMenuSeparator />
                      <DropdownMenuItem className="text-red-600">
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete Function
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
        {filteredFunctions.length === 0 && (
          <div className="text-center py-12 border">
            <p className="text-muted-foreground">No functions found</p>
          </div>
        )}
      </div>
    </div>
  )
}