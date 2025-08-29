"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { 
  Inbox, Search, MoreVertical, Copy, ArrowLeft, Plus
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
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useToast } from "@/hooks/use-toast"

export default function EventsPage() {
  const { toast } = useToast()
  const router = useRouter()
  const [searchQuery, setSearchQuery] = useState("")
  const [filterType, setFilterType] = useState("all")

  // Mock data - replace with actual API calls
  const events = [
    {
      id: "msg-001",
      type: "order",
      position: 1,
      data: { orderId: "ORD-1234", amount: 1500, customerId: "CUST-5678" },
      created: "2024-01-15T10:30:00Z",
      ttl: 3600000 // 1 hour
    },
    {
      id: "msg-002",
      type: "notification",
      position: 2,
      data: { userId: "USR-5678", subject: "Welcome", template: "welcome_email" },
      created: "2024-01-15T10:30:05Z",
      ttl: 3600000
    },
    {
      id: "msg-003",
      type: "payment",
      position: 3,
      data: { transactionId: "TXN-9012", amount: 2000 },
      created: "2024-01-15T10:29:50Z",
      ttl: 7200000 // 2 hours
    },
    {
      id: "msg-004",
      type: "order",
      position: 4,
      data: { orderId: "ORD-1235", amount: 750, customerId: "CUST-5679" },
      created: "2024-01-15T10:30:08Z",
      ttl: 3600000
    },
    {
      id: "msg-005",
      type: "image",
      position: 5,
      data: { fileId: "FILE-3456", size: "2.5MB", operation: "resize" },
      created: "2024-01-15T10:29:30Z",
      ttl: 1800000 // 30 minutes
    },
    {
      id: "msg-006",
      type: "analytics",
      position: 6,
      data: { eventType: "page_view", userId: "USR-1234" },
      created: "2024-01-15T10:28:00Z",
      ttl: 86400000 // 24 hours
    },
    {
      id: "msg-007",
      type: "order",
      position: 7,
      data: { orderId: "ORD-1236", amount: 3200 },
      created: "2024-01-15T10:31:00Z",
      ttl: 3600000
    },
    {
      id: "msg-008",
      type: "notification",
      position: 8,
      data: { userId: "USR-5677", subject: "Order Confirmation" },
      created: "2024-01-15T10:31:10Z",
      ttl: 3600000
    }
  ]

  const filteredEvents = events.filter(event => {
    if (filterType !== "all" && event.type !== filterType) return false
    if (!searchQuery) return true
    const searchLower = searchQuery.toLowerCase()
    return event.id.toLowerCase().includes(searchLower) ||
           event.type.toLowerCase().includes(searchLower) ||
           JSON.stringify(event.data).toLowerCase().includes(searchLower)
  })

  const handleDeleteEvent = (eventId: string) => {
    toast({
      title: "Event Deleted",
      description: `Event ${eventId} has been removed from the store`,
    })
  }

  return (
    <div className="container mx-auto py-8 max-w-7xl">
      {/* Breadcrumbs */}
      <Breadcrumb className="mb-6">
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
            <BreadcrumbPage>Events</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-semibold">Events</h1>
          <p className="text-sm text-muted-foreground mt-1">
            All events currently in the store
          </p>
        </div>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          Add Event
        </Button>
      </div>

      {/* Events Table */}
      <Card className="rounded-none shadow-none">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Stored Events</CardTitle>
              <CardDescription>
                All events in the store waiting for processing
              </CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <select 
                className="h-9 border border-input bg-background px-3 py-1 text-sm"
                value={filterType}
                onChange={(e) => setFilterType(e.target.value)}
              >
                <option value="all">All Types</option>
                <option value="order">Order</option>
                <option value="notification">Notification</option>
                <option value="payment">Payment</option>
                <option value="image">Image</option>
                <option value="analytics">Analytics</option>
              </select>
              <div className="relative w-64">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                <Input 
                  placeholder="Search events..." 
                  className="pl-10"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                />
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Event ID</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Position</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>TTL</TableHead>
                <TableHead>Data</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredEvents.map((event) => (
                <TableRow key={event.id}>
                  <TableCell className="font-mono text-sm">{event.id}</TableCell>
                  <TableCell>
                    <Badge variant="outline" className="text-xs">
                      {event.type}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm">
                    <span className="font-mono">#{event.position}</span>
                  </TableCell>
                  <TableCell className="text-sm">
                    {new Date(event.created).toLocaleString()}
                  </TableCell>
                  <TableCell className="text-sm">
                    {event.ttl ? `${Math.round(event.ttl / 60000)}m` : "-"}
                  </TableCell>
                  <TableCell>
                    <code className="text-xs bg-muted px-2 py-1">
                      {JSON.stringify(event.data).substring(0, 40)}...
                    </code>
                  </TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <MoreVertical className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => navigator.clipboard.writeText(JSON.stringify(event.data, null, 2))}>
                          <Copy className="h-4 w-4 mr-2" />
                          Copy Data
                        </DropdownMenuItem>
                        <DropdownMenuItem 
                          className="text-red-600"
                          onClick={() => handleDeleteEvent(event.id)}
                        >
                          Delete
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}