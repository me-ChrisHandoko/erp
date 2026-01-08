"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { selectLogoutReason, clearLogoutReason } from "@/store/slices/authSlice";
import { useLogoutMutation } from "@/store/services/authApi";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle, LogOut, Loader2 } from "lucide-react";

export default function LogoutPage() {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const logoutReason = useAppSelector(selectLogoutReason);
  const [logout, { isLoading }] = useLogoutMutation();
  const [countdown, setCountdown] = useState(5);
  const [isRedirecting, setIsRedirecting] = useState(false);

  // Call logout API to clear server-side session
  useEffect(() => {
    const performLogout = async () => {
      try {
        await logout().unwrap();
        console.log('[Logout] Server-side logout successful');
      } catch (error) {
        console.error('[Logout] Server-side logout failed:', error);
        // Continue anyway - client-side state already cleared
      }
    };

    performLogout();
  }, [logout]);

  // Auto-redirect countdown
  useEffect(() => {
    if (countdown === 0) {
      setIsRedirecting(true);
      // Clear logout reason before redirect
      dispatch(clearLogoutReason());
      router.push('/login');
      return;
    }

    const timer = setTimeout(() => {
      setCountdown(countdown - 1);
    }, 1000);

    return () => clearTimeout(timer);
  }, [countdown, router, dispatch]);

  const handleLoginNow = () => {
    setIsRedirecting(true);
    // Clear logout reason before redirect
    dispatch(clearLogoutReason());
    router.push('/login');
  };

  // Determine logout message based on reason
  const getLogoutMessage = () => {
    if (logoutReason === 'session_expired') {
      return {
        title: 'Sesi Anda Telah Berakhir',
        description: 'Sesi login Anda telah berakhir setelah 7 hari. Silahkan login kembali untuk melanjutkan.',
        icon: <AlertCircle className="h-5 w-5" />,
      };
    }

    // Default message for manual logout
    return {
      title: 'Anda Telah Logout',
      description: 'Anda telah berhasil keluar dari sistem. Terima kasih telah menggunakan aplikasi.',
      icon: <LogOut className="h-5 w-5" />,
    };
  };

  const message = getLogoutMessage();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardHeader className="text-center space-y-2">
          <div className="mx-auto w-12 h-12 bg-orange-100 rounded-full flex items-center justify-center mb-2">
            {message.icon}
          </div>
          <CardTitle className="text-2xl font-bold text-gray-900">
            {message.title}
          </CardTitle>
          <CardDescription className="text-base text-gray-600">
            {message.description}
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* Show session expired alert for session_expired reason */}
          {logoutReason === 'session_expired' && (
            <Alert variant="default" className="bg-amber-50 border-amber-200">
              <AlertCircle className="h-4 w-4 text-amber-600" />
              <AlertDescription className="text-sm text-amber-800">
                Untuk keamanan, sesi login otomatis berakhir setelah 7 hari tanpa aktivitas.
              </AlertDescription>
            </Alert>
          )}

          {/* Countdown info */}
          <div className="text-center text-sm text-gray-500">
            {isRedirecting ? (
              <div className="flex items-center justify-center gap-2">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>Mengalihkan ke halaman login...</span>
              </div>
            ) : (
              <span>
                Anda akan dialihkan ke halaman login dalam{" "}
                <span className="font-semibold text-blue-600">{countdown}</span>{" "}
                detik
              </span>
            )}
          </div>
        </CardContent>

        <CardFooter className="flex flex-col gap-3">
          <Button
            onClick={handleLoginNow}
            disabled={isRedirecting || isLoading}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white"
            size="lg"
          >
            {isRedirecting || isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Mengalihkan...
              </>
            ) : (
              <>
                <LogOut className="mr-2 h-4 w-4" />
                Login Sekarang
              </>
            )}
          </Button>

          <p className="text-xs text-center text-gray-500">
            Sistem akan otomatis mengalihkan Anda ke halaman login
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}
