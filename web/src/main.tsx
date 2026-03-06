import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)

// Uncomment to enable PWA install prompt:
// if ('serviceWorker' in navigator) {
//   navigator.serviceWorker.register('/sw.js')
// }
