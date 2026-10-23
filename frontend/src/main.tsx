import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { App } from '@/App.tsx'
import '@/styles/tokens.css'
import '@/styles/base.css'

const rootElement = document.getElementById('root')

if (!rootElement) {
  throw new Error('Không tìm thấy phần tử #root trong index.html')
}

createRoot(rootElement).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
