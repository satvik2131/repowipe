import { Button } from "@/components/ui/button";
import { useAppStore } from "@/store/useAppStore";
import { Github, Code, LogOut, Gitlab } from "lucide-react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useShallow } from "zustand/react/shallow";
import { Avatar, AvatarImage } from "./ui/avatar";
import { AvatarFallback } from "@radix-ui/react-avatar";
import { getProviderLoginUrl } from "@/api/auth.api";

const PROVIDERS: { id: Provider; label: string }[] = [
  { id: "github", label: "GitHub" },
  { id: "gitlab", label: "GitLab" },
  { id: "bitbucket", label: "Bitbucket" },
];

export const Header = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const isSearchPage = location.pathname === "/search";
  const { isAuthenticated, user, logout } = useAppStore(
    useShallow((state) => ({
      isAuthenticated: state.isAuthenticated,
      user: state.user,
      logout: state.logout,
    })),
  );

  const startOAuth = async (provider: Provider) => {
    try {
      const url = await getProviderLoginUrl(provider, "login");
      window.location.href = url;
    } catch (err) {
      console.error(`Failed to get ${provider} login URL:`, err);
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate("/");
  };

  const ProviderIcon = ({ provider }: { provider: Provider }) => {
    if (provider === "gitlab") return <Gitlab className="mr-2 h-4 w-4" />;
    if (provider === "bitbucket") {
      return (
        <span className="mr-2 inline-flex h-4 w-4 items-center justify-center text-[10px] font-bold">
          BB
        </span>
      );
    }
    return <Github className="mr-2 h-4 w-4" />;
  };

  return (
    <header className="sticky top-0 z-50 bg-background/80 backdrop-blur-sm border-b border-border">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-20 gap-4">
          <Link
            to="/"
            className="flex items-center gap-4 hover:opacity-80 transition-opacity shrink-0"
          >
            <Code className="h-8 w-8 text-primary" />
            <h2 className="text-2xl font-bold leading-tight tracking-tighter font-space">
              RepoWipe
            </h2>
          </Link>

          <nav className="hidden md:flex items-center gap-8">
            {!isSearchPage && (
              <>
                <a
                  className="text-sm font-medium leading-normal text-muted-foreground hover:text-foreground transition-colors"
                  href="#features"
                >
                  Features
                </a>
                <a
                  className="text-sm font-medium leading-normal text-muted-foreground hover:text-foreground transition-colors"
                  href="#how-it-works"
                >
                  How It Works
                </a>
                <a
                  className="text-sm font-medium leading-normal text-muted-foreground hover:text-foreground transition-colors"
                  href="/search"
                >
                  Repo Search
                </a>
              </>
            )}
          </nav>

          {!isAuthenticated || !user ? (
            <div className="flex flex-wrap items-center justify-end gap-2">
              {PROVIDERS.map((p) => (
                <Button
                  key={p.id}
                  variant={p.id === "github" ? "default" : "outline"}
                  size="sm"
                  className="font-space"
                  onClick={() => startOAuth(p.id)}
                >
                  <ProviderIcon provider={p.id} />
                  <span className="truncate hidden sm:inline">
                    {p.label}
                  </span>
                </Button>
              ))}
            </div>
          ) : (
            <div className="flex items-center gap-3">
              <Avatar>
                <AvatarImage src={user.avatar_url} />
                <AvatarFallback>
                  <Github />
                </AvatarFallback>
              </Avatar>
              <span className="truncate max-w-[8rem]">{user.login}</span>
              <Button
                variant="outline"
                className="font-space"
                onClick={handleLogout}
              >
                <LogOut className="mr-2 h-4 w-4" />
                <span className="truncate">Logout</span>
              </Button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};
