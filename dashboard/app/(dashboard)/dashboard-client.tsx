"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Plus, Database, Eye, Key, Users as UsersIcon } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

export function QuickActions() {
  const router = useRouter();

  const actions = [
    {
      title: "Create Collection",
      description: "Set up a new data collection",
      icon: Database,
      onClick: () => router.push("/collections"),
      color: "text-blue-500",
    },
    {
      title: "Create View",
      description: "Design a custom data view",
      icon: Eye,
      onClick: () => router.push("/views"),
      color: "text-green-500",
    },
    {
      title: "Generate API Key",
      description: "Create a new access key",
      icon: Key,
      onClick: () => router.push("/access-keys"),
      color: "text-yellow-500",
    },
    {
      title: "Manage Users",
      description: "Add or modify user accounts",
      icon: UsersIcon,
      onClick: () => router.push("/users"),
      color: "text-purple-500",
    },
  ];

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {actions.map((action) => (
        <Card
          key={action.title}
          className="cursor-pointer hover:shadow-md transition-shadow"
          onClick={action.onClick}
        >
          <CardContent className="p-6">
            <div className="flex items-center gap-4">
              <div className={`p-2 rounded-lg bg-muted ${action.color}`}>
                <action.icon className="h-6 w-6" />
              </div>
              <div className="flex-1">
                <h3 className="font-semibold text-sm">{action.title}</h3>
                <p className="text-xs text-muted-foreground mt-1">
                  {action.description}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

export function RecentActivity() {
  const activities = [
    {
      id: 1,
      type: "collection",
      action: "created",
      target: "users_data",
      timestamp: "2 hours ago",
      user: "admin@example.com",
    },
    {
      id: 2,
      type: "document",
      action: "updated",
      target: "config.json",
      timestamp: "5 hours ago",
      user: "dev@example.com",
    },
    {
      id: 3,
      type: "api_key",
      action: "generated",
      target: "Production API Key",
      timestamp: "1 day ago",
      user: "admin@example.com",
    },
    {
      id: 4,
      type: "view",
      action: "created",
      target: "Analytics Dashboard",
      timestamp: "2 days ago",
      user: "analyst@example.com",
    },
  ];

  return (
    <div className="space-y-4">
      {activities.map((activity) => (
        <div
          key={activity.id}
          className="flex items-center justify-between py-3 border-b last:border-0"
        >
          <div className="flex items-center gap-3">
            <div className="w-2 h-2 rounded-full bg-blue-500" />
            <div>
              <p className="text-sm">
                <span className="font-medium">{activity.user}</span>{" "}
                {activity.action}{" "}
                <span className="font-medium">{activity.target}</span>
              </p>
              <p className="text-xs text-muted-foreground">
                {activity.timestamp}
              </p>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}