"use client"

import { useState, useEffect } from "react"
import { useParams, useRouter } from "next/navigation"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { 
  Inbox, Play, Pause, Settings, Clock, Plus,
  Activity, CheckCircle, AlertCircle, Filter,
  Code2, Database, RotateCcw, ArrowLeft,
  Edit2, Trash2, MoreVertical, FileCode,
  Save, X, Copy, Info
} from "lucide-react"
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
import { Separator } from "@/components/ui/separator"
import { useToast } from "@/hooks/use-toast"

export default function SubscriptionDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { toast } = useToast()
  
  const subscriptionName = params?.subscription as string
  
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState("overview")
  
  // Edit mode states
  const [editMode, setEditMode] = useState(false)
  const [editedSubscription, setEditedSubscription] = useState<any>(null)
  
  // Mock data - replace with actual API calls
  const [subscription, setSubscription] = useState({
    id: "sub-1",
    name: subscriptionName,
    description: "Process incoming orders and update inventory",
    status: "active",
    created: "2024-01-01T00:00:00Z",
    modified: "2024-01-15T10:30:00Z",
    filter: { type: "order", status: "pending" },
    function: "processOrder",
    outputCollection: "processed_orders",
    // Queue position data
    pending: 23,
    currentPosition: 5,
    processing: 2,
    throughput: 45,
    // Processing stats
    stats: {
      totalProcessed: 5234,
      processedToday: 234,
      failedToday: 2,
      avgProcessTime: 450,
      successRate: 99.1
    },
    config: {
      maxConcurrency: 10,
      timeout: 30000,
      maxRetries: 3,
      backoffMultiplier: 2,
      initialDelay: 1000
    }
  })


  useEffect(() => {
    setEditedSubscription(subscription)
  }, [subscription])

  const handleSaveSubscription = async () => {
    try {
      // API call would go here
      setSubscription(editedSubscription)
      setEditMode(false)
      toast({
        title: "Success",
        description: "Subscription updated successfully",
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update subscription",
        variant: "destructive",
      })
    }
  }

  const handleToggleStatus = async () => {
    try {
      const newStatus = subscription.status === "active" ? "paused" : "active"
      setSubscription({ ...subscription, status: newStatus })
      toast({
        title: "Success",
        description: `Subscription ${newStatus === "active" ? "resumed" : "paused"} successfully`,
      })
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update subscription status",
        variant: "destructive",
      })
    }
  }


  return (
    <div className="flex flex-col h-full">
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
                <BreadcrumbLink href="/event-store">Event Store</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>{subscriptionName}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          
          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2">
                <Inbox className="h-5 w-5 text-muted-foreground" />
                <h1 className="text-2xl font-bold">{subscriptionName}</h1>
              </div>
              <p className="text-sm text-muted-foreground mt-1">
                {subscription.description || "No description provided"}
              </p>
            </div>
            <div className="flex items-center gap-2">
              {editMode ? (
                <>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setEditedSubscription(subscription)
                      setEditMode(false)
                    }}
                  >
                    <X className="h-4 w-4 mr-2" />
                    Cancel
                  </Button>
                  <Button onClick={handleSaveSubscription}>
                    <Save className="h-4 w-4 mr-2" />
                    Save Changes
                  </Button>
                </>
              ) : (
                <>
                  <Badge variant={subscription.status === "active" ? "default" : "secondary"}>
                    {subscription.status}
                  </Badge>
                  <Button
                    variant="outline"
                    onClick={() => setEditMode(true)}
                  >
                    <Edit2 className="h-4 w-4 mr-2" />
                    Edit
                  </Button>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="outline" size="icon">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={handleToggleStatus}>
                        {subscription.status === "active" ? (
                          <>
                            <Pause className="h-4 w-4 mr-2" />
                            Pause Subscription
                          </>
                        ) : (
                          <>
                            <Play className="h-4 w-4 mr-2" />
                            Resume Subscription
                          </>
                        )}
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <RotateCcw className="h-4 w-4 mr-2" />
                        Reset Stats
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <FileCode className="h-4 w-4 mr-2" />
                        View Function
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem className="text-red-600">
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete Subscription
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto">
        <div className="container mx-auto px-6 py-6 max-w-7xl">
        {/* Queue Position Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
          <Card className="rounded-none shadow-none">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Pending Events</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-yellow-600">{subscription.pendingEvents}</div>
              <p className="text-xs text-muted-foreground mt-1">Events waiting to be processed</p>
            </CardContent>
          </Card>
          <Card className="rounded-none shadow-none">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Current Position</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">#{subscription.currentPosition}</div>
              <p className="text-xs text-muted-foreground mt-1">Next event position in queue</p>
            </CardContent>
          </Card>
          <Card className="rounded-none shadow-none">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Processing</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-blue-600">{subscription.processingEvents}</div>
              <p className="text-xs text-muted-foreground mt-1">Currently being processed</p>
            </CardContent>
          </Card>
          <Card className="rounded-none shadow-none">
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Throughput</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{subscription.throughput}/min</div>
              <p className="text-xs text-muted-foreground mt-1">Average processing rate</p>
            </CardContent>
          </Card>
        </div>

        {/* Tabs */}
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="mb-4">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="configuration">Configuration</TabsTrigger>
            <TabsTrigger value="function">Function</TabsTrigger>
          </TabsList>

          {/* Overview Tab */}
          <TabsContent value="overview">
            <div className="grid gap-4">
              {/* Status Info */}
              <Card className="rounded-none shadow-none">
                <CardHeader>
                  <CardTitle>Processor Status</CardTitle>
                  <CardDescription>
                    Current state and position in the event store
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-6 md:grid-cols-2">
                    <div className="space-y-4">
                      <div>
                        <Label className="text-sm font-medium">Status</Label>
                        <div className="mt-2">
                          <Badge 
                            variant={subscription.status === "active" ? "default" : "secondary"}
                            className="text-sm"
                          >
                            {subscription.status === "active" ? "Active" : "Paused"}
                          </Badge>
                        </div>
                      </div>
                      <div>
                        <Label className="text-sm font-medium">Event Filter</Label>
                        <pre className="text-sm bg-muted p-3 mt-2">
{JSON.stringify(subscription.filter, null, 2)}
                        </pre>
                      </div>
                    </div>
                    <div className="space-y-4">
                      <div>
                        <Label className="text-sm font-medium">Queue Position</Label>
                        <p className="text-2xl font-bold mt-2">Position #{subscription.currentPosition}</p>
                        <p className="text-sm text-muted-foreground">Next event to process</p>
                      </div>
                      <div>
                        <Label className="text-sm font-medium">Pending Events</Label>
                        <p className="text-2xl font-bold text-yellow-600 mt-2">{subscription.pendingEvents}</p>
                        <p className="text-sm text-muted-foreground">Events matching filter</p>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Processing Stats */}
              <Card className="rounded-none shadow-none">
                <CardHeader>
                  <CardTitle>Processing Statistics</CardTitle>
                  <CardDescription>
                    Historical performance metrics
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-4 md:grid-cols-3">
                    <div>
                      <Label className="text-sm font-medium text-muted-foreground">Total Processed</Label>
                      <p className="text-2xl font-bold mt-1">{subscription.stats.totalProcessed.toLocaleString()}</p>
                    </div>
                    <div>
                      <Label className="text-sm font-medium text-muted-foreground">Processed Today</Label>
                      <p className="text-2xl font-bold text-green-600 mt-1">{subscription.stats.processedToday}</p>
                    </div>
                    <div>
                      <Label className="text-sm font-medium text-muted-foreground">Failed Today</Label>
                      <p className="text-2xl font-bold text-red-600 mt-1">{subscription.stats.failedToday}</p>
                    </div>
                    <div>
                      <Label className="text-sm font-medium text-muted-foreground">Avg Process Time</Label>
                      <p className="text-2xl font-bold mt-1">{subscription.stats.avgProcessTime}ms</p>
                    </div>
                    <div>
                      <Label className="text-sm font-medium text-muted-foreground">Success Rate</Label>
                      <p className="text-2xl font-bold mt-1">{subscription.stats.successRate}%</p>
                    </div>
                    <div>
                      <Label className="text-sm font-medium text-muted-foreground">Throughput</Label>
                      <p className="text-2xl font-bold mt-1">{subscription.throughput}/min</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* Configuration Tab */}
          <TabsContent value="configuration">
            <div className="grid gap-4 md:grid-cols-2">
              <Card className="rounded-none shadow-none">
                <CardHeader>
                  <CardTitle>Event Filter</CardTitle>
                  <CardDescription>
                    JSON filter that determines which events this processor consumes
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <pre className="text-sm bg-muted p-4 overflow-x-auto">
{JSON.stringify(subscription.filter, null, 2)}
                  </pre>
                </CardContent>
              </Card>

              <Card className="rounded-none shadow-none">
                <CardHeader>
                  <CardTitle>Processing Configuration</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label className="text-sm font-medium">Function</Label>
                    <p className="text-sm font-mono mt-1">{subscription.function}()</p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Output Collection</Label>
                    <p className="text-sm font-mono mt-1">{subscription.outputCollection}</p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Max Concurrency</Label>
                    <p className="text-sm text-muted-foreground mt-1">{subscription.config.maxConcurrency} workers</p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Timeout</Label>
                    <p className="text-sm text-muted-foreground mt-1">{subscription.config.timeout}ms</p>
                  </div>
                </CardContent>
              </Card>

              <Card className="rounded-none shadow-none">
                <CardHeader>
                  <CardTitle>Retry Policy</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label className="text-sm font-medium">Max Retries</Label>
                    <p className="text-sm text-muted-foreground mt-1">{subscription.config.maxRetries}</p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Backoff Multiplier</Label>
                    <p className="text-sm text-muted-foreground mt-1">{subscription.config.backoffMultiplier}x</p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Initial Delay</Label>
                    <p className="text-sm text-muted-foreground mt-1">{subscription.config.initialDelay}ms</p>
                  </div>
                </CardContent>
              </Card>

              <Card className="rounded-none shadow-none">
                <CardHeader>
                  <CardTitle>Processor Information</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label className="text-sm font-medium">Created</Label>
                    <p className="text-sm text-muted-foreground mt-1">
                      {new Date(subscription.created).toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Last Modified</Label>
                    <p className="text-sm text-muted-foreground mt-1">
                      {new Date(subscription.modified).toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <Label className="text-sm font-medium">Status</Label>
                    <div className="mt-1">
                      <Badge variant={subscription.status === "active" ? "default" : "secondary"}>
                        {subscription.status}
                      </Badge>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* Function Tab */}
          <TabsContent value="function">
            <Card className="rounded-none shadow-none">
              <CardHeader>
                <CardTitle>Processing Function</CardTitle>
                <CardDescription>
                  Function code that processes events for this processor
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <FileCode className="h-5 w-5 text-muted-foreground" />
                      <span className="font-mono text-sm">{subscription.function}.js</span>
                    </div>
                    <Button variant="outline" size="sm">
                      <Edit2 className="h-4 w-4 mr-2" />
                      Edit Function
                    </Button>
                  </div>
                  <pre className="text-sm bg-muted p-4 overflow-x-auto">
{`export async function ${subscription.function}(event) {
  // Extract order data from event
  const { orderId, customerId, amount, items } = event.data;
  
  try {
    // Validate order
    if (!orderId || !customerId || !amount) {
      throw new Error('Invalid order data');
    }
    
    // Process payment
    const paymentResult = await processPayment({
      customerId,
      amount,
      orderId
    });
    
    // Update inventory
    await updateInventory(items);
    
    // Write to output collection
    return {
      orderId,
      customerId,
      amount,
      items,
      paymentId: paymentResult.id,
      processedAt: new Date().toISOString(),
      status: 'completed'
    };
  } catch (error) {
    console.error('Failed to process order:', error);
    throw error;
  }
}`}
                  </pre>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
        </div>
      </div>
    </div>
  )
}