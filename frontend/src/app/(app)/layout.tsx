"use client";

import { AppSidebar } from "@/components/app-sidebar";
import { CompanyInitializer } from "@/components/company-initializer";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import { useState, useEffect } from "react";

export default function AppLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  // Auto-close sidebar below 1366px width
  const [isSidebarOpen, setIsSidebarOpen] = useState(() => {
    if (typeof window !== "undefined") {
      return window.innerWidth >= 1366;
    }
    return true;
  });

  useEffect(() => {
    const handleResize = () => {
      const shouldBeOpen = window.innerWidth >= 1366;
      setIsSidebarOpen(shouldBeOpen);
    };

    // Add resize listener
    window.addEventListener("resize", handleResize);

    // Cleanup
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  return (
    <SidebarProvider open={isSidebarOpen} onOpenChange={setIsSidebarOpen}>
      <CompanyInitializer />
      <AppSidebar />
      <SidebarInset>{children}</SidebarInset>
    </SidebarProvider>
  );
}
