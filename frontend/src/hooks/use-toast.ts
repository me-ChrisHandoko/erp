/**
 * Simple toast hook
 * For production, consider using sonner or react-hot-toast
 */

import { useState, useCallback } from "react";

export interface Toast {
  title: string;
  description?: string;
  variant?: "default" | "destructive";
}

export function useToast() {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const toast = useCallback((props: Toast) => {
    // For now, just use console.log
    // In production, integrate with a proper toast library
    if (props.variant === "destructive") {
      console.error(`[Toast] ${props.title}:`, props.description);
    } else {
      console.log(`[Toast] ${props.title}:`, props.description);
    }

    // TODO: Integrate with actual toast UI library (sonner recommended)
    setToasts((prev) => [...prev, props]);
  }, []);

  return { toast, toasts };
}
