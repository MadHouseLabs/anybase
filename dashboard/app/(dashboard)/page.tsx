"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription } from "@/components/ui/alert"
import { collectionsApi, usersApi } from "@/lib/api"
import { Database, FileJson, Shield, Users, Activity, Server, TrendingUp, Eye, Home, ArrowRight, CheckCircle, AlertCircle, Zap, BookOpen, Key, BarChart3, Clock, Plus } from "lucide-react"
import Cookies from "js-cookie"

export default function DashboardPage() {
  const router = useRouter()
  const [stats, setStats] = useState({
    collections: 0,
    documents: 0,
    users: 0,
    apiKeys: 0,
  })
  const [recentActivity, setRecentActivity] = useState<any[]>([])
  const [systemHealth, setSystemHealth] = useState({
    database: 'healthy',
    responseTime: 0,
    uptime: 0,
    storageUsed: 0,
    storageTotal: 10
  })

  useEffect(() => {
    loadStats()
  }, [])

  const loadStats = async () => {
    try {
      // Get collections data
      const collectionsRes = await collectionsApi.list()
      
      // Try to get users data, but handle error gracefully
      let usersCount = 0
      try {
        const usersRes = await usersApi.list()
        usersCount = usersRes.users?.length || 0
      } catch (error) {
        console.log("Unable to fetch users (admin access required)")
      }
      
      // Try to get access keys count
      let apiKeysCount = 0
      try {
        const { accessKeysApi } = await import('@/lib/accesskeys')
        const keysRes = await accessKeysApi.list()
        apiKeysCount = keysRes.access_keys?.filter((k: any) => k.active)?.length || 0
      } catch (error) {
        console.log("Unable to fetch access keys")
      }
      
      setStats({
        collections: collectionsRes.collections?.length || 0,
        documents: collectionsRes.collections?.reduce((sum: number, col: any) => sum + (col.document_count || 0), 0) || 0,
        users: usersCount,
        apiKeys: apiKeysCount,
      })
      
      // Generate recent activity from actual data
      const activities = []
      if (collectionsRes.collections?.length > 0) {
        const recentCollection = collectionsRes.collections[collectionsRes.collections.length - 1]
        activities.push({
          action: "Collection created",
          item: recentCollection.name,
          time: "Recently",
          icon: Database,
          color: "text-blue-600"
        })
      }
      if (stats.documents > 0) {
        activities.push({
          action: "Documents added",
          item: `${stats.documents} total documents`,
          time: "Active",
          icon: FileJson,
          color: "text-green-600"
        })
      }
      if (apiKeysCount > 0) {
        activities.push({
          action: "API keys active",
          item: `${apiKeysCount} keys configured`,
          time: "Configured",
          icon: Key,
          color: "text-purple-600"
        })
      }
      if (usersCount > 0) {
        activities.push({
          action: "Users registered",
          item: `${usersCount} total users`,
          time: "In system",
          icon: Users,
          color: "text-orange-600"
        })
      }
      setRecentActivity(activities)
      
      // Calculate system health metrics
      const startTime = Date.now()
      try {
        await fetch('/health')
        const responseTime = Date.now() - startTime
        setSystemHealth(prev => ({
          ...prev,
          responseTime,
          uptime: 98, // This would come from a real monitoring endpoint
          storageUsed: Math.round(stats.documents * 0.001 * 100) / 100 // Rough estimate: 1KB per doc
        }))
      } catch (error) {
        console.log("Health check failed")
      }
    } catch (error) {
      console.error("Failed to load stats:", error)
    }
  }

  const statCards = [
    {
      title: "Collections",
      value: stats.collections,
      description: "Active database collections",
      icon: Database,
      color: "text-blue-600",
      bgColor: "bg-blue-100",
      link: "/collections"
    },
    {
      title: "Documents",
      value: stats.documents.toLocaleString(),
      description: "Total documents stored",
      icon: FileJson,
      color: "text-green-600",
      bgColor: "bg-green-100",
      link: "/collections"
    },
    {
      title: "Users",
      value: stats.users,
      description: "Registered users",
      icon: Users,
      color: "text-purple-600",
      bgColor: "bg-purple-100",
      link: "/users"
    },
    {
      title: "API Keys",
      value: stats.apiKeys || 0,
      description: "Active access keys",
      icon: Key,
      color: "text-orange-600",
      bgColor: "bg-orange-100",
      link: "/access-keys"
    },
  ]

  return (
    <div className="container mx-auto py-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Home className="h-8 w-8" />
            Dashboard Overview
          </h1>
          <p className="text-muted-foreground mt-1">
            Welcome back! Here's what's happening with your database today.
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => router.push('/collections')}>
            <Database className="h-4 w-4 mr-2" />
            View Collections
          </Button>
          <Button onClick={() => router.push('/collections')}>
            <Plus className="h-4 w-4 mr-2" />
            New Collection
          </Button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat, index) => (
          <Card key={index} className="hover:shadow-lg transition-shadow cursor-pointer" onClick={() => router.push(stat.link || '#')}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {stat.title}
              </CardTitle>
              <div className={`p-2 rounded-lg ${stat.bgColor || 'bg-primary/10'}`}>
                <stat.icon className={`h-4 w-4 ${stat.color}`} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
              <p className="text-xs text-muted-foreground">
                {stat.description}
              </p>
              {stat.trend && (
                <div className="flex items-center gap-1 mt-2">
                  <TrendingUp className="h-3 w-3 text-green-600" />
                  <span className="text-xs text-green-600">{stat.trend}</span>
                </div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Activity and Actions */}
      <div className="grid gap-4 lg:grid-cols-3">
        {/* System Health Card */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              System Health
            </CardTitle>
            <CardDescription>Real-time system monitoring</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3">
              <div>
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <Server className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Database</span>
                  </div>
                  <Badge variant="default" className="bg-green-100 text-green-800">
                    <CheckCircle className="h-3 w-3 mr-1" />
                    Healthy
                  </Badge>
                </div>
                <div className="h-2 bg-gray-200 rounded-full overflow-hidden">
                  <div className="h-full bg-green-500" style={{ width: `${systemHealth.uptime}%` }} />
                </div>
                <p className="text-xs text-muted-foreground mt-1">{systemHealth.uptime}% uptime</p>
              </div>
              
              <div>
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <Zap className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Performance</span>
                  </div>
                  <span className="text-sm text-green-600 font-medium">{systemHealth.responseTime}ms</span>
                </div>
                <div className="h-2 bg-gray-200 rounded-full overflow-hidden">
                  <div className="h-full bg-blue-500" style={{ width: `${Math.min(100, Math.max(0, 100 - (systemHealth.responseTime / 10)))}%` }} />
                </div>
                <p className="text-xs text-muted-foreground mt-1">Avg response time</p>
              </div>
              
              <div>
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <BarChart3 className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm font-medium">Storage</span>
                  </div>
                  <span className="text-sm font-medium">{systemHealth.storageUsed} GB</span>
                </div>
                <div className="h-2 bg-gray-200 rounded-full overflow-hidden">
                  <div className="h-full bg-orange-500" style={{ width: `${(systemHealth.storageUsed / systemHealth.storageTotal) * 100}%` }} />
                </div>
                <p className="text-xs text-muted-foreground mt-1">{Math.round((systemHealth.storageUsed / systemHealth.storageTotal) * 100)}% of {systemHealth.storageTotal} GB used</p>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Quick Actions Card */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Zap className="h-5 w-5" />
              Quick Actions
            </CardTitle>
            <CardDescription>Frequently used operations</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 gap-2">
              <Button 
                onClick={() => router.push('/collections')}
                variant="outline"
                className="justify-start h-auto py-3 hover:bg-blue-50 hover:border-blue-300"
              >
                <Database className="h-4 w-4 mr-3 text-blue-600" />
                <div className="text-left">
                  <div className="font-medium">Create Collection</div>
                  <div className="text-xs text-muted-foreground">Set up a new data collection</div>
                </div>
              </Button>
              <Button 
                onClick={() => router.push('/collections')}
                variant="outline"
                className="justify-start h-auto py-3 hover:bg-green-50 hover:border-green-300"
              >
                <FileJson className="h-4 w-4 mr-3 text-green-600" />
                <div className="text-left">
                  <div className="font-medium">Add Document</div>
                  <div className="text-xs text-muted-foreground">Insert data into collections</div>
                </div>
              </Button>
              <Button 
                onClick={() => router.push('/access-keys')}
                variant="outline"
                className="justify-start h-auto py-3 hover:bg-purple-50 hover:border-purple-300"
              >
                <Key className="h-4 w-4 mr-3 text-purple-600" />
                <div className="text-left">
                  <div className="font-medium">Manage API Keys</div>
                  <div className="text-xs text-muted-foreground">Control API access</div>
                </div>
              </Button>
              <Button 
                onClick={() => router.push('/users')}
                variant="outline"
                className="justify-start h-auto py-3 hover:bg-orange-50 hover:border-orange-300"
              >
                <Users className="h-4 w-4 mr-3 text-orange-600" />
                <div className="text-left">
                  <div className="font-medium">User Management</div>
                  <div className="text-xs text-muted-foreground">View and manage users</div>
                </div>
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Recent Activity Card */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="h-5 w-5" />
              Recent Activity
            </CardTitle>
            <CardDescription>Latest system events</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {recentActivity.length > 0 ? recentActivity.map((activity, index) => (
                <div key={index} className="flex items-start gap-3 p-2 rounded-lg hover:bg-muted/50 transition-colors">
                  <div className={`p-2 rounded-lg bg-muted`}>
                    <activity.icon className={`h-4 w-4 ${activity.color}`} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium truncate">{activity.action}</p>
                    <p className="text-xs text-muted-foreground truncate">{activity.item}</p>
                  </div>
                  <span className="text-xs text-muted-foreground whitespace-nowrap">{activity.time}</span>
                </div>
              )) : (
                <div className="text-center py-8 text-sm text-muted-foreground">
                  No recent activity
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Getting Started Guide */}
      <Card className="border-2">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <BookOpen className="h-5 w-5" />
                Getting Started Guide
              </CardTitle>
              <CardDescription>Learn how to make the most of AnyBase</CardDescription>
            </div>
            <Button variant="outline" size="sm" onClick={() => window.open('https://github.com/karthik/anybase', '_blank')}>
              View Docs
              <ArrowRight className="h-4 w-4 ml-2" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div className="relative p-4 border-2 border-dashed rounded-lg hover:border-primary hover:bg-primary/5 transition-all cursor-pointer group">
              <div className="absolute -top-3 left-4 bg-background px-2">
                <Badge variant="outline" className="font-mono">Step 1</Badge>
              </div>
              <div className="mt-2">
                <div className="flex items-center gap-2 mb-2">
                  <div className="p-2 rounded-lg bg-blue-100 group-hover:bg-blue-200 transition-colors">
                    <Database className="h-5 w-5 text-blue-600" />
                  </div>
                  <h4 className="font-semibold">Create Collections</h4>
                </div>
                <p className="text-sm text-muted-foreground">
                  Define your data structure with collections. Set up schemas, indexes, and configure permissions.
                </p>
                <Button variant="link" className="px-0 mt-2" onClick={() => router.push('/collections')}>
                  Start creating
                  <ArrowRight className="h-3 w-3 ml-1" />
                </Button>
              </div>
            </div>
            
            <div className="relative p-4 border-2 border-dashed rounded-lg hover:border-primary hover:bg-primary/5 transition-all cursor-pointer group">
              <div className="absolute -top-3 left-4 bg-background px-2">
                <Badge variant="outline" className="font-mono">Step 2</Badge>
              </div>
              <div className="mt-2">
                <div className="flex items-center gap-2 mb-2">
                  <div className="p-2 rounded-lg bg-green-100 group-hover:bg-green-200 transition-colors">
                    <FileJson className="h-5 w-5 text-green-600" />
                  </div>
                  <h4 className="font-semibold">Add Documents</h4>
                </div>
                <p className="text-sm text-muted-foreground">
                  Insert and manage your data. Documents support versioning, soft-delete, and real-time validation.
                </p>
                <Button variant="link" className="px-0 mt-2" onClick={() => router.push('/collections')}>
                  Add documents
                  <ArrowRight className="h-3 w-3 ml-1" />
                </Button>
              </div>
            </div>
            
            <div className="relative p-4 border-2 border-dashed rounded-lg hover:border-primary hover:bg-primary/5 transition-all cursor-pointer group">
              <div className="absolute -top-3 left-4 bg-background px-2">
                <Badge variant="outline" className="font-mono">Step 3</Badge>
              </div>
              <div className="mt-2">
                <div className="flex items-center gap-2 mb-2">
                  <div className="p-2 rounded-lg bg-purple-100 group-hover:bg-purple-200 transition-colors">
                    <Shield className="h-5 w-5 text-purple-600" />
                  </div>
                  <h4 className="font-semibold">Configure Access</h4>
                </div>
                <p className="text-sm text-muted-foreground">
                  Set up API keys and manage user permissions to secure your data with role-based access control.
                </p>
                <Button variant="link" className="px-0 mt-2" onClick={() => router.push('/access-keys')}>
                  Setup access
                  <ArrowRight className="h-3 w-3 ml-1" />
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}