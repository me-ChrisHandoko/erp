"use client"

import { useEffect, useRef } from "react"
import { useSelector } from "react-redux"
import { useInitializeCompanyContextMutation } from "@/store/services/multiCompanyApi"
import type { RootState } from "@/store"

/**
 * CompanyInitializer
 *
 * Client component that initializes company context after user login.
 * This component:
 * 1. Waits for auth to be ready (token restored from localStorage)
 * 2. Fetches available companies from API
 * 3. Checks localStorage for previously active company
 * 4. Switches to target company (stored or first active)
 * 5. Updates Redux state with activeCompany and availableCompanies
 *
 * This makes TeamSwitcher and company menu visible in the sidebar.
 *
 * IMPORTANT: Waits for auth to be ready before initializing to prevent 401 errors on hard refresh
 */
export function CompanyInitializer() {
  const [initializeCompanyContext, { isLoading, isError, error }] =
    useInitializeCompanyContextMutation()

  // Wait for auth to be ready before initializing companies
  const isAuthenticated = useSelector((state: RootState) => state.auth.isAuthenticated)
  const accessToken = useSelector((state: RootState) => state.auth.accessToken)

  // Prevent double initialization in React StrictMode
  const initialized = useRef(false)

  useEffect(() => {
    // CRITICAL: Reset initialized flag when user logs out
    // This ensures new user gets fresh company data
    if (!isAuthenticated || !accessToken) {
      if (initialized.current) {
        console.log("🔄 CompanyInitializer: User logged out, resetting initialization flag...")
        initialized.current = false
      }
      console.log("⏳ CompanyInitializer: Waiting for auth to be ready...", {
        isAuthenticated,
        hasToken: !!accessToken,
      })
      return
    }

    // Skip if already initialized (React StrictMode calls useEffect twice)
    if (initialized.current) {
      console.log("⏭️  CompanyInitializer: Already initialized, skipping...")
      return
    }

    // Initialize company context once auth is ready
    const initialize = async () => {
      try {
        console.log("🔄 CompanyInitializer: Auth ready, starting company context initialization...")
        initialized.current = true
        await initializeCompanyContext().unwrap()
        console.log("✅ CompanyInitializer: Company context initialized successfully")
      } catch (err) {
        console.error("❌ CompanyInitializer: Failed to initialize company context:", err)
        initialized.current = false // Reset on error to allow retry
      }
    }

    initialize()
  }, [initializeCompanyContext, isAuthenticated, accessToken])

  // This component doesn't render anything visible
  // It just handles the initialization logic
  if (isLoading) {
    console.log("⏳ CompanyInitializer: Loading company context...")
  }

  if (isError) {
    console.error("❌ CompanyInitializer: Error loading company context:", error)
  }

  return null
}
