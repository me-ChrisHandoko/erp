"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useLoginMutation } from "@/store/services/authApi";
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

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [login, { isLoading, error }] = useLoginMutation();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      // Call login mutation
      await login({ email, password }).unwrap();

      // Redirect to dashboard on success
      router.push("/dashboard");
    } catch (err) {
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
        const apiError = error.data as any;
        return apiError.error?.message || "Login gagal. Silakan coba lagi.";
      }
      return "Login gagal. Silakan coba lagi.";
    }

    // SerializedError
    return error.message || "Terjadi kesalahan. Silakan coba lagi.";
  };

  return (
    <Card>
      <form onSubmit={handleSubmit}>
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold">Login</CardTitle>
          <CardDescription>
            Masukkan email dan password untuk mengakses sistem ERP
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <div className="rounded-md bg-red-50 p-3 text-sm text-red-800 border border-red-200">
              {getErrorMessage()}
            </div>
          )}
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              placeholder="nama@perusahaan.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isLoading}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isLoading}
              required
            />
          </div>
        </CardContent>
        <CardFooter className="flex flex-col space-y-4">
          <Button className="w-full" size="lg" type="submit" disabled={isLoading}>
            {isLoading ? "Logging in..." : "Login"}
          </Button>
          <p className="text-center text-sm text-muted-foreground">
            Belum punya akun?{" "}
            <a href="#" className="underline hover:text-primary">
              Hubungi administrator
            </a>
          </p>
        </CardFooter>
      </form>
    </Card>
  );
}
