import { Book, Download, History, Search, Settings, type LucideIcon } from 'lucide-react'
import type { ReactElement } from 'react'
import { NavLink, Outlet } from 'react-router'
import { AI_SETTINGS_PATH, DOWNLOADS_PATH, HISTORY_PATH, LIBRARY_PATH, SEARCH_PATH } from './paths'

interface NavItem {
  readonly to: string
  readonly label: string
  readonly icon: LucideIcon
}

const NAV_ITEMS: readonly NavItem[] = [
  { to: LIBRARY_PATH, label: 'Thư viện', icon: Book },
  { to: SEARCH_PATH, label: 'Tìm phụ đề', icon: Search },
  { to: DOWNLOADS_PATH, label: 'Tải khoá học', icon: Download },
  { to: HISTORY_PATH, label: 'Lịch sử', icon: History },
  { to: AI_SETTINGS_PATH, label: 'Cài đặt AI', icon: Settings },
]

const SIDEBAR_LINK = 'flex h-[46px] items-center gap-[13px] rounded-[10px] px-[15px] text-sm'
const TAB_LINK = 'grid justify-items-center gap-1 py-2 text-[11px]'

export function AppShell(): ReactElement {
  return (
    <div className="min-h-dvh md:grid md:grid-cols-[232px_minmax(0,1fr)]">
      <aside className="hidden border-r border-sidebar-edge bg-sidebar md:sticky md:top-0 md:flex md:h-dvh md:flex-col md:gap-12 md:rounded-r-2xl md:px-5 md:py-8">
        <span className="px-1.5 text-[21px] font-semibold">udemy-studio</span>
        <nav aria-label="Điều hướng chính">
          <ul className="m-0 grid list-none gap-2.5 p-0">
            {NAV_ITEMS.map((item) => (
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
      <div className="pb-20 md:pb-0">
        <Outlet />
      </div>
      <nav
        aria-label="Điều hướng chính"
        className="fixed inset-x-0 bottom-0 border-t border-line bg-sidebar md:hidden"
      >
        <ul className="m-0 grid list-none grid-cols-5 p-0">
          {NAV_ITEMS.map((item) => (
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
