import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * Format currency in Indonesian Rupiah format
 * @param amount - Amount as string or number
 * @returns Formatted string like "Rp 1.000.000"
 */
export function formatCurrency(amount: string | number): string {
  return `Rp ${Number(amount).toLocaleString("id-ID")}`;
}

/**
 * Format ISO date to Indonesian date format
 * @param isoDate - ISO 8601 date string
 * @returns Formatted string like "15 Jan 2024"
 */
export function formatDate(isoDate: string): string {
  const date = new Date(isoDate);
  return date.toLocaleDateString("id-ID", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

/**
 * Format ISO date to Indonesian date and time format
 * @param isoDate - ISO 8601 date string
 * @returns Formatted string like "15 Jan 2024, 14:30"
 */
export function formatDateTime(isoDate: string): string {
  const date = new Date(isoDate);
  return date.toLocaleDateString("id-ID", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
