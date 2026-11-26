import { Book, Cloud, Download, History, Search, Settings, type LucideIcon } from 'lucide-react'
import type { ReactElement } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router'
import { useIsLibraryCopy } from '../api/server-queries'
import {
  AI_SETTINGS_PATH,
  DOWNLOADS_PATH,
  HISTORY_PATH,
  LIBRARY_PATH,
  SEARCH_PATH,
  STORAGE_SETTINGS_PATH,
} from './paths'

interface NavItem {
  readonly to: string
  readonly label: string
  readonly icon: LucideIcon
  readonly isLibraryOnly?: boolean
}

const NAV_ITEMS: readonly NavItem[] = [
  { to: LIBRARY_PATH, label: 'Library', icon: Book },
  { to: SEARCH_PATH, label: 'Search', icon: Search },
  { to: DOWNLOADS_PATH, label: 'Downloads', icon: Download, isLibraryOnly: true },
  { to: HISTORY_PATH, label: 'History', icon: History },
  { to: AI_SETTINGS_PATH, label: 'AI settings', icon: Settings, isLibraryOnly: true },
  { to: STORAGE_SETTINGS_PATH, label: 'Storage', icon: Cloud },
]

const SIDEBAR_LINK =
  'flex h-[46px] items-center gap-[13px] rounded-[10px] px-[15px] text-sm transition-[translate,background-color] duration-200 ease-soft motion-safe:hover:translate-x-1'
const TAB_LINK = 'grid justify-items-center gap-1 py-2 text-[11px]'

export function AppShell(): ReactElement {
  const { pathname } = useLocation()
  const isLibraryCopy = useIsLibraryCopy()
  const navItems = NAV_ITEMS.filter((item) => isLibraryCopy || item.isLibraryOnly !== true)
  return (
    <div className="min-h-dvh md:grid md:grid-cols-[232px_minmax(0,1fr)]">
      <aside className="hidden border-r border-sidebar-edge bg-sidebar md:sticky md:top-0 md:flex md:h-dvh md:flex-col md:gap-12 md:rounded-r-2xl md:px-5 md:py-8">
        <span className="px-1.5 text-[21px] font-semibold">udemy-studio</span>
        <nav aria-label="Main navigation">
          <ul className="m-0 grid list-none gap-2.5 p-0">
            {navItems.map((item) => (
              <li key={item.to}>
                <NavLink
                  to={item.to}
                  end={item.to === LIBRARY_PATH}
                  className={({ isActive }) =>
                    `${SIDEBAR_LINK} ${isActive ? 'bg-selected font-semibold text-ink' : 'text-muted hover:bg-white/6'}`
                  }
                >
                  <item.icon className="size-[22px] text-accent-soft" aria-hidden="true" />
                  {item.label}
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>
      </aside>
      <div key={pathname} className="pb-20 motion-safe:animate-page-enter md:pb-0">
        <Outlet />
      </div>
      <nav
        aria-label="Main navigation"
        className="fixed inset-x-0 bottom-0 border-t border-line bg-sidebar md:hidden"
      >
        <ul className="m-0 grid list-none auto-cols-fr grid-flow-col p-0">
          {navItems.map((item) => (
            <li key={item.to}>
              <NavLink
                to={item.to}
                end={item.to === LIBRARY_PATH}
                className={({ isActive }) =>
                  `${TAB_LINK} ${isActive ? 'font-semibold text-ink' : 'text-muted'}`
                }
              >
                <item.icon className="size-5 text-accent-soft" aria-hidden="true" />
                {item.label}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>
    </div>
  )
}
