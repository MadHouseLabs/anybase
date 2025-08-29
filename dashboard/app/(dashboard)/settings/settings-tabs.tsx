"use client";

import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useToast } from "@/hooks/use-toast";
import { updateUserSettings, updateSystemSettings } from "@/app/actions/settings-actions";
import { User, Shield, Bell, Globe, Palette, Clock, Save } from "lucide-react";

interface SettingsTabsProps {
  initialUserSettings: any;
  initialSystemSettings: any;
  userRole: string;
}

export function SettingsTabs({ initialUserSettings, initialSystemSettings, userRole }: SettingsTabsProps) {
  const { toast } = useToast();
  const [userSettings, setUserSettings] = useState(initialUserSettings);
  const [systemSettings, setSystemSettings] = useState(initialSystemSettings);
  const [isSaving, setIsSaving] = useState(false);

  const handleSaveUserSettings = async () => {
    setIsSaving(true);
    try {
      const result = await updateUserSettings(userSettings);
      if (result.success) {
        toast({
          title: "Success",
          description: "User settings updated successfully",
        });
      } else {
        throw new Error(result.error);
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update user settings",
        variant: "destructive",
      });
    } finally {
      setIsSaving(false);
    }
  };

  const handleSaveSystemSettings = async () => {
    setIsSaving(true);
    try {
      const result = await updateSystemSettings(systemSettings);
      if (result.success) {
        toast({
          title: "Success",
          description: "System settings updated successfully",
        });
      } else {
        throw new Error(result.error);
      }
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to update system settings",
        variant: "destructive",
      });
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Tabs defaultValue="user" className="space-y-4">
      <TabsList>
        <TabsTrigger value="user" className="flex items-center gap-2">
          <User className="h-4 w-4" />
          User Settings
        </TabsTrigger>
        {(userRole === "admin" || userRole === "developer") && (
          <TabsTrigger value="system" className="flex items-center gap-2">
            <Shield className="h-4 w-4" />
            System Settings
          </TabsTrigger>
        )}
      </TabsList>

      <TabsContent value="user" className="space-y-4">
        <Card className="rounded-none shadow-none">
          <CardHeader>
            <CardTitle>Appearance</CardTitle>
            <CardDescription>
              Customize how the application looks
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="theme">Theme</Label>
              <Select
                value={userSettings.theme}
                onValueChange={(value) => setUserSettings({ ...userSettings, theme: value })}
              >
                <SelectTrigger id="theme">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="light">Light</SelectItem>
                  <SelectItem value="dark">Dark</SelectItem>
                  <SelectItem value="system">System</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="language">Language</Label>
              <Select
                value={userSettings.language}
                onValueChange={(value) => setUserSettings({ ...userSettings, language: value })}
              >
                <SelectTrigger id="language">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="en">English</SelectItem>
                  <SelectItem value="es">Spanish</SelectItem>
                  <SelectItem value="fr">French</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        <Card className="rounded-none shadow-none">
          <CardHeader>
            <CardTitle>Regional Settings</CardTitle>
            <CardDescription>
              Configure date, time, and timezone preferences
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="timezone">Timezone</Label>
              <Input
                id="timezone"
                value={userSettings.timezone}
                onChange={(e) => setUserSettings({ ...userSettings, timezone: e.target.value })}
                placeholder="UTC"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="date-format">Date Format</Label>
              <Select
                value={userSettings.date_format}
                onValueChange={(value) => setUserSettings({ ...userSettings, date_format: value })}
              >
                <SelectTrigger id="date-format">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="MM/DD/YYYY">MM/DD/YYYY</SelectItem>
                  <SelectItem value="DD/MM/YYYY">DD/MM/YYYY</SelectItem>
                  <SelectItem value="YYYY-MM-DD">YYYY-MM-DD</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="time-format">Time Format</Label>
              <Select
                value={userSettings.time_format}
                onValueChange={(value) => setUserSettings({ ...userSettings, time_format: value })}
              >
                <SelectTrigger id="time-format">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="12">12 Hour</SelectItem>
                  <SelectItem value="24">24 Hour</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        <Card className="rounded-none shadow-none">
          <CardHeader>
            <CardTitle>Notifications</CardTitle>
            <CardDescription>
              Configure your notification preferences
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="email-notifications">Email Notifications</Label>
                <p className="text-sm text-muted-foreground">
                  Receive email notifications for important events
                </p>
              </div>
              <Switch
                id="email-notifications"
                checked={userSettings.email_notifications}
                onCheckedChange={(checked) => 
                  setUserSettings({ ...userSettings, email_notifications: checked })
                }
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="security-alerts">Security Alerts</Label>
                <p className="text-sm text-muted-foreground">
                  Get notified about security-related events
                </p>
              </div>
              <Switch
                id="security-alerts"
                checked={userSettings.security_alerts}
                onCheckedChange={(checked) => 
                  setUserSettings({ ...userSettings, security_alerts: checked })
                }
              />
            </div>
          </CardContent>
        </Card>

        <div className="flex justify-end">
          <Button onClick={handleSaveUserSettings} disabled={isSaving}>
            <Save className="h-4 w-4 mr-2" />
            {isSaving ? "Saving..." : "Save User Settings"}
          </Button>
        </div>
      </TabsContent>

      {(userRole === "admin" || userRole === "developer") && (
        <TabsContent value="system" className="space-y-4">
          <Card className="rounded-none shadow-none">
            <CardHeader>
              <CardTitle>System Configuration</CardTitle>
              <CardDescription>
                Configure system-wide settings (Admin only)
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {systemSettings ? (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="connection-pool">Connection Pool Size</Label>
                    <Input
                      id="connection-pool"
                      type="number"
                      value={systemSettings.connection_pool_size || 10}
                      onChange={(e) => 
                        setSystemSettings({ 
                          ...systemSettings, 
                          connection_pool_size: parseInt(e.target.value) 
                        })
                      }
                      disabled={userRole !== "admin"}
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="query-timeout">Query Timeout (seconds)</Label>
                    <Input
                      id="query-timeout"
                      type="number"
                      value={systemSettings.query_timeout || 30}
                      onChange={(e) => 
                        setSystemSettings({ 
                          ...systemSettings, 
                          query_timeout: parseInt(e.target.value) 
                        })
                      }
                      disabled={userRole !== "admin"}
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label htmlFor="compression">Compression Enabled</Label>
                      <p className="text-sm text-muted-foreground">
                        Enable data compression for storage optimization
                      </p>
                    </div>
                    <Switch
                      id="compression"
                      checked={systemSettings.compression_enabled}
                      onCheckedChange={(checked) => 
                        setSystemSettings({ ...systemSettings, compression_enabled: checked })
                      }
                      disabled={userRole !== "admin"}
                    />
                  </div>

                  <div className="flex items-center justify-between">
                    <div className="space-y-0.5">
                      <Label htmlFor="encryption">Encryption Enabled</Label>
                      <p className="text-sm text-muted-foreground">
                        Enable encryption for sensitive data
                      </p>
                    </div>
                    <Switch
                      id="encryption"
                      checked={systemSettings.encryption_enabled}
                      onCheckedChange={(checked) => 
                        setSystemSettings({ ...systemSettings, encryption_enabled: checked })
                      }
                      disabled={userRole !== "admin"}
                    />
                  </div>
                </>
              ) : (
                <p className="text-muted-foreground">
                  System settings not available. Please check your permissions.
                </p>
              )}
            </CardContent>
          </Card>

          {userRole === "admin" && systemSettings && (
            <div className="flex justify-end">
              <Button onClick={handleSaveSystemSettings} disabled={isSaving}>
                <Save className="h-4 w-4 mr-2" />
                {isSaving ? "Saving..." : "Save System Settings"}
              </Button>
            </div>
          )}
        </TabsContent>
      )}
    </Tabs>
  );
}