import type { NextConfig } from "next";

const isDevelopment = process.env.NODE_ENV === "development";
const isProduction = process.env.NODE_ENV === "production";

const nextConfig: NextConfig = {
  // ========================================
  // Image Optimization Configuration
  // ========================================
  images: {
    remotePatterns: [
      // Development: Support both HTTP (current) and HTTPS (future)
      {
        protocol: "http",
        hostname: "localhost",
        port: "8080",
        pathname: "/uploads/**",
      },
      {
        protocol: "https",
        hostname: "localhost",
        port: "8080",
        pathname: "/uploads/**",
      },
      // Production: Only HTTPS
      {
        protocol: "https",
        hostname: "**",
      },
    ],
    // Optimize images with modern formats
    formats: ["image/avif", "image/webp"],
  },

  // ========================================
  // Security Headers (Phase 2: HSTS Ready)
  // ========================================
  async headers() {
    const headers: Array<{
      source: string;
      headers: Array<{ key: string; value: string }>;
    }> = [];

    // Apply security headers to all routes
    if (isProduction) {
      headers.push({
        source: "/:path*",
        headers: [
          // HSTS: Force HTTPS (only in production)
          {
            key: "Strict-Transport-Security",
            value: "max-age=31536000; includeSubDomains",
          },
          // Upgrade HTTP requests to HTTPS
          {
            key: "Content-Security-Policy",
            value: "upgrade-insecure-requests",
          },
        ],
      });
    }

    return headers;
  },

  // ========================================
  // Production Optimizations
  // ========================================
  ...(isProduction && {
    // Compress responses
    compress: true,
    // Generate build ID for cache busting
    generateBuildId: async () => {
      return `build-${Date.now()}`;
    },
  }),

  // ========================================
  // Development Configuration
  // ========================================
  ...(isDevelopment && {
    // Disable static optimization for faster refresh
    reactStrictMode: true,
  }),
};

export default nextConfig;
