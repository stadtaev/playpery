import { useState, useEffect } from "react"
import { AnimatePresence, motion } from "framer-motion"
import { Building2, Map, Gamepad2, LogOut, Menu, Compass } from "lucide-react"
import { getMe, logout } from "./adminApi"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { navigate } from "@/lib/navigate"

interface NavItem {
  label: string
  path: string
  icon: React.ReactNode
}

function getNavItems(client?: string): NavItem[] {
  const items: NavItem[] = [
    { label: "Clients", path: "/admin/clients", icon: <Building2 size={18} /> },
    { label: "Scenarios", path: "/admin/scenarios", icon: <Map size={18} /> },
  ]
  if (client) {
    items.push({
      label: "Games",
      path: `/admin/clients/${client}/games`,
      icon: <Gamepad2 size={18} />,
    })
  }
  return items
}

function isActive(path: string): boolean {
  const current = window.location.pathname
  if (path === "/admin/clients") {
    return current === "/admin" || current === "/admin/clients"
  }
  return current.startsWith(path)
}

function SidebarContent({
  client,
  onLogout,
  onNavigate,
}: {
  client?: string
  onLogout: () => void
  onNavigate?: () => void
}) {
  const items = getNavItems(client)

  function handleNav(e: React.MouseEvent, path: string) {
    e.preventDefault()
    navigate(path)
    onNavigate?.()
  }

  return (
    <div className="flex flex-col h-full">
      {/* Logo */}
      <div className="p-5 border-b border-border">
        <a
          href="/admin/clients"
          onClick={(e) => handleNav(e, "/admin/clients")}
          className="flex items-center gap-2 text-text-primary no-underline"
        >
          <Compass size={22} className="text-accent" />
          <span className="font-semibold text-base">CityQuest</span>
        </a>
      </div>

      {/* Nav links */}
      <nav className="flex-1 py-3 px-2">
        {client && (
          <div className="px-3 py-2 mb-1">
            <span className="text-xs font-medium text-text-muted uppercase tracking-wider">
              {client}
            </span>
          </div>
        )}
        {items.map((item) => {
          const active = isActive(item.path)
          return (
            <a
              key={item.path}
              href={item.path}
              onClick={(e) => handleNav(e, item.path)}
              className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm no-underline transition-colors mb-0.5 ${
                active
                  ? "bg-accent/10 text-accent border-l-2 border-accent -ml-0.5 pl-[10px]"
                  : "text-text-secondary hover:bg-card hover:text-text-primary"
              }`}
            >
              {item.icon}
              {item.label}
            </a>
          )
        })}
      </nav>

      {/* Logout */}
      <div className="p-3 border-t border-border">
        <Button
          variant="ghost"
          size="sm"
          onClick={onLogout}
          className="w-full justify-start gap-2 text-text-secondary"
        >
          <LogOut size={16} />
          Log out
        </Button>
      </div>
    </div>
  )
}

export function AdminLayout({
  client,
  children,
}: {
  client?: string
  children: React.ReactNode
}) {
  const [authed, setAuthed] = useState<boolean | null>(null)
  const [mobileOpen, setMobileOpen] = useState(false)

  useEffect(() => {
    getMe()
      .then(() => setAuthed(true))
      .catch(() => {
        setAuthed(false)
        navigate("/admin/login")
      })
  }, [])

  // Close mobile sidebar on route change
  useEffect(() => {
    function onNav() {
      setMobileOpen(false)
    }
    window.addEventListener("popstate", onNav)
    return () => window.removeEventListener("popstate", onNav)
  }, [])

  async function handleLogout() {
    await logout().catch(() => {})
    navigate("/admin/login")
  }

  if (authed === null) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Spinner size={32} />
      </div>
    )
  }

  if (!authed) return null

  return (
    <div className="flex min-h-screen">
      {/* Desktop sidebar */}
      <aside className="hidden lg:flex lg:flex-col lg:w-60 lg:fixed lg:inset-y-0 bg-card border-r border-border">
        <SidebarContent client={client} onLogout={handleLogout} />
      </aside>

      {/* Mobile sidebar */}
      <AnimatePresence>
        {mobileOpen && (
          <>
            {/* Backdrop */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 lg:hidden"
              onClick={() => setMobileOpen(false)}
            />
            {/* Drawer */}
            <motion.aside
              initial={{ x: "-100%" }}
              animate={{ x: 0 }}
              exit={{ x: "-100%" }}
              transition={{ type: "spring", damping: 25, stiffness: 300 }}
              className="fixed inset-y-0 left-0 w-60 bg-card border-r border-border z-50 lg:hidden"
            >
              <SidebarContent
                client={client}
                onLogout={handleLogout}
                onNavigate={() => setMobileOpen(false)}
              />
            </motion.aside>
          </>
        )}
      </AnimatePresence>

      {/* Main content */}
      <div className="flex-1 lg:pl-60">
        {/* Mobile top bar */}
        <header className="lg:hidden sticky top-0 z-30 flex items-center justify-between h-14 px-4 bg-card/80 backdrop-blur border-b border-border">
          <button
            onClick={() => setMobileOpen(true)}
            className="p-2 text-text-secondary hover:text-text-primary transition-colors"
            aria-label="Open navigation menu"
          >
            <Menu size={20} />
          </button>
          <div className="flex items-center gap-2">
            <Compass size={18} className="text-accent" />
            <span className="font-semibold text-sm text-text-primary">CityQuest</span>
          </div>
          <div className="w-9" /> {/* Spacer for centering */}
        </header>

        {/* Page content */}
        <main className="p-6 max-w-5xl mx-auto">{children}</main>
      </div>
    </div>
  )
}
