import { Bell } from "lucide-react";
import { Button } from "@/components/ui/button";
import { SearchBar } from "./search-bar";
import { UserMenu } from "./user-menu";
import { getCurrentUser } from "@/lib/auth-server";

export async function Header() {
  const user = await getCurrentUser();

  return (
    <header className="border-b">
      <div className="flex h-16 items-center px-4">
        <div className="ml-auto flex items-center space-x-4">
          <SearchBar />
          <Button variant="ghost" size="icon">
            <Bell className="h-4 w-4" />
          </Button>
          <UserMenu initialUser={user} />
        </div>
      </div>
    </header>
  );
}