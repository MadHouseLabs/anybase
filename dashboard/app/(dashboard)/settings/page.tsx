import { getUserSettings, getSystemSettings } from "@/lib/api-server";
import { getCurrentUser, getAuthToken } from "@/lib/auth-server";
import { Settings } from "lucide-react";
import { SettingsTabs } from "./settings-tabs";

export default async function SettingsPage() {
  const [userSettingsData, systemSettingsData, currentUser, token] = await Promise.all([
    getUserSettings(),
    getSystemSettings(),
    getCurrentUser(),
    getAuthToken()
  ]);

  // Default user settings if none exist
  const defaultUserSettings = {
    theme: "system",
    language: "en",
    timezone: "UTC",
    date_format: "MM/DD/YYYY",
    time_format: "12",
    email_notifications: true,
    security_alerts: true
  };

  const userSettings = userSettingsData || defaultUserSettings;
  
  // Get user role from current user or token
  let userRole = currentUser?.role || "user";
  if (!userRole && token) {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      userRole = payload.role || "user";
    } catch (error) {
      // Fallback to user role
      userRole = "user";
    }
  }

  return (
    <div className="container mx-auto py-6 px-4 lg:px-8">
      <div className="max-w-7xl mx-auto space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Settings className="h-8 w-8" />
            Settings
          </h1>
          <p className="text-muted-foreground mt-2">
            Configure your preferences and view system information
          </p>
        </div>

        <SettingsTabs 
          initialUserSettings={userSettings}
          initialSystemSettings={systemSettingsData}
          userRole={userRole}
        />
      </div>
    </div>
  );
}