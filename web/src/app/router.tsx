import { createBrowserRouter } from 'react-router'
import { CourseIndexRedirect } from '../features/course/CourseIndexRedirect'
import { CourseLayout } from '../features/course/CourseLayout'
import { DownloadsPage } from '../features/downloads/DownloadsPage'
import { LibraryPage } from '../features/library/LibraryPage'
import { LecturePage } from '../features/player/LecturePage'
import { SearchPage } from '../features/search/SearchPage'
import { AISettingsPage } from '../features/ai-settings/AISettingsPage'
import { AppShell } from './AppShell'
import { NotFoundPage } from './NotFoundPage'
import { AI_SETTINGS_PATH, DOWNLOADS_PATH, LIBRARY_PATH, SEARCH_PATH } from './paths'

export const router = createBrowserRouter([
  {
    element: <AppShell />,
    children: [
      { path: LIBRARY_PATH, element: <LibraryPage /> },
      { path: DOWNLOADS_PATH, element: <DownloadsPage /> },
      { path: SEARCH_PATH, element: <SearchPage /> },
      { path: AI_SETTINGS_PATH, element: <AISettingsPage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
  {
    path: '/courses/:courseId',
    element: <CourseLayout />,
    children: [
      { index: true, element: <CourseIndexRedirect /> },
      { path: 'lectures/:lectureId', element: <LecturePage /> },
    ],
  },
])
