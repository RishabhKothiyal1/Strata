import { lazy, Suspense } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider, RequireAuth } from './context/AuthContext';
import { Login } from './pages/Login';
import { Signup } from './pages/Signup';

// Lazy load layout and pages for performance
const DashboardLayout = lazy(() =>
  import('./components/DashboardLayout').then((m) => ({ default: m.DashboardLayout }))
);
const DashboardOverview = lazy(() =>
  import('./pages/DashboardOverview').then((m) => ({ default: m.DashboardOverview }))
);
const DataRecords = lazy(() =>
  import('./pages/DataRecords').then((m) => ({ default: m.DataRecords }))
);
const Storage = lazy(() =>
  import('./pages/Storage').then((m) => ({ default: m.Storage }))
);
const Functions = lazy(() =>
  import('./pages/Functions').then((m) => ({ default: m.Functions }))
);
const Settings = lazy(() =>
  import('./pages/Settings').then((m) => ({ default: m.Settings }))
);
const Support = lazy(() =>
  import('./pages/Support').then((m) => ({ default: m.Support }))
);
const ApiKeys = lazy(() =>
  import('./pages/ApiKeys').then((m) => ({ default: m.ApiKeys }))
);

const Home = lazy(() =>
  import('./pages/Home').then((m) => ({ default: m.Home }))
);

// Initialize React Query Client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<Login />} />
            <Route path="/signup" element={<Signup />} />

            {/* Public Landing Page */}
            <Route
              path="/"
              element={
                <Suspense
                  fallback={
                    <div className="min-h-screen bg-[#030303] flex items-center justify-center">
                      <div className="w-8 h-8 border-4 border-[#f26500] border-t-transparent rounded-full animate-spin" />
                    </div>
                  }
                >
                  <Home />
                </Suspense>
              }
            />

            {/* Protected Nested Routes under DashboardLayout */}
            <Route
              element={
                <RequireAuth>
                  <Suspense
                    fallback={
                      <div className="min-h-screen flex items-center justify-center bg-[#f8f9fa]">
                        <div className="flex flex-col items-center space-y-md">
                          <div className="w-10 h-10 border-4 border-[#4648d4] border-t-transparent rounded-full animate-spin" />
                          <p className="font-body-md text-[#464554] animate-pulse">
                            Loading section...
                          </p>
                        </div>
                      </div>
                    }
                  >
                    <DashboardLayout />
                  </Suspense>
                </RequireAuth>
              }
            >
              <Route path="/dashboard" element={<DashboardOverview />} />
              <Route path="/data" element={<DataRecords />} />
              <Route path="/storage" element={<Storage />} />
              <Route path="/functions" element={<Functions />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="/support" element={<Support />} />
              <Route path="/api-keys" element={<ApiKeys />} />
            </Route>

            {/* Redirect fallback */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;
