"use server";

import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface CreateAccessKeyData {
  name: string;
  permissions: string[];
  expires_at?: string;
}

interface UpdateAccessKeyData {
  permissions?: string[];
  expires_at?: string;
}

async function getAuthToken() {
  const cookieStore = await cookies();
  return cookieStore.get("token")?.value;
}

export async function createAccessKey(data: CreateAccessKeyData) {
  try {
    const token = await getAuthToken();
    if (!token) {
      return { success: false, error: "Not authenticated" };
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/access-keys`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`,
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.text();
      return { success: false, error };
    }

    const result = await response.json();
    revalidatePath("/access-keys");
    return { success: true, data: result };
  } catch (error) {
    console.error("Error creating access key:", error);
    return { success: false, error: "Failed to create access key" };
  }
}

export async function updateAccessKey(id: string, data: UpdateAccessKeyData) {
  try {
    const token = await getAuthToken();
    if (!token) {
      return { success: false, error: "Not authenticated" };
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/access-keys/${id}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`,
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.text();
      return { success: false, error };
    }

    const result = await response.json();
    revalidatePath("/access-keys");
    return { success: true, data: result };
  } catch (error) {
    console.error("Error updating access key:", error);
    return { success: false, error: "Failed to update access key" };
  }
}

export async function deleteAccessKey(id: string) {
  try {
    const token = await getAuthToken();
    if (!token) {
      return { success: false, error: "Not authenticated" };
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/access-keys/${id}`, {
      method: "DELETE",
      headers: {
        "Authorization": `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      const error = await response.text();
      return { success: false, error };
    }

    revalidatePath("/access-keys");
    return { success: true };
  } catch (error) {
    console.error("Error deleting access key:", error);
    return { success: false, error: "Failed to delete access key" };
  }
}

export async function getAccessKeys() {
  try {
    const token = await getAuthToken();
    if (!token) {
      return { success: false, error: "Not authenticated" };
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/access-keys`, {
      method: "GET",
      headers: {
        "Authorization": `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      const error = await response.text();
      return { success: false, error };
    }

    const result = await response.json();
    return { success: true, data: result };
  } catch (error) {
    console.error("Error fetching access keys:", error);
    return { success: false, error: "Failed to fetch access keys" };
  }
}