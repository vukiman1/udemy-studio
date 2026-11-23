import { QueryClientProvider } from '@tanstack/react-query'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router'
import { createQueryClient } from './app/query-client'
import { router } from './app/router'
import '@videojs/react/i18n/locales/vi/register'
import './index.css'

const ROOT_ELEMENT_ID = 'root'

class MissingRootElementError extends Error {
  constructor(elementId: string) {
    super(`index.html has no element with id "${elementId}"`)
    this.name = 'MissingRootElementError'
  }
}

const rootElement = document.getElementById(ROOT_ELEMENT_ID)
if (rootElement === null) {
  throw new MissingRootElementError(ROOT_ELEMENT_ID)
}

createRoot(rootElement).render(
  <StrictMode>
    <QueryClientProvider client={createQueryClient()}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
