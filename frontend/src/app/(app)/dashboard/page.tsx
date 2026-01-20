import { PageHeader } from "@/components/shared/page-header";
import { DashboardClient } from "./dashboard-client";

export default function DashboardPage() {
  return (
    <>
      <PageHeader breadcrumbs={[{ label: "Dashboard" }]} />
      <DashboardClient />
    </>
  );
}
