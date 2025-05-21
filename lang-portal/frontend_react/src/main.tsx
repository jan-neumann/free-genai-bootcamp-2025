import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import './index.css';
import App from './App.tsx';

// Start the mock API server in development
async function enableMocking() {
  if (import.meta.env.DEV) {
    if (window.location.pathname === '/lang-portal/frontend_react') {
      window.location.pathname = '/';
      return;
    }
    
    const { worker } = await import('./mocks/browser');
    await worker.start({
      onUnhandledRequest: 'bypass',
      serviceWorker: {
        url: '/mockServiceWorker.js',
      },
    });
  }
  return Promise.resolve();
}

// Initialize the app after the mock service worker is started
const initializeApp = async () => {
  try {
    await enableMocking();
    
    const root = createRoot(document.getElementById('root')!);
    
    root.render(
      <StrictMode>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </StrictMode>
    );
  } catch (error) {
    console.error('Failed to initialize the app:', error);
  }
};

initializeApp();
