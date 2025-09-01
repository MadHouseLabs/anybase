import { format, formatDistanceToNow } from 'date-fns';

/**
 * Convert UTC date string to local timezone Date object
 * Backend stores all dates in UTC, this ensures proper local display
 */
export function toLocalDate(dateString: string | Date | null | undefined): Date | null {
  if (!dateString) return null;
  
  // If it's already a Date object, ensure it's properly handling timezone
  if (dateString instanceof Date) {
    return dateString;
  }
  
  // Parse the UTC date string
  const date = new Date(dateString);
  
  // Check if date is valid
  if (isNaN(date.getTime())) {
    return null;
  }
  
  return date;
}

/**
 * Format a UTC date string to local time display
 */
export function formatLocalDate(
  dateString: string | Date | null | undefined, 
  formatStr: string = 'PPP'
): string {
  const date = toLocalDate(dateString);
  if (!date) return 'N/A';
  
  return format(date, formatStr);
}

/**
 * Format a UTC date string to relative time (e.g., "2 hours ago")
 */
export function formatRelativeTime(
  dateString: string | Date | null | undefined
): string {
  const date = toLocalDate(dateString);
  if (!date) return 'N/A';
  
  return formatDistanceToNow(date, { addSuffix: true });
}

/**
 * Format a UTC date string to local date only
 */
export function formatLocalDateOnly(
  dateString: string | Date | null | undefined
): string {
  return formatLocalDate(dateString, 'MMM d, yyyy');
}

/**
 * Format a UTC date string to local date and time
 */
export function formatLocalDateTime(
  dateString: string | Date | null | undefined
): string {
  return formatLocalDate(dateString, 'MMM d, yyyy HH:mm');
}