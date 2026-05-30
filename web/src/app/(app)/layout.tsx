// Auth guard + org context provider — required by all authenticated pages.
// TODO: Implement session check and redirect to /login if unauthenticated (Story 2.x).
export default function AppLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}
