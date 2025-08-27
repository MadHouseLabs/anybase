import { useState, useEffect } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { collectionsApi, usersApi } from "@/lib/api"
import { Database, FileJson, Shield, Users, Activity, Server, TrendingUp, Eye } from "lucide-react"

export function DashboardPage() {
  const [stats, setStats] = useState({
    collections: 0,
    documents: 0,
    users: 0,
    views: 0,
  })

  useEffect(() => {
    loadStats()
  }, [])

  const loadStats = async () => {
    try {
      const [collectionsRes, usersRes] = await Promise.all([
        collectionsApi.list(),
        usersApi.list(),
      ])
      
      setStats({
        collections: collectionsRes.collections?.length || 0,
        documents: collectionsRes.collections?.reduce((sum: number, col: any) => sum + (col.document_count || 0), 0) || 0,
        users: usersRes.users?.length || 0,
        views: collectionsRes.views?.length || 0,
      })
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
    },
    {
      title: "Documents",
      value: stats.documents,
      description: "Total documents stored",
      icon: FileJson,
      color: "text-green-600",
    },
    {
      title: "Users",
      value: stats.users,
      description: "Registered users",
      icon: Users,
      color: "text-purple-600",
    },
    {
      title: "Views",
      value: stats.views,
      description: "Data views configured",
      icon: Eye,
      color: "text-orange-600",
    },
  ]

  return (
    <div className="container mx-auto py-6 space-y-8">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">Welcome to AnyBase Database Management System</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat, index) => (
          <Card key={index}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {stat.title}
              </CardTitle>
              <stat.icon className={`h-4 w-4 ${stat.color}`} />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
              <p className="text-xs text-muted-foreground">
                {stat.description}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              System Status
            </CardTitle>
            <CardDescription>Current system health and performance</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Server className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm">Database Connection</span>
                </div>
                <span className="text-sm font-medium text-green-600">Connected</span>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <TrendingUp className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm">API Response Time</span>
                </div>
                <span className="text-sm font-medium">~45ms</span>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Eye className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm">Active Sessions</span>
                </div>
                <span className="text-sm font-medium">1</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Common tasks and operations</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-2">
              <button className="flex items-center gap-2 p-3 rounded-lg border hover:bg-muted transition-colors">
                <Database className="h-4 w-4" />
                <span className="text-sm">New Collection</span>
              </button>
              <button className="flex items-center gap-2 p-3 rounded-lg border hover:bg-muted transition-colors">
                <FileJson className="h-4 w-4" />
                <span className="text-sm">Add Document</span>
              </button>
              <button className="flex items-center gap-2 p-3 rounded-lg border hover:bg-muted transition-colors">
                <Shield className="h-4 w-4" />
                <span className="text-sm">Manage Roles</span>
              </button>
              <button className="flex items-center gap-2 p-3 rounded-lg border hover:bg-muted transition-colors">
                <Users className="h-4 w-4" />
                <span className="text-sm">View Users</span>
              </button>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Getting Started</CardTitle>
          <CardDescription>Learn how to use AnyBase effectively</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="p-4 border rounded-lg">
              <h4 className="font-medium mb-2">1. Create Collections</h4>
              <p className="text-sm text-muted-foreground">
                Start by creating collections to organize your data. Each collection can have its own schema and permissions.
              </p>
            </div>
            <div className="p-4 border rounded-lg">
              <h4 className="font-medium mb-2">2. Add Documents</h4>
              <p className="text-sm text-muted-foreground">
                Insert documents into your collections. Documents are automatically versioned and can be soft-deleted.
              </p>
            </div>
            <div className="p-4 border rounded-lg">
              <h4 className="font-medium mb-2">3. Configure Access</h4>
              <p className="text-sm text-muted-foreground">
                Set up roles and permissions to control who can access your data and what operations they can perform.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}