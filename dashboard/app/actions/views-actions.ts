"use server";

import { revalidatePath } from "next/cache";
import { cookies } from "next/headers";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface CreateViewData {
  name: string;
  description?: string;
  collection: string;
  filter?: Record<string, any>;
  fields?: string[];
}

interface UpdateViewData {
  description?: string;
  filter?: Record<string, any>;
  fields?: string[];
}

interface QueryViewParams {
  limit?: number;
  skip?: number;
  filter?: Record<string, any>;
  sort?: Record<string, number>;
}

async function getAuthToken() {
  const cookieStore = await cookies();
  return cookieStore.get("token")?.value;
}

export async function createView(data: CreateViewData) {
  try {
    const token = await getAuthToken();
    if (!token) {
      return { success: false, error: "Not authenticated" };
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/views`, {
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
    revalidatePath("/views");
    return { success: true, data: result };
  } catch (error) {
    console.error("Error creating view:", error);
    return { success: false, error: "Failed to create view" };
  }
}

export async function updateView(name: string, data: UpdateViewData) {
  try {
    const token = await getAuthToken();
    if (!token) {
      return { success: false, error: "Not authenticated" };
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/views/${name}`, {
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
    revalidatePath("/views");
    revalidatePath(`/views/${name}`);
    return { success: true, data: result };
  } catch (error) {
    console.error("Error updating view:", error);
    return { success: false, error: "Failed to update view" };
  }
}

export async function deleteView(name: string) {
  try {
    const token = await getAuthToken();
    if (!token) {
      return { success: false, error: "Not authenticated" };
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/views/${name}`, {
      method: "DELETE",
      headers: {
        "Authorization": `Bearer ${token}`,
      },
    });

    if (!response.ok) {
      const error = await response.text();
      return { success: false, error };
    }

    revalidatePath("/views");
    return { success: true };
  } catch (error) {
    console.error("Error deleting view:", error);
    return { success: false, error: "Failed to delete view" };
  }
}

export async function queryView(name: string, params?: QueryViewParams) {
  try {
    const token = await getAuthToken();
    if (!token) {
      return { success: false, error: "Not authenticated" };
    }

    const queryParams = new URLSearchParams();
    if (params?.limit) queryParams.append("limit", params.limit.toString());
    if (params?.skip) queryParams.append("skip", params.skip.toString());
    if (params?.filter) queryParams.append("filter", JSON.stringify(params.filter));
    if (params?.sort) queryParams.append("sort", JSON.stringify(params.sort));

    const url = `${API_BASE_URL}/api/v1/views/${name}/query${queryParams.toString() ? `?${queryParams}` : ""}`;
    
    const response = await fetch(url, {
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
    
    // Handle different response formats from the API
    if (Array.isArray(result)) {
      return { success: true, data: result, total: result.length };
    } else if (result && typeof result === 'object') {
      // Check for various possible field names
      const data = result.data || result.results || result.documents || result.items || [];
      const total = result.total || result.count || result.totalCount || data.length;
      return { success: true, data, total };
    }
    
    return { success: true, data: [], total: 0 };
  } catch (error) {
    console.error("Error querying view:", error);
    return { success: false, error: "Failed to query view" };
  }
}

export async function getViews() {
  try {
    const token = await getAuthToken();
    if (!token) {
      return { success: false, error: "Not authenticated" };
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/views`, {
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
    console.error("Error fetching views:", error);
    return { success: false, error: "Failed to fetch views" };
  }
}

export async function getView(name: string) {
  try {
    const token = await getAuthToken();
    if (!token) {
      return { success: false, error: "Not authenticated" };
    }

    const response = await fetch(`${API_BASE_URL}/api/v1/views/${name}`, {
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
    console.error("Error fetching view:", error);
    return { success: false, error: "Failed to fetch view" };
  }
}