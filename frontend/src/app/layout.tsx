import type { Metadata } from "next";
import localFont from "next/font/local";
import { Providers } from "@/components/providers";
import { Toaster } from "@/components/ui/sonner";
import "./globals.css";

const openSans = localFont({
  src: [
    {
      path: "../../public/fonts/OpenSans-Light.woff2",
      weight: "300",
      style: "normal",
    },
    {
      path: "../../public/fonts/OpenSans-Regular.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "../../public/fonts/OpenSans-Medium.woff2",
      weight: "500",
      style: "normal",
    },
    {
      path: "../../public/fonts/OpenSans-SemiBold.woff2",
      weight: "600",
      style: "normal",
    },
    {
      path: "../../public/fonts/OpenSans-Bold.woff2",
      weight: "700",
      style: "normal",
    },
  ],
  variable: "--font-open-sans",
  display: "swap",
});

export const metadata: Metadata = {
  title: "ERP Distribution - Distribusi Sembako",
  description: "Sistem ERP untuk distribusi sembako dan bahan pokok",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" suppressHydrationWarning>
      <body className={`${openSans.variable} font-sans antialiased`}>
        <Providers>{children}</Providers>
        <Toaster />
      </body>
    </html>
  );
}
