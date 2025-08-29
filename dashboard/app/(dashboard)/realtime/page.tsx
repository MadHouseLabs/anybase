import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { 
  Radio, Plus, Send, Users, MessageCircle, 
  Activity, Zap, Globe, Lock, Signal, WifiOff,
  BarChart3, TrendingUp, MoreVertical, Settings
} from "lucide-react";
import { Input } from "@/components/ui/input";

export default async function RealtimePage() {
  // Mock data - replace with actual API calls
  const channels = [
    {
      id: "1",
      name: "global-events",
      description: "Global application events",
      type: "public",
      subscribers: 234,
      messages: 1543,
      messagesPerMin: 45,
      retention: "24h",
      lastActivity: "2024-01-15T10:30:00Z",
    },
    {
      id: "2",
      name: "user-notifications",
      description: "User-specific notifications",
      type: "private",
      subscribers: 89,
      messages: 3421,
      messagesPerMin: 12,
      retention: "7d",
      lastActivity: "2024-01-15T10:28:00Z",
    },
    {
      id: "3",
      name: "system-metrics",
      description: "Real-time system metrics",
      type: "public",
      subscribers: 45,
      messages: 8934,
      messagesPerMin: 120,
      retention: "1h",
      lastActivity: "2024-01-15T10:29:45Z",
    },
  ];

  const activeConnections = [
    { client: "web-app-1", channel: "global-events", duration: "2h 15m", messages: 234 },
    { client: "mobile-ios-23", channel: "user-notifications", duration: "45m", messages: 89 },
    { client: "admin-dashboard", channel: "system-metrics", duration: "1h 30m", messages: 456 },
    { client: "web-app-2", channel: "global-events", duration: "3h 10m", messages: 345 },
  ];

  const stats = {
    totalChannels: channels.length,
    activeConnections: activeConnections.length,
    totalMessages: channels.reduce((sum, c) => sum + c.messages, 0),
    totalSubscribers: channels.reduce((sum, c) => sum + c.subscribers, 0),
  };

  return (
    <div className="container mx-auto py-8 max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Radio className="h-8 w-8" />
            Realtime
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage PubSub channels and real-time connections
          </p>
        </div>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          Create Channel
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Channels</CardTitle>
            <MessageCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalChannels}</div>
            <p className="text-xs text-muted-foreground">
              PubSub channels
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Connections</CardTitle>
            <Signal className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.activeConnections}</div>
            <p className="text-xs text-muted-foreground">
              Connected clients
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Messages</CardTitle>
            <Send className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalMessages.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">
              Last 24 hours
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Subscribers</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalSubscribers}</div>
            <p className="text-xs text-muted-foreground">
              Total subscribers
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Channels Grid */}
      <div className="grid gap-4 md:grid-cols-3">
        {channels.map((channel) => (
          <Card key={channel.id} className="hover:shadow-md transition-shadow">
            <CardHeader>
              <div className="flex items-start justify-between">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <Radio className="h-5 w-5" />
                    <CardTitle className="text-lg">{channel.name}</CardTitle>
                  </div>
                  <CardDescription>{channel.description}</CardDescription>
                </div>
                <Badge variant={channel.type === "public" ? "default" : "secondary"}>
                  {channel.type === "public" ? (
                    <Globe className="h-3 w-3 mr-1" />
                  ) : (
                    <Lock className="h-3 w-3 mr-1" />
                  )}
                  {channel.type}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <p className="text-muted-foreground">Subscribers</p>
                    <p className="text-xl font-semibold">{channel.subscribers}</p>
                  </div>
                  <div>
                    <p className="text-muted-foreground">Messages</p>
                    <p className="text-xl font-semibold">{channel.messages.toLocaleString()}</p>
                  </div>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="flex items-center gap-1 text-muted-foreground">
                    <Activity className="h-3 w-3" />
                    {channel.messagesPerMin}/min
                  </span>
                  <span className="text-muted-foreground">
                    Retention: {channel.retention}
                  </span>
                </div>
                <div className="flex gap-2 pt-2">
                  <Button variant="outline" size="sm" className="flex-1">
                    <Send className="h-4 w-4 mr-2" />
                    Publish
                  </Button>
                  <Button variant="outline" size="sm" className="flex-1">
                    Subscribe
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Active Connections */}
      <Card>
        <CardHeader>
          <CardTitle>Active Connections</CardTitle>
          <CardDescription>
            Currently connected clients
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {activeConnections.map((connection, index) => (
              <div key={index} className="flex items-center justify-between p-3 rounded-lg hover:bg-muted/50 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-muted">
                    <Signal className="h-4 w-4 text-green-600" />
                  </div>
                  <div>
                    <p className="font-medium">{connection.client}</p>
                    <p className="text-sm text-muted-foreground">
                      Channel: {connection.channel} • Duration: {connection.duration}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <div className="text-right">
                    <p className="text-sm font-medium">{connection.messages}</p>
                    <p className="text-xs text-muted-foreground">messages</p>
                  </div>
                  <Button variant="ghost" size="sm">
                    <WifiOff className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}