"use client"

import { useState } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { 
  Plus, Search, Layers, Pause, Play, Clock, Activity
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

export default function EventStorePage() {
  const [searchQuery, setSearchQuery] = useState("")

  // Mock data - replace with actual API calls
  const processors = [
    {
      id: "proc-1",
      name: "order-processing",
      displayName: "Order Processing",
      description: "Handles incoming customer orders and inventory updates",
      status: "active",
      stats: {
        pending: 3,
        throughput: 45,
        avgLatency: 243
      }
    },
    {
      id: "proc-2",
      name: "email-notifications",
      displayName: "Email Notifications",
      description: "Sends transactional emails and customer notifications",
      status: "active",
      stats: {
        pending: 2,
        throughput: 120,
        avgLatency: 89
      }
    },
    {
      id: "proc-3",
      name: "image-processing",
      displayName: "Image Processing",
      description: "Resizes and optimizes uploaded images",
      status: "active",
      stats: {
        pending: 15,
        throughput: 15,
        avgLatency: 1250
      }
    },
    {
      id: "proc-4",
      name: "payment-processing",
      displayName: "Payment Processing",
      description: "Processes payment transactions and reconciliation",
      status: "active",
      stats: {
        pending: 1,
        throughput: 8,
        avgLatency: 450
      }
    },
    {
      id: "proc-5",
      name: "analytics-events",
      displayName: "Analytics Events",
      description: "Processes user behavior and analytics events",
      status: "paused",
      stats: {
        pending: 147,
        throughput: 0,
        avgLatency: 0
      }
    }
  ]

  const filteredProcessors = processors.filter(proc => {
    if (!searchQuery) return true
    const searchLower = searchQuery.toLowerCase()
    return proc.displayName.toLowerCase().includes(searchLower) ||
           proc.description.toLowerCase().includes(searchLower)
  })

  const activeCount = processors.filter(p => p.status === "active").length
  const pausedCount = processors.filter(p => p.status === "paused").length
  const totalPending = processors.reduce((sum, p) => sum + p.stats.pending, 0)
  const totalThroughput = processors.reduce((sum, p) => sum + p.stats.throughput, 0)

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
                <BreadcrumbPage>Event Store</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-semibold">Event Store</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Real-time event processing and stream management
              </p>
            </div>
            <div className="flex gap-2">
              <Link href="/event-store/events">
                <Button variant="outline" size="sm">
                  <Layers className="h-4 w-4 mr-2" />
                  Events
                </Button>
              </Link>
              <Button size="sm">
                <Plus className="h-4 w-4 mr-2" />
                New Processor
              </Button>
            </div>
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
              <span className="text-2xl font-semibold">{totalPending}</span>
              <span className="text-sm text-muted-foreground">Pending Events</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{totalThroughput}</span>
              <span className="text-sm text-muted-foreground">Events/min</span>
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
              placeholder="Search processors..." 
              className="pl-10"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>

        {/* Processors Table */}
        <div className="border">
        <table className="w-full">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="text-left p-4 font-medium text-sm">Processor</th>
              <th className="text-left p-4 font-medium text-sm">Status</th>
              <th className="text-right p-4 font-medium text-sm">Pending</th>
              <th className="text-right p-4 font-medium text-sm">Throughput</th>
              <th className="text-right p-4 font-medium text-sm">Avg Latency</th>
              <th className="w-10"></th>
            </tr>
          </thead>
          <tbody>
            {filteredProcessors.map((proc, index) => (
              <tr 
                key={proc.id} 
                className={`border-b hover:bg-muted/30 transition-colors cursor-pointer ${
                  index === filteredProcessors.length - 1 ? 'border-b-0' : ''
                }`}
                onClick={() => window.location.href = `/event-store/${proc.name}`}
              >
                <td className="p-4">
                  <div>
                    <p className="font-medium">{proc.displayName}</p>
                    <p className="text-sm text-muted-foreground">{proc.description}</p>
                  </div>
                </td>
                <td className="p-4">
                  <Badge 
                    variant={proc.status === "active" ? "default" : "secondary"} 
                    className="font-normal"
                  >
                    {proc.status === "active" ? 
                      <Play className="h-3 w-3 mr-1.5" /> : 
                      <Pause className="h-3 w-3 mr-1.5" />
                    }
                    {proc.status}
                  </Badge>
                </td>
                <td className="p-4 text-right">
                  <span className="font-mono font-medium">{proc.stats.pending}</span>
                </td>
                <td className="p-4 text-right">
                  <div className="flex items-center justify-end gap-1">
                    <span className="font-mono font-medium">{proc.stats.throughput}</span>
                    <span className="text-xs text-muted-foreground">/min</span>
                  </div>
                </td>
                <td className="p-4 text-right">
                  <div className="flex items-center justify-end gap-1">
                    <span className="font-mono font-medium">
                      {proc.status === "active" ? proc.stats.avgLatency : "-"}
                    </span>
                    {proc.status === "active" && (
                      <span className="text-xs text-muted-foreground">ms</span>
                    )}
                  </div>
                </td>
                <td className="p-4">
                  <Activity className="h-4 w-4 text-muted-foreground" />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>

        {/* Empty State */}
        {filteredProcessors.length === 0 && (
          <div className="text-center py-12 border">
            <p className="text-muted-foreground">No processors found</p>
          </div>
        )}
      </div>
    </div>
  )
}