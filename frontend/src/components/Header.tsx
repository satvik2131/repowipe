import { Button } from "@/components/ui/button";
import { useAppStore } from "@/store/useAppStore";
import { Github, Code } from "lucide-react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { useShallow } from "zustand/react/shallow";
import { Avatar, AvatarImage } from "./ui/avatar";
import { AvatarFallback } from "@radix-ui/react-avatar";

export const Header = () => {
  const location = useLocation();
  const isSearchPage = location.pathname === "/search";
  const { isAuthenticated, user } = useAppStore(
    useShallow((state) => ({
      isAuthenticated: state.isAuthenticated,
      user: state.user,
    }))
  );

  //sign in with github oauth
  const githubOAuth = () => {
    const git_redirect_url: string = import.meta.env.VITE_GITHUB_REDIRECT_URI;
    const client_id: string = import.meta.env.VITE_CLIENT_ID;
    const state: string = import.meta.env.VITE_GITHUB_SECRET;
    const redirect_url: string = import.meta.env.VITE_REDIRECT_URL;
    const scope: string = "repo,user,delete_repo";

    window.location.href = `${git_redirect_url}?client_id=${client_id}&redirect_uri=${redirect_url}&scope=${scope}&state=${state}`;
  };

  return (
    <header className="sticky top-0 z-50 bg-background/80 backdrop-blur-sm border-b border-border">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-20">
          <Link
            to="/"
            className="flex items-center gap-4 hover:opacity-80 transition-opacity"
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

          {!isAuthenticated ? (
            <div className="flex items-center gap-4">
              <Button className="font-space" onClick={githubOAuth}>
                <Github className="mr-2 h-5 w-5" />
                <span className="truncate">Login with GitHub</span>
              </Button>
            </div>
          ) : (
            <div className="flex items-center gap-4">
              <Avatar>
                <AvatarImage src={user.avatar_url} />
                <AvatarFallback>
                  <Github />
                </AvatarFallback>
              </Avatar>
              <span className="truncate">{user.login}</span>
            </div>
          )}
        </div>
      </div>
    </header>
  );
};
