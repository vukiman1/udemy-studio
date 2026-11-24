import { createBrowserRouter } from 'react-router'
import { CourseIndexRedirect } from '../features/course/CourseIndexRedirect'
import { CourseLayout } from '../features/course/CourseLayout'
import { DownloadsPage } from '../features/downloads/DownloadsPage'
import { LibraryPage } from '../features/library/LibraryPage'
import { LecturePage } from '../features/player/LecturePage'
import { SearchPage } from '../features/search/SearchPage'
import { NotFoundPage } from './NotFoundPage'
import { SEARCH_PATH } from './paths'

export const router = createBrowserRouter([
  { path: '/', element: <LibraryPage /> },
  { path: '/downloads', element: <DownloadsPage /> },
  { path: SEARCH_PATH, element: <SearchPage /> },
  {
    path: '/courses/:courseId',
    element: <CourseLayout />,
    children: [
      { index: true, element: <CourseIndexRedirect /> },
      { path: 'lectures/:lectureId', element: <LecturePage /> },
    ],
  },
  { path: '*', element: <NotFoundPage /> },
])
