import { createBrowserRouter } from 'react-router'
import { CourseIndexRedirect } from '../features/course/CourseIndexRedirect'
import { CourseLayout } from '../features/course/CourseLayout'
import { LibraryPage } from '../features/library/LibraryPage'
import { LecturePage } from '../features/player/LecturePage'
import { NotFoundPage } from './NotFoundPage'

export const router = createBrowserRouter([
  { path: '/', element: <LibraryPage /> },
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
