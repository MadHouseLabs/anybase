"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { 
  Plus, Search, Package, Upload, Download, Shield,
  CheckCircle, AlertCircle, FileArchive, HardDrive,
  MoreVertical, Trash2, ExternalLink, Copy, Info,
  Package2, Box, Archive
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

export default function BinariesPage() {
  const [searchQuery, setSearchQuery] = useState("")

  // Mock data - replace with actual API calls
  const binaries = [
    {
      id: "bin-1",
      name: "payment-processor-v2.3.1",
      displayName: "Payment Processor",
      description: "Compiled payment processing service with fraud detection",
      version: "v2.3.1",
      type: "executable",
      platform: "linux/amd64",
      size: "45.2 MB",
      checksum: "sha256:a1b2c3d4...",
      status: "verified",
      uploadedAt: "2024-01-15T10:30:00Z",
      functions: ["processPayment", "validateCard", "detectFraud"],
      dependencies: ["openssl", "libcurl"]
    },
    {
      id: "bin-2",
      name: "ml-inference-runtime",
      displayName: "ML Inference Runtime",
      description: "TensorFlow runtime for ML model inference",
      version: "v1.5.0",
      type: "library",
      platform: "linux/amd64",
      size: "128.5 MB",
      checksum: "sha256:e5f6g7h8...",
      status: "verified",
      uploadedAt: "2024-01-14T15:20:00Z",
      functions: ["predict", "preprocess", "postprocess"],
      dependencies: ["cuda-11.8", "cudnn-8.6"]
    },
    {
      id: "bin-3",
      name: "image-processor",
      displayName: "Image Processor",
      description: "High-performance image manipulation library",
      version: "v3.0.0",
      type: "library",
      platform: "linux/arm64",
      size: "22.8 MB",
      checksum: "sha256:i9j0k1l2...",
      status: "scanning",
      uploadedAt: "2024-01-15T11:00:00Z",
      functions: ["resize", "compress", "watermark", "convert"],
      dependencies: ["libvips", "imagemagick"]
    },
    {
      id: "bin-4",
      name: "data-transformer.wasm",
      displayName: "Data Transformer",
      description: "WebAssembly module for data transformation",
      version: "v1.2.0",
      type: "wasm",
      platform: "wasm32",
      size: "8.4 MB",
      checksum: "sha256:m3n4o5p6...",
      status: "verified",
      uploadedAt: "2024-01-13T09:15:00Z",
      functions: ["transform", "validate", "aggregate"],
      dependencies: []
    },
    {
      id: "bin-5",
      name: "notification-service.jar",
      displayName: "Notification Service",
      description: "Java service for multi-channel notifications",
      version: "v4.1.2",
      type: "jar",
      platform: "jvm",
      size: "35.6 MB",
      checksum: "sha256:q7r8s9t0...",
      status: "quarantined",
      uploadedAt: "2024-01-12T14:00:00Z",
      functions: ["sendEmail", "sendSMS", "sendPush"],
      dependencies: ["java-11"],
      issue: "Suspicious network calls detected"
    }
  ]

  const filteredBinaries = binaries.filter(binary => {
    if (!searchQuery) return true
    const searchLower = searchQuery.toLowerCase()
    return binary.displayName.toLowerCase().includes(searchLower) ||
           binary.description.toLowerCase().includes(searchLower) ||
           binary.type.toLowerCase().includes(searchLower) ||
           binary.platform.toLowerCase().includes(searchLower)
  })

  const verifiedCount = binaries.filter(b => b.status === "verified").length
  const totalSize = binaries.reduce((sum, b) => {
    const size = parseFloat(b.size.split(' ')[0])
    return sum + size
  }, 0)
  const uniquePlatforms = [...new Set(binaries.map(b => b.platform))].length

  const getTypeIcon = (type: string) => {
    switch(type) {
      case 'executable': return <Package className="h-4 w-4" />
      case 'library': return <Archive className="h-4 w-4" />
      case 'wasm': return <Box className="h-4 w-4" />
      case 'jar': return <Package2 className="h-4 w-4" />
      default: return <FileArchive className="h-4 w-4" />
    }
  }

  const getStatusBadge = (status: string, issue?: string) => {
    switch(status) {
      case 'verified':
        return (
          <Badge variant="default" className="font-normal">
            <CheckCircle className="h-3 w-3 mr-1.5" />
            Verified
          </Badge>
        )
      case 'scanning':
        return (
          <Badge variant="secondary" className="font-normal">
            <Shield className="h-3 w-3 mr-1.5 animate-pulse" />
            Scanning
          </Badge>
        )
      case 'quarantined':
        return (
          <Badge variant="destructive" className="font-normal" title={issue}>
            <AlertCircle className="h-3 w-3 mr-1.5" />
            Quarantined
          </Badge>
        )
      default:
        return null
    }
  }

  const formatUploadTime = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    })
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
                <BreadcrumbPage>Binaries</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-2xl font-semibold">Binaries & Packages</h1>
              <p className="text-sm text-muted-foreground mt-1">
                Manage compiled binaries and packages for function execution
              </p>
            </div>
            <Button size="sm">
              <Upload className="h-4 w-4 mr-2" />
              Upload Binary
            </Button>
          </div>

          {/* Metrics */}
          <div className="flex items-center gap-6 pt-4 border-t">
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{binaries.length}</span>
              <span className="text-sm text-muted-foreground">Binaries</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{verifiedCount}</span>
              <span className="text-sm text-muted-foreground">Verified</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{totalSize.toFixed(1)} MB</span>
              <span className="text-sm text-muted-foreground">Total Size</span>
            </div>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2">
              <span className="text-2xl font-semibold">{uniquePlatforms}</span>
              <span className="text-sm text-muted-foreground">Platforms</span>
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
              placeholder="Search binaries..." 
              className="pl-10"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>

        {/* Binaries Table */}
        <div className="border">
        <table className="w-full">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="text-left p-4 font-medium text-sm">Binary</th>
              <th className="text-left p-4 font-medium text-sm">Type</th>
              <th className="text-left p-4 font-medium text-sm">Platform</th>
              <th className="text-left p-4 font-medium text-sm">Status</th>
              <th className="text-right p-4 font-medium text-sm">Size</th>
              <th className="text-right p-4 font-medium text-sm">Uploaded</th>
              <th className="w-10"></th>
            </tr>
          </thead>
          <tbody>
            {filteredBinaries.map((binary, index) => (
              <tr 
                key={binary.id} 
                className={`border-b hover:bg-muted/30 transition-colors cursor-pointer ${
                  index === filteredBinaries.length - 1 ? 'border-b-0' : ''
                }`}
                onClick={() => window.location.href = `/binaries/${binary.name}`}
              >
                <td className="p-4">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-primary/10">
                      {getTypeIcon(binary.type)}
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <p className="font-medium">{binary.displayName}</p>
                        <Badge variant="outline" className="text-xs font-normal">
                          {binary.version}
                        </Badge>
                      </div>
                      <p className="text-sm text-muted-foreground">{binary.description}</p>
                    </div>
                  </div>
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-2">
                    {getTypeIcon(binary.type)}
                    <span className="text-sm uppercase">{binary.type}</span>
                  </div>
                </td>
                <td className="p-4">
                  <code className="text-sm bg-muted px-2 py-1">
                    {binary.platform}
                  </code>
                </td>
                <td className="p-4">
                  {getStatusBadge(binary.status, binary.issue)}
                </td>
                <td className="p-4 text-right">
                  <div className="flex items-center justify-end gap-1">
                    <HardDrive className="h-3 w-3 text-muted-foreground" />
                    <span className="font-mono font-medium">{binary.size}</span>
                  </div>
                </td>
                <td className="p-4 text-right">
                  <span className="text-sm text-muted-foreground">
                    {formatUploadTime(binary.uploadedAt)}
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
                        <Info className="h-4 w-4 mr-2" />
                        View Details
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <Copy className="h-4 w-4 mr-2" />
                        Copy Checksum
                      </DropdownMenuItem>
                      <DropdownMenuItem>
                        <Download className="h-4 w-4 mr-2" />
                        Download
                      </DropdownMenuItem>
                      {binary.status === 'quarantined' && (
                        <DropdownMenuItem>
                          <Shield className="h-4 w-4 mr-2" />
                          View Security Report
                        </DropdownMenuItem>
                      )}
                      <DropdownMenuSeparator />
                      <DropdownMenuItem className="text-red-600">
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete Binary
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
        {filteredBinaries.length === 0 && (
          <div className="text-center py-12 border">
            <p className="text-muted-foreground">No binaries found</p>
          </div>
        )}
      </div>
    </div>
  )
}