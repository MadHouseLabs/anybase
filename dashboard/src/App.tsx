import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { MainLayout } from '@/components/layout/main-layout'
import { LoginPage } from '@/pages/auth/login'
import { DashboardPage } from '@/pages/dashboard/dashboard'
import { CollectionsPage } from '@/pages/collections/collections'
import { CollectionDetailPage } from '@/pages/collections/collection-detail'
import { AccessKeysPage } from '@/pages/accesskeys/accesskeys'
import { ViewsPage } from '@/pages/views/views'
import { UsersPage } from '@/pages/users/users'
import { SettingsPage } from '@/pages/settings/settings'
import { ApiDocsPage } from '@/pages/docs/api-docs'
import { Toaster } from '@/components/ui/toaster'
import './index.css'

function App() {
  const isAuthenticated = !!localStorage.getItem('token')

  const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
    return isAuthenticated ? children : <Navigate to="/login" />
  }


  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={!isAuthenticated ? <LoginPage /> : <Navigate to="/" />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <MainLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<DashboardPage />} />
          <Route path="collections" element={<CollectionsPage />} />
          <Route path="collections/:collectionName" element={<CollectionDetailPage />} />
          <Route path="views" element={<ViewsPage />} />
          <Route path="users" element={<UsersPage />} />
          <Route path="access-keys" element={<AccessKeysPage />} />
          <Route path="api-docs" element={<ApiDocsPage />} />
          <Route path="settings" element={<SettingsPage />} />
          <Route path="*" element={<Navigate to="/" />} />
        </Route>
      </Routes>
      <Toaster />
    </BrowserRouter>
  )
}

export default App;