"use server";

import { revalidatePath } from "next/cache";
import { getAuthToken } from "@/lib/auth-server";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface CreateCollectionData {
  name: string;
  description?: string;
  settings?: {
    versioning?: boolean;
    soft_delete?: boolean;
    auditing?: boolean;
  };
}

export async function createCollection(data: CreateCollectionData) {
  try {
    const token = await getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/api/v1/collections`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${token}`,
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to create collection");
    }

    const result = await response.json();
    
    // Revalidate the collections page to show the new collection
    revalidatePath("/collections");
    
    return { success: true, data: result };
  } catch (error) {
    return { 
      success: false, 
      error: error instanceof Error ? error.message : "Failed to create collection" 
    };
  }
}

export async function deleteCollection(name: string) {
  try {
    const token = await getAuthToken();
    
    const response = await fetch(`${API_BASE_URL}/api/v1/collections/${name}`, {
      method: "DELETE",
      headers: {
        "Authorization": `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to delete collection");
    }

    // Revalidate the collections page to remove the deleted collection
    revalidatePath("/collections");
    
    return { success: true };
  } catch (error) {
    return { 
      success: false, 
      error: error instanceof Error ? error.message : "Failed to delete collection" 
    };
  }
}