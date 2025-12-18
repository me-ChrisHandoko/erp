export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="relative min-h-screen flex flex-col">
      {/* Background with gradient and pattern */}
      <div className="absolute inset-0 bg-gradient-to-br from-background via-muted/30 to-background">
        {/* Subtle grid pattern */}
        <div
          className="absolute inset-0 opacity-[0.02]"
          style={{
            backgroundImage: `linear-gradient(to right, hsl(var(--foreground)) 1px, transparent 1px),
                             linear-gradient(to bottom, hsl(var(--foreground)) 1px, transparent 1px)`,
            backgroundSize: "4rem 4rem",
          }}
        />
        {/* Radial gradient overlay */}
        <div className="absolute inset-0 bg-gradient-radial from-transparent via-transparent to-background/50" />
      </div>

      {/* Content container */}
      <div className="relative flex-1 flex items-center justify-center p-4 sm:p-6 md:p-8">
        <div className="w-full max-w-md animate-in fade-in-50 slide-in-from-bottom-4 duration-700">
          {children}
        </div>
      </div>

      {/* Footer */}
      <footer className="relative border-t border-border/40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container mx-auto px-4 py-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-2 text-xs text-muted-foreground">
            <p className="text-center sm:text-left">
              © {new Date().getFullYear()} ERP Distribusi. Hak cipta dilindungi.
            </p>
            <div className="flex items-center gap-4">
              <button
                type="button"
                className="hover:text-foreground transition-colors underline-offset-4 hover:underline"
              >
                Bantuan
              </button>
              <span className="text-border">|</span>
              <button
                type="button"
                className="hover:text-foreground transition-colors underline-offset-4 hover:underline"
              >
                Kebijakan Privasi
              </button>
              <span className="text-border">|</span>
              <button
                type="button"
                className="hover:text-foreground transition-colors underline-offset-4 hover:underline"
              >
                Syarat & Ketentuan
              </button>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
