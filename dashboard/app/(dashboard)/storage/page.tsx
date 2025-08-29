import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { 
  HardDrive, Plus, Upload, Download, Folder, File,
  Image, FileText, Film, Music, Archive, Trash2,
  Share2, Lock, Unlock, MoreVertical, Search
} from "lucide-react";
import { Input } from "@/components/ui/input";

export default async function StoragePage() {
  // Mock data - replace with actual API calls
  const buckets = [
    {
      id: "1",
      name: "user-uploads",
      description: "User uploaded files",
      visibility: "private",
      files: 1234,
      size: 5.2, // GB
      lastModified: "2024-01-15T10:30:00Z",
    },
    {
      id: "2",
      name: "public-assets",
      description: "Public CDN assets",
      visibility: "public",
      files: 456,
      size: 2.8,
      lastModified: "2024-01-14T15:20:00Z",
    },
    {
      id: "3",
      name: "backups",
      description: "System backups",
      visibility: "private",
      files: 89,
      size: 12.5,
      lastModified: "2024-01-10T09:15:00Z",
    },
  ];

  const recentFiles = [
    { name: "report-2024.pdf", type: "pdf", size: "2.4 MB", bucket: "user-uploads" },
    { name: "profile-image.jpg", type: "image", size: "450 KB", bucket: "public-assets" },
    { name: "backup-2024-01-15.tar.gz", type: "archive", size: "1.2 GB", bucket: "backups" },
    { name: "presentation.pptx", type: "document", size: "8.9 MB", bucket: "user-uploads" },
    { name: "video-demo.mp4", type: "video", size: "125 MB", bucket: "public-assets" },
  ];

  const stats = {
    totalBuckets: buckets.length,
    totalFiles: buckets.reduce((sum, b) => sum + b.files, 0),
    totalStorage: buckets.reduce((sum, b) => sum + b.size, 0),
    storageLimit: 50, // GB
  };

  const getFileIcon = (type: string) => {
    switch (type) {
      case "image": return <Image className="h-4 w-4" />;
      case "pdf": return <FileText className="h-4 w-4" />;
      case "video": return <Film className="h-4 w-4" />;
      case "audio": return <Music className="h-4 w-4" />;
      case "archive": return <Archive className="h-4 w-4" />;
      default: return <File className="h-4 w-4" />;
    }
  };

  return (
    <div className="container mx-auto py-8 max-w-7xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <HardDrive className="h-8 w-8" />
            Storage
          </h1>
          <p className="text-muted-foreground mt-1">
            Manage files and object storage
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline">
            <Upload className="h-4 w-4 mr-2" />
            Upload Files
          </Button>
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            Create Bucket
          </Button>
        </div>
      </div>

      {/* Storage Overview */}
      <Card>
        <CardHeader>
          <CardTitle>Storage Usage</CardTitle>
          <CardDescription>
            {stats.totalStorage.toFixed(1)} GB of {stats.storageLimit} GB used
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Progress value={(stats.totalStorage / stats.storageLimit) * 100} className="h-3" />
          <div className="grid grid-cols-4 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground">Total Buckets</p>
              <p className="text-2xl font-bold">{stats.totalBuckets}</p>
            </div>
            <div>
              <p className="text-muted-foreground">Total Files</p>
              <p className="text-2xl font-bold">{stats.totalFiles.toLocaleString()}</p>
            </div>
            <div>
              <p className="text-muted-foreground">Storage Used</p>
              <p className="text-2xl font-bold">{stats.totalStorage.toFixed(1)} GB</p>
            </div>
            <div>
              <p className="text-muted-foreground">Available</p>
              <p className="text-2xl font-bold">{(stats.storageLimit - stats.totalStorage).toFixed(1)} GB</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Buckets Grid */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">Storage Buckets</h2>
          <div className="relative w-64">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <Input placeholder="Search buckets..." className="pl-10" />
          </div>
        </div>
        <div className="grid gap-4 md:grid-cols-3">
          {buckets.map((bucket) => (
            <Card key={bucket.id} className="hover:shadow-md transition-shadow">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <Folder className="h-5 w-5" />
                      <CardTitle className="text-lg">{bucket.name}</CardTitle>
                    </div>
                    <CardDescription>{bucket.description}</CardDescription>
                  </div>
                  <Badge variant={bucket.visibility === "public" ? "default" : "secondary"}>
                    {bucket.visibility === "public" ? (
                      <Unlock className="h-3 w-3 mr-1" />
                    ) : (
                      <Lock className="h-3 w-3 mr-1" />
                    )}
                    {bucket.visibility}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Files</span>
                    <span className="font-medium">{bucket.files.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">Size</span>
                    <span className="font-medium">{bucket.size} GB</span>
                  </div>
                  <div className="flex gap-2 pt-2">
                    <Button variant="outline" size="sm" className="flex-1">
                      <Upload className="h-4 w-4 mr-2" />
                      Upload
                    </Button>
                    <Button variant="outline" size="sm" className="flex-1">
                      Browse
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* Recent Files */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Files</CardTitle>
          <CardDescription>
            Recently uploaded or modified files
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {recentFiles.map((file, index) => (
              <div key={index} className="flex items-center justify-between p-3 rounded-lg hover:bg-muted/50 transition-colors">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-muted">
                    {getFileIcon(file.type)}
                  </div>
                  <div>
                    <p className="font-medium">{file.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {file.size} • {file.bucket}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button variant="ghost" size="sm">
                    <Download className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="sm">
                    <Share2 className="h-4 w-4" />
                  </Button>
                  <Button variant="ghost" size="sm">
                    <Trash2 className="h-4 w-4" />
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