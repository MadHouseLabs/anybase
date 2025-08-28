"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useToast } from "@/hooks/use-toast";
import { settingsApi } from "@/lib/settings";
import type { UserSettings, SystemSettings } from "@/lib/settings";
import { 
  Globe, 
  Moon,
  Sun,
  Monitor,
  Save,
  Server,
  Shield,
  Bell,
  Mail,
  Activity,
  Info,
  Loader2
} from "lucide-react";

interface SettingsClientProps {
  initialUserSettings: UserSettings;
  initialSystemSettings: SystemSettings | null;
  userRole: string;
}

export function SettingsTabs({ initialUserSettings, initialSystemSettings, userRole }: SettingsClientProps) {
  const { toast } = useToast();
  const [saving, setSaving] = useState(false);
  const [userSettings, setUserSettings] = useState<UserSettings>(initialUserSettings);
  const [systemSettings] = useState<SystemSettings | null>(initialSystemSettings);

  const handleSaveUserSettings = async () => {
    try {
      setSaving(true);
      await settingsApi.updateUserSettings(userSettings);
      toast({
        title: "Settings saved",
        description: "Your preferences have been updated successfully.",
      });
    } catch (error) {
      toast({
        title: "Error",
        description: "Failed to save settings. Please try again.",
        variant: "destructive"
      });
    } finally {
      setSaving(false);
    }
  };

  return (
    <Tabs defaultValue="general" className="space-y-4">
      <TabsList className="grid w-full max-w-[500px] grid-cols-3">
        <TabsTrigger value="general">General</TabsTrigger>
        <TabsTrigger value="notifications">Notifications</TabsTrigger>
        <TabsTrigger value="system">System Info</TabsTrigger>
      </TabsList>

      {/* General Settings */}
      <TabsContent value="general" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Monitor className="h-5 w-5" />
              Appearance
            </CardTitle>
            <CardDescription>
              Customize the look and feel of your dashboard
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-4">
              <div>
                <Label htmlFor="theme">Theme Preference</Label>
                <p className="text-sm text-muted-foreground mb-2">
                  Select your preferred color scheme
                </p>
                <div className="grid grid-cols-3 gap-4 mt-2">
                  <Button
                    variant={userSettings.theme === "light" ? "default" : "outline"}
                    className="justify-start"
                    onClick={() => setUserSettings({...userSettings, theme: "light"})}
                  >
                    <Sun className="h-4 w-4 mr-2" />
                    Light
                  </Button>
                  <Button
                    variant={userSettings.theme === "dark" ? "default" : "outline"}
                    className="justify-start"
                    onClick={() => setUserSettings({...userSettings, theme: "dark"})}
                  >
                    <Moon className="h-4 w-4 mr-2" />
                    Dark
                  </Button>
                  <Button
                    variant={userSettings.theme === "system" ? "default" : "outline"}
                    className="justify-start"
                    onClick={() => setUserSettings({...userSettings, theme: "system"})}
                  >
                    <Monitor className="h-4 w-4 mr-2" />
                    System
                  </Button>
                </div>
              </div>
            </div>
            
            <div className="flex justify-end">
              <Button onClick={handleSaveUserSettings} disabled={saving}>
                {saving ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <Save className="h-4 w-4 mr-2" />
                )}
                Save Changes
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe className="h-5 w-5" />
              Regional Settings
            </CardTitle>
            <CardDescription>
              Configure language, timezone, and date formats
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <div>
                <Label htmlFor="language">Language</Label>
                <Select 
                  value={userSettings.language}
                  onValueChange={(value) => setUserSettings({...userSettings, language: value})}
                >
                  <SelectTrigger id="language">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="en">English</SelectItem>
                    <SelectItem value="es">Spanish</SelectItem>
                    <SelectItem value="fr">French</SelectItem>
                    <SelectItem value="de">German</SelectItem>
                    <SelectItem value="zh">Chinese</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              
              <div>
                <Label htmlFor="timezone">Timezone</Label>
                <Select 
                  value={userSettings.timezone}
                  onValueChange={(value) => setUserSettings({...userSettings, timezone: value})}
                >
                  <SelectTrigger id="timezone">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="UTC">UTC</SelectItem>
                    <SelectItem value="EST">EST (Eastern)</SelectItem>
                    <SelectItem value="CST">CST (Central)</SelectItem>
                    <SelectItem value="MST">MST (Mountain)</SelectItem>
                    <SelectItem value="PST">PST (Pacific)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              
              <div>
                <Label htmlFor="dateformat">Date Format</Label>
                <Select 
                  value={userSettings.date_format}
                  onValueChange={(value) => setUserSettings({...userSettings, date_format: value})}
                >
                  <SelectTrigger id="dateformat">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="MM/DD/YYYY">MM/DD/YYYY</SelectItem>
                    <SelectItem value="DD/MM/YYYY">DD/MM/YYYY</SelectItem>
                    <SelectItem value="YYYY-MM-DD">YYYY-MM-DD</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              
              <div>
                <Label htmlFor="timeformat">Time Format</Label>
                <Select 
                  value={userSettings.time_format}
                  onValueChange={(value) => setUserSettings({...userSettings, time_format: value})}
                >
                  <SelectTrigger id="timeformat">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="12">12 Hour</SelectItem>
                    <SelectItem value="24">24 Hour</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            
            <div className="flex justify-end">
              <Button onClick={handleSaveUserSettings} disabled={saving}>
                {saving ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <Save className="h-4 w-4 mr-2" />
                )}
                Save Changes
              </Button>
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      {/* Notifications Settings */}
      <TabsContent value="notifications" className="space-y-4">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5" />
              Notification Preferences
            </CardTitle>
            <CardDescription>
              Choose how you want to receive notifications
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label htmlFor="email-notif">Email Notifications</Label>
                  <p className="text-sm text-muted-foreground">
                    Receive important updates via email
                  </p>
                </div>
                <Switch 
                  id="email-notif"
                  checked={userSettings.email_notifications}
                  onCheckedChange={(checked) => setUserSettings({...userSettings, email_notifications: checked})}
                />
              </div>
              
              <div className="flex items-center justify-between">
                <div>
                  <Label htmlFor="security-alerts">Security Alerts</Label>
                  <p className="text-sm text-muted-foreground">
                    Immediate alerts for security-related events
                  </p>
                </div>
                <Switch 
                  id="security-alerts"
                  checked={userSettings.security_alerts}
                  onCheckedChange={(checked) => setUserSettings({...userSettings, security_alerts: checked})}
                />
              </div>
            </div>
            
            <div className="flex justify-end">
              <Button onClick={handleSaveUserSettings} disabled={saving}>
                {saving ? (
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <Save className="h-4 w-4 mr-2" />
                )}
                Save Changes
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Mail className="h-5 w-5" />
              Notification Status
            </CardTitle>
            <CardDescription>
              Current notification configuration
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert>
              <Info className="h-4 w-4" />
              <AlertDescription>
                {userSettings.email_notifications ? (
                  <span>Email notifications are <strong className="text-green-600">enabled</strong>. You will receive updates about important system events.</span>
                ) : (
                  <span>Email notifications are <strong className="text-orange-600">disabled</strong>. You won't receive email updates.</span>
                )}
              </AlertDescription>
            </Alert>
            
            <Alert>
              <Shield className="h-4 w-4" />
              <AlertDescription>
                {userSettings.security_alerts ? (
                  <span>Security alerts are <strong className="text-green-600">enabled</strong>. You will be notified of security events immediately.</span>
                ) : (
                  <span>Security alerts are <strong className="text-orange-600">disabled</strong>. Consider enabling for better security.</span>
                )}
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>
      </TabsContent>

      {/* System Information */}
      <TabsContent value="system" className="space-y-4">
        {systemSettings && (
          <>
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Server className="h-5 w-5" />
                  Database Configuration
                </CardTitle>
                <CardDescription>
                  Current database settings (read-only)
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-4 md:grid-cols-2">
                  <div>
                    <Label>Connection Pool Size</Label>
                    <p className="text-2xl font-semibold">{systemSettings.connection_pool_size}</p>
                    <p className="text-xs text-muted-foreground">Maximum concurrent connections</p>
                  </div>
                  
                  <div>
                    <Label>Query Timeout</Label>
                    <p className="text-2xl font-semibold">{systemSettings.query_timeout}s</p>
                    <p className="text-xs text-muted-foreground">Maximum query execution time</p>
                  </div>
                  
                  <div>
                    <Label>Max Retries</Label>
                    <p className="text-2xl font-semibold">{systemSettings.max_retries}</p>
                    <p className="text-xs text-muted-foreground">Retry attempts on failure</p>
                  </div>
                  
                  <div>
                    <Label>Features</Label>
                    <div className="flex gap-2 mt-1">
                      {systemSettings.compression_enabled && (
                        <Badge variant="secondary">Compression</Badge>
                      )}
                      {systemSettings.encryption_enabled && (
                        <Badge variant="secondary">Encryption</Badge>
                      )}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>

            {userRole === 'admin' && systemSettings.session_timeout !== undefined && (
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Shield className="h-5 w-5" />
                    Security Configuration
                  </CardTitle>
                  <CardDescription>
                    Security settings (admin view)
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="grid gap-4 md:grid-cols-2">
                    <div>
                      <Label>Session Timeout</Label>
                      <p className="text-2xl font-semibold">{systemSettings.session_timeout}h</p>
                      <p className="text-xs text-muted-foreground">Auto-logout after inactivity</p>
                    </div>
                    
                    <div>
                      <Label>Password Policy</Label>
                      <Badge className="mt-1">{systemSettings.password_policy}</Badge>
                    </div>
                    
                    <div>
                      <Label>Security Features</Label>
                      <div className="flex gap-2 mt-1">
                        {systemSettings.mfa_required && (
                          <Badge variant="secondary">2FA Required</Badge>
                        )}
                        {systemSettings.audit_log_enabled && (
                          <Badge variant="secondary">Audit Logging</Badge>
                        )}
                      </div>
                    </div>
                    
                    <div>
                      <Label>API Rate Limit</Label>
                      <p className="text-2xl font-semibold">{systemSettings.rate_limit}/min</p>
                      <p className="text-xs text-muted-foreground">Requests per minute</p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Activity className="h-5 w-5" />
                  System Status
                </CardTitle>
                <CardDescription>
                  API configuration and status
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <Alert>
                  <Activity className="h-4 w-4" />
                  <AlertDescription className="flex items-center justify-between">
                    <span>API Rate Limiting</span>
                    <Badge variant="default">{systemSettings.rate_limit} requests/min</Badge>
                  </AlertDescription>
                </Alert>
                
                <Alert>
                  <Server className="h-4 w-4" />
                  <AlertDescription className="flex items-center justify-between">
                    <span>CORS</span>
                    <Badge variant={systemSettings.cors_enabled ? "default" : "secondary"}>
                      {systemSettings.cors_enabled ? "Enabled" : "Disabled"}
                    </Badge>
                  </AlertDescription>
                </Alert>
                
                <Alert>
                  <Shield className="h-4 w-4" />
                  <AlertDescription className="flex items-center justify-between">
                    <span>Data Encryption</span>
                    <Badge variant={systemSettings.encryption_enabled ? "default" : "secondary"}>
                      {systemSettings.encryption_enabled ? "Enabled" : "Disabled"}
                    </Badge>
                  </AlertDescription>
                </Alert>
              </CardContent>
            </Card>
          </>
        )}
      </TabsContent>
    </Tabs>
  );
}