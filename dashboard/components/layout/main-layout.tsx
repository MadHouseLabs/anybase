import { Outlet } from "react-router-dom"
import { Sidebar } from "./sidebar"
import { Header } from "./header"

export function MainLayout() {
  return (
    <div className="flex h-screen">
      <div className="hidden lg:block border-r">
        <Sidebar />
      </div>
      <div className="flex-1 flex flex-col">
        <Header />
        <main className="flex-1 overflow-y-auto bg-muted/10">
          <Outlet />
        </main>
      </div>
    </div>
  )
}