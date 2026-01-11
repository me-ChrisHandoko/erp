/**
 * Toast hook using sonner
 */

import { toast as sonnerToast } from "sonner";
import { useCallback } from "react";

export interface Toast {
  title: string;
  description?: string;
  variant?: "default" | "destructive";
}

export function useToast() {
  const toast = useCallback((props: Toast) => {
    if (props.variant === "destructive") {
      sonnerToast.error(props.title, {
        description: props.description,
      });
    } else {
      sonnerToast.success(props.title, {
        description: props.description,
      });
    }
  }, []);

  return { toast };
}
