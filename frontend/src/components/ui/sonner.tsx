"use client"

import { useTheme } from "next-themes"
import { Toaster as Sonner, type ToasterProps } from "sonner"

const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "system" } = useTheme()

  return (
    <Sonner
      richColors
      theme={theme as ToasterProps["theme"]}
      className="toaster group"
      position="top-right"
      style={
        {
          "--normal-bg": "var(--background)",
          "--normal-border": "var(--border)",
          "--normal-text": "var(--foreground)",
          "--success-bg": "hsl(143 85% 96%)",
          "--success-border": "hsl(145 92% 91%)",
          "--success-text": "hsl(140 100% 27%)",
          "--error-bg": "hsl(0 93% 94%)",
          "--error-border": "hsl(0 93% 94%)",
          "--error-text": "hsl(0 74% 42%)",
          "--info-bg": "hsl(208 100% 97%)",
          "--info-border": "hsl(221 91% 91%)",
          "--info-text": "hsl(210 92% 45%)",
          "--warning-bg": "hsl(49 100% 97%)",
          "--warning-border": "hsl(49 91% 91%)",
          "--warning-text": "hsl(31 92% 45%)",
          "--border-radius": "var(--radius)",
        } as React.CSSProperties
      }
      {...props}
    />
  )
}

export { Toaster }
