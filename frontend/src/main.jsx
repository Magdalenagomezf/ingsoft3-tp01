import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import * as Sentry from '@sentry/react'
import './index.css'
import App from './App.jsx'

Sentry.init({
  dsn: 'https://1cb6d8741346fe2db14973fdb2db751b@o4512301876838400.ingest.us.sentry.io/4512301899513856',
  environment: window.location.hostname, // de qué sitio vino: QA, PROD o localhost
  integrations: [Sentry.browserTracingIntegration()],
  tracesSampleRate: 1.0, // medir todas las cargas y pedidos
})

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <App />
  </StrictMode>,
)