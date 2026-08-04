import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AuthProvider, useAuth } from "./hooks/useAuth";
import Layout from "./components/Layout";
import LoginPage from "./pages/login/LoginPage";
import ServersPage from "./pages/servers/ServersPage";
import ServerDetailPage from "./pages/servers/ServerDetailPage";
import ToolDetailPage from "./pages/tools/ToolDetailPage";
import InvocationsPage from "./pages/invocations/InvocationsPage";
import AdminServersPage from "./pages/admin/AdminServersPage";
import ServerFormPage from "./pages/admin/ServerFormPage";
import ToolFormPage from "./pages/admin/ToolFormPage";

const queryClient = new QueryClient();

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={<Navigate to="/servers" replace />} />
              <Route path="servers" element={<ServersPage />} />
              <Route path="servers/:serverId" element={<ServerDetailPage />} />
              <Route
                path="servers/:serverId/tools/:toolId"
                element={<ToolDetailPage />}
              />
              <Route path="invocations" element={<InvocationsPage />} />

              {/* ✅ CORRECT: Admin routes as siblings */}
              <Route path="admin/servers" element={<AdminServersPage />} />
              <Route path="admin/servers/new" element={<ServerFormPage />} />
              <Route
                path="admin/servers/:serverId/edit"
                element={<ServerFormPage />}
              />
              <Route
                path="admin/servers/:serverId/tools/new"
                element={<ToolFormPage />}
              />
              <Route
                path="admin/servers/:serverId/tools/:toolId/edit"
                element={<ToolFormPage />}
              />
            </Route>
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;
