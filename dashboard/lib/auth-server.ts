import { cookies } from "next/headers";

export interface User {
  id: string;
  email: string;
  role: string;
  roles?: string[];
  first_name?: string;
  last_name?: string;
}

export async function getCurrentUser(): Promise<User | null> {
  const cookieStore = await cookies();
  const userCookie = cookieStore.get("user");
  
  if (!userCookie) {
    return null;
  }
  
  try {
    return JSON.parse(userCookie.value);
  } catch {
    return null;
  }
}

export async function getAuthToken(): Promise<string | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("token");
  
  return token ? token.value : null;
}