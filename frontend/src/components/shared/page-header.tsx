"use client";

/**
 * PageHeader Component
 *
 * Reusable header component for app pages with:
 * - Sidebar trigger
 * - Breadcrumb navigation (single or multi-level)
 * - Responsive design
 * - Optional border styling
 *
 * @example
 * // Simple single-level breadcrumb
 * <PageHeader breadcrumbs={[{ label: "Dashboard" }]} />
 *
 * @example
 * // Multi-level breadcrumb with links
 * <PageHeader
 *   breadcrumbs={[
 *     { label: "Dashboard", href: "/dashboard" },
 *     { label: "Perusahaan", href: "/company" },
 *     { label: "Profil" }
 *   ]}
 *   showBorder
 * />
 */

import { Fragment } from "react";
import { cn } from "@/lib/utils";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";

export interface BreadcrumbItemType {
  /** Display label for the breadcrumb item */
  label: string;
  /** Optional href - if provided, renders as link. Last item should not have href. */
  href?: string;
}

interface PageHeaderProps {
  /** Array of breadcrumb items. Last item is typically the current page (no href). */
  breadcrumbs: BreadcrumbItemType[];
  /** Whether to show bottom border. Useful for visual separation. */
  showBorder?: boolean;
  /** Additional CSS classes for customization */
  className?: string;
}

export function PageHeader({
  breadcrumbs,
  showBorder = false,
  className,
}: PageHeaderProps) {
  return (
    <header
      className={cn(
        "flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12",
        showBorder && "border-b",
        className
      )}
    >
      <div className="flex items-center gap-2 px-4">
        <SidebarTrigger className="-ml-1" />
        <Separator
          orientation="vertical"
          className="mr-2 data-[orientation=vertical]:h-4"
        />
        <Breadcrumb>
          <BreadcrumbList>
            {breadcrumbs.map((item, index) => (
              <Fragment key={`${item.label}-${index}`}>
                {index > 0 && (
                  <BreadcrumbSeparator className="hidden md:block" />
                )}
                <BreadcrumbItem className={index > 0 ? "hidden md:block" : ""}>
                  {item.href ? (
                    <BreadcrumbLink href={item.href}>
                      {item.label}
                    </BreadcrumbLink>
                  ) : (
                    <BreadcrumbPage>{item.label}</BreadcrumbPage>
                  )}
                </BreadcrumbItem>
              </Fragment>
            ))}
          </BreadcrumbList>
        </Breadcrumb>
      </div>
    </header>
  );
}
