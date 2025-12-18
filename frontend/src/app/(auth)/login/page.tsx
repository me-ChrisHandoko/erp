"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useLoginMutation } from "@/store/services/authApi";
import type { ApiErrorResponse } from "@/types/api";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Mail, Lock, Eye, EyeOff, Package, Loader2, AlertCircle } from "lucide-react";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(false);
  const [isNavigating, setIsNavigating] = useState(false);
  const [login, { isLoading, error }] = useLoginMutation();
  const router = useRouter();

  // Restore email from localStorage on component mount
  // This is intentional initialization from localStorage on mount, not a cascading side effect
  useEffect(() => {
    if (typeof window !== "undefined") {
      const savedEmail = localStorage.getItem("rememberEmail");
      if (savedEmail) {
        setEmail(savedEmail);
        setRememberMe(true);
      }
    }
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      // Call login mutation
      await login({ email, password }).unwrap();

      // Set navigating state to keep spinner visible during navigation
      setIsNavigating(true);

      // Handle remember me preference
      if (typeof window !== "undefined") {
        if (rememberMe) {
          // Save email if "Ingat saya" is checked
          localStorage.setItem("rememberEmail", email);
        } else {
          // Clear saved email if "Ingat saya" is unchecked
          localStorage.removeItem("rememberEmail");
        }
      }

      // Redirect to dashboard on success
      router.push("/dashboard");
      // isNavigating will auto-reset when component unmounts
    } catch (err) {
      // Reset navigating state on error
      setIsNavigating(false);
      // Error is handled by RTK Query and displayed below
      console.error("Login failed:", err);
    }
  };

  // Extract error message from RTK Query error
  const getErrorMessage = () => {
    if (!error) return null;

    // Type guard for FetchBaseQueryError
    if ("status" in error) {
      // Backend error response
      if (error.data && typeof error.data === "object") {
        const apiError = error.data as ApiErrorResponse;
        return apiError.error?.message || "Login gagal. Silakan coba lagi.";
      }
      return "Login gagal. Silakan coba lagi.";
    }

    // SerializedError
    return error.message || "Terjadi kesalahan. Silakan coba lagi.";
  };

  return (
    <div className="w-full space-y-6">
      {/* Logo/Branding Section */}
      <div className="flex flex-col items-center space-y-2 text-center">
        <div className="flex items-center justify-center w-16 h-16 rounded-xl bg-primary/10 ring-1 ring-primary/20">
          <Package className="w-8 h-8 text-primary" />
        </div>
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight">ERP Distribusi</h1>
          <p className="text-sm text-muted-foreground">
            Sistem Manajemen Distribusi Sembako
          </p>
        </div>
      </div>

      {/* Login Card */}
      <Card className="border-border/50 shadow-xl">
        <form onSubmit={handleSubmit}>
          <CardHeader className="space-y-1 pb-4">
            <CardTitle className="text-2xl font-semibold">Masuk ke Akun</CardTitle>
            <CardDescription className="text-base">
              Masukkan kredensial Anda untuk mengakses sistem
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Error Alert */}
            {error && (
              <Alert variant="destructive" className="animate-in fade-in-50 duration-300">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{getErrorMessage()}</AlertDescription>
              </Alert>
            )}

            {/* Email Field */}
            <div className="space-y-2">
              <Label htmlFor="email" className="text-sm font-medium">
                Alamat Email
              </Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  placeholder="nama@perusahaan.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  disabled={isLoading || isNavigating}
                  required
                  className="pl-10 h-11 transition-all duration-200 focus:ring-2 focus:ring-primary/20"
                />
              </div>
            </div>

            {/* Password Field */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password" className="text-sm font-medium">
                  Kata Sandi
                </Label>
                <button
                  type="button"
                  className="text-sm font-medium text-primary hover:underline underline-offset-4 transition-colors"
                  onClick={() => {
                    // TODO: Implement forgot password functionality
                    console.log("Forgot password clicked");
                  }}
                >
                  Lupa password?
                </button>
              </div>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  placeholder="••••••••"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  disabled={isLoading || isNavigating}
                  required
                  className="pl-10 pr-10 h-11 transition-all duration-200 focus:ring-2 focus:ring-primary/20"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
                  disabled={isLoading || isNavigating}
                  tabIndex={-1}
                >
                  {showPassword ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </button>
              </div>
            </div>

            {/* Remember Me */}
            <div className="flex items-center space-x-2">
              <Checkbox
                id="remember"
                checked={rememberMe}
                onCheckedChange={(checked) => {
                  setRememberMe(checked as boolean);
                }}
                disabled={isLoading || isNavigating}
              />
              <label
                htmlFor="remember"
                className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
              >
                Ingat saya
              </label>
            </div>
          </CardContent>

          <CardFooter className="flex flex-col space-y-4 pt-2">
            {/* Login Button */}
            <Button
              className="w-full h-11 text-base font-semibold shadow-md hover:shadow-lg transition-all duration-200"
              size="lg"
              type="submit"
              disabled={isLoading || isNavigating}
            >
              {isLoading || isNavigating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Memproses...
                </>
              ) : (
                "Masuk"
              )}
            </Button>

            {/* Contact Admin Link */}
            <div className="text-center">
              <p className="text-sm text-muted-foreground">
                Belum memiliki akses?{" "}
                <button
                  type="button"
                  className="font-medium text-primary hover:underline underline-offset-4 transition-colors"
                  onClick={() => {
                    // TODO: Implement contact admin functionality
                    console.log("Contact admin clicked");
                  }}
                >
                  Hubungi administrator
                </button>
              </p>
            </div>
          </CardFooter>
        </form>
      </Card>

      {/* Security Badge */}
      <div className="flex items-center justify-center text-xs text-muted-foreground">
        <Lock className="h-3 w-3 mr-1" />
        <span>Koneksi aman dan terenkripsi</span>
      </div>
    </div>
  );
}
