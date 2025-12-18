import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"

export default function DashboardPage() {
  return (
    <>
      <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
        <div className="flex items-center gap-2 px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Dashboard</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </header>
      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        <h1 className="text-2xl font-bold">Dashboard ERP</h1>
        <div className="grid auto-rows-min gap-4 md:grid-cols-3">
          <div className="aspect-video rounded-xl bg-muted/50 p-4">
            <h3 className="font-semibold">Total Penjualan</h3>
            <p className="text-muted-foreground text-sm mt-2">
              Data penjualan akan ditampilkan di sini
            </p>
          </div>
          <div className="aspect-video rounded-xl bg-muted/50 p-4">
            <h3 className="font-semibold">Stok Menipis</h3>
            <p className="text-muted-foreground text-sm mt-2">
              Peringatan stok akan ditampilkan di sini
            </p>
          </div>
          <div className="aspect-video rounded-xl bg-muted/50 p-4">
            <h3 className="font-semibold">Hutang/Piutang</h3>
            <p className="text-muted-foreground text-sm mt-2">
              Informasi keuangan akan ditampilkan di sini
            </p>
          </div>
        </div>
        <div className="min-h-screen flex-1 rounded-xl bg-muted/50 p-4 md:min-h-min">
          <p className="text-muted-foreground">
            Konten dashboard akan ditampilkan di sini.
          </p>
        </div>
      </div>
    </>
  )
}
