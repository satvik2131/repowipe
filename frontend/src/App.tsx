import { Toaster } from "@/components/ui/toaster";
import { TooltipProvider } from "@/components/ui/tooltip";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Index from "./pages/Index";
import Search from "./pages/Search";
import NotFound from "./pages/NotFound";
import HandleAuth from "./pages/HandleAuth";
import { ProtectedRoute } from "./routes/ProtectedRoute";
import { ErrorBoundary } from "./components/ErrorBoundary";

const App = () => (
  <ErrorBoundary>
    <TooltipProvider>
      <Toaster />
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Index />} />
          {/** Handles temporary code & state from github */}
          <Route path="/auth" element={<HandleAuth />} />
          <Route
            path="/search"
            element={
              <ProtectedRoute>
                <Search />
              </ProtectedRoute>
            }
          />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </BrowserRouter>
    </TooltipProvider>
  </ErrorBoundary>
);

export default App;
