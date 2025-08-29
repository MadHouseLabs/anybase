import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { 
  Clock, Plus, Play, Pause, Settings, Calendar,
  Activity, CheckCircle, XCircle, AlertCircle, Timer,
  RotateCw, History, MoreVertical, Edit, Trash2
} from "lucide-react";

export default async function CronPage() {
  // Mock data - replace with actual API calls
  const jobs = [
    {
      id: "1",
      name: "Database Backup",
      description: "Daily backup of all databases",
      schedule: "0 2 * * *", // 2 AM daily
      status: "active",
      lastRun: "2024-01-15T02:00:00Z",
      nextRun: "2024-01-16T02:00:00Z",
      successCount: 45,
      failureCount: 1,
      avgDuration: 320, // seconds
    },
    {
      id: "2",
      name: "Send Weekly Reports",
      description: "Email weekly analytics reports",
      schedule: "0 9 * * MON", // 9 AM every Monday
      status: "active",
      lastRun: "2024-01-15T09:00:00Z",
      nextRun: "2024-01-22T09:00:00Z",
      successCount: 12,
      failureCount: 0,
      avgDuration: 45,
    },
    {
      id: "3",
      name: "Clear Cache",
      description: "Clear temporary cache files",
      schedule: "*/30 * * * *", // Every 30 minutes
      status: "paused",
      lastRun: "2024-01-15T10:30:00Z",
      nextRun: null,
      successCount: 1420,
      failureCount: 3,
      avgDuration: 5,
    },
    {
      id: "4",
      name: "Sync External Data",
      description: "Sync data from external APIs",
      schedule: "0 */6 * * *", // Every 6 hours
      status: "active",
      lastRun: "2024-01-15T06:00:00Z",
      nextRun: "2024-01-15T12:00:00Z",
      successCount: 180,
      failureCount: 12,
      avgDuration: 120,
    },
  ];

  const recentRuns = [
    { job: "Database Backup", status: "success", startTime: "2024-01-15T02:00:00Z", duration: 318 },
    { job: "Sync External Data", status: "success", startTime: "2024-01-15T06:00:00Z", duration: 125 },
    { job: "Clear Cache", status: "failed", startTime: "2024-01-15T10:30:00Z", duration: 2, error: "Permission denied" },
    { job: "Send Weekly Reports", status: "success", startTime: "2024-01-15T09:00:00Z", duration: 43 },
    { job: "Database Backup", status: "success", startTime: "2024-01-14T02:00:00Z", duration: 322 },
  ];

  const stats = {
    totalJobs: jobs.length,
    activeJobs: jobs.filter(j => j.status === "active").length,
    totalRuns: jobs.reduce((sum, j) => sum + j.successCount + j.failureCount, 0),
    successRate: Math.round(
      (jobs.reduce((sum, j) => sum + j.successCount, 0) / 
       jobs.reduce((sum, j) => sum + j.successCount + j.failureCount, 0)) * 100
    ),
  };

  const formatCronSchedule = (schedule: string) => {
    // Simple cron to human-readable conversion
    if (schedule === "0 2 * * *") return "Daily at 2:00 AM";
    if (schedule === "0 9 * * MON") return "Weekly on Monday at 9:00 AM";
    if (schedule === "*/30 * * * *") return "Every 30 minutes";
    if (schedule === "0 */6 * * *") return "Every 6 hours";
    return schedule;
  };

  const formatDuration = (seconds: number) => {
    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}m ${remainingSeconds}s`;
  };

  return (
    <div className="container mx-auto py-8 max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Clock className="h-8 w-8" />
            Cron Jobs
          </h1>
          <p className="text-muted-foreground mt-1">
            Schedule and manage recurring tasks
          </p>
        </div>
        <Button>
          <Plus className="h-4 w-4 mr-2" />
          Create Job
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Jobs</CardTitle>
            <Timer className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalJobs}</div>
            <p className="text-xs text-muted-foreground">
              {stats.activeJobs} active
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Runs</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.totalRuns.toLocaleString()}</div>
            <p className="text-xs text-muted-foreground">
              All time
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.successRate}%</div>
            <p className="text-xs text-muted-foreground">
              Overall performance
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Next Run</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">5 min</div>
            <p className="text-xs text-muted-foreground">
              Clear Cache job
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Jobs List */}
      <Card>
        <CardHeader>
          <CardTitle>Scheduled Jobs</CardTitle>
          <CardDescription>
            Manage your automated tasks
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {jobs.map((job) => (
              <div key={job.id} className="border rounded-lg p-4 hover:bg-muted/50 transition-colors">
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <RotateCw className="h-5 w-5" />
                      <h3 className="font-semibold">{job.name}</h3>
                      <Badge variant={job.status === "active" ? "default" : "secondary"}>
                        {job.status}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">{job.description}</p>
                    <div className="flex items-center gap-4 text-sm">
                      <span className="flex items-center gap-1">
                        <Clock className="h-3 w-3" />
                        {formatCronSchedule(job.schedule)}
                      </span>
                      <span className="flex items-center gap-1 text-green-600">
                        <CheckCircle className="h-3 w-3" />
                        {job.successCount} successful
                      </span>
                      {job.failureCount > 0 && (
                        <span className="flex items-center gap-1 text-red-600">
                          <XCircle className="h-3 w-3" />
                          {job.failureCount} failed
                        </span>
                      )}
                      <span className="flex items-center gap-1">
                        <Timer className="h-3 w-3" />
                        Avg: {formatDuration(job.avgDuration)}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button variant="outline" size="sm">
                      <Play className="h-4 w-4 mr-2" />
                      Run Now
                    </Button>
                    {job.status === "active" ? (
                      <Button variant="outline" size="sm">
                        <Pause className="h-4 w-4" />
                      </Button>
                    ) : (
                      <Button variant="outline" size="sm">
                        <Play className="h-4 w-4" />
                      </Button>
                    )}
                    <Button variant="outline" size="sm">
                      <Settings className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Recent Runs */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Runs</CardTitle>
          <CardDescription>
            Latest job execution history
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {recentRuns.map((run, index) => (
              <div key={index} className="flex items-center justify-between p-3 rounded-lg hover:bg-muted/50">
                <div className="flex items-center gap-3">
                  {run.status === "success" ? (
                    <CheckCircle className="h-5 w-5 text-green-600" />
                  ) : (
                    <XCircle className="h-5 w-5 text-red-600" />
                  )}
                  <div>
                    <p className="font-medium">{run.job}</p>
                    <p className="text-sm text-muted-foreground">
                      {new Date(run.startTime).toLocaleString()} • {formatDuration(run.duration)}
                      {run.error && <span className="text-red-600"> • {run.error}</span>}
                    </p>
                  </div>
                </div>
                <Button variant="ghost" size="sm">
                  <History className="h-4 w-4 mr-2" />
                  View Logs
                </Button>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}