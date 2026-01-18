/**
 * NavigationProgress Component
 *
 * Menampilkan progress bar di bagian atas halaman saat navigasi antar halaman.
 * Memberikan visual feedback tambahan untuk user experience yang lebih baik.
 *
 * Cara kerja:
 * 1. Mendeteksi perubahan pathname/searchParams via Next.js hooks
 * 2. Menampilkan progress bar animasi saat navigasi berlangsung
 * 3. Menyembunyikan progress bar setelah navigasi selesai
 *
 * Catatan: Component ini harus dibungkus dengan Suspense karena
 * menggunakan useSearchParams() yang membutuhkan Suspense boundary.
 */

"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { usePathname, useSearchParams } from "next/navigation";
import { cn } from "@/lib/utils";

function NavigationProgressBar() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [isNavigating, setIsNavigating] = useState(false);
  const [progress, setProgress] = useState(0);
  const progressInterval = useRef<NodeJS.Timeout | null>(null);
  const previousUrl = useRef<string>("");

  // Build current URL for comparison
  const currentUrl = `${pathname}${searchParams.toString() ? `?${searchParams.toString()}` : ""}`;

  // Cleanup interval
  const clearProgressInterval = useCallback(() => {
    if (progressInterval.current) {
      clearInterval(progressInterval.current);
      progressInterval.current = null;
    }
  }, []);

  // Start progress animation
  const startProgress = useCallback(() => {
    clearProgressInterval();
    setIsNavigating(true);
    setProgress(0);

    // Animate progress from 0 to 90% (never reaches 100% until navigation completes)
    progressInterval.current = setInterval(() => {
      setProgress((prev) => {
        if (prev >= 90) {
          clearProgressInterval();
          return 90;
        }
        // Slow down as we approach 90%
        const increment = Math.max(1, (90 - prev) / 10);
        return Math.min(90, prev + increment);
      });
    }, 100);
  }, [clearProgressInterval]);

  // Complete progress animation
  const completeProgress = useCallback(() => {
    clearProgressInterval();
    setProgress(100);

    // Hide after animation completes
    const hideTimer = setTimeout(() => {
      setIsNavigating(false);
      setProgress(0);
    }, 300);

    return () => clearTimeout(hideTimer);
  }, [clearProgressInterval]);

  // Detect navigation by URL change
  useEffect(() => {
    if (previousUrl.current && previousUrl.current !== currentUrl) {
      // URL changed - navigation completed
      completeProgress();
    }
    previousUrl.current = currentUrl;
  }, [currentUrl, completeProgress]);

  // Listen for click events on navigation links to start progress early
  useEffect(() => {
    const handleClick = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      const link = target.closest("a");

      if (link) {
        const href = link.getAttribute("href");
        // Check if it's an internal navigation link
        if (
          href &&
          href.startsWith("/") &&
          !href.startsWith("//") &&
          !link.hasAttribute("target")
        ) {
          // Don't start progress if clicking current page
          if (href !== pathname && !href.startsWith(pathname + "#")) {
            startProgress();
          }
        }
      }
    };

    document.addEventListener("click", handleClick);
    return () => document.removeEventListener("click", handleClick);
  }, [pathname, startProgress]);

  // Cleanup on unmount
  useEffect(() => {
    return () => clearProgressInterval();
  }, [clearProgressInterval]);

  if (!isNavigating && progress === 0) {
    return null;
  }

  return (
    <div
      className="fixed top-0 left-0 right-0 z-[100] h-0.5 bg-transparent"
      role="progressbar"
      aria-valuenow={progress}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-label="Memuat halaman"
    >
      <div
        className={cn(
          "h-full bg-primary shadow-sm shadow-primary/50 transition-all ease-out",
          progress < 100 ? "duration-100" : "duration-200",
          progress === 100 && "opacity-0"
        )}
        style={{ width: `${progress}%` }}
      />
    </div>
  );
}

// Export wrapped component with built-in error boundary consideration
export function NavigationProgress() {
  return <NavigationProgressBar />;
}
