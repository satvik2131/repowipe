import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { getProviderLoginUrl, unlinkProvider } from "@/api/auth.api";
import { useAppStore } from "@/store/useAppStore";
import { Link2, Unlink, Loader2 } from "lucide-react";
import { useState } from "react";
import { useShallow } from "zustand/react/shallow";
import { useToast } from "@/hooks/use-toast";

const ALL: Provider[] = ["github", "gitlab", "bitbucket"];

const label = (p: Provider) =>
  p === "github" ? "GitHub" : p === "gitlab" ? "GitLab" : "Bitbucket";

export const ConnectionsPanel = () => {
  const { toast } = useToast();
  const [busy, setBusy] = useState<string | null>(null);
  const { connections, primary, refreshConnections, setConnections } =
    useAppStore(
      useShallow((s) => ({
        connections: s.connections,
        primary: s.primary,
        refreshConnections: s.refreshConnections,
        setConnections: s.setConnections,
      }))
    );

  const link = async (provider: Provider) => {
    setBusy(provider);
    try {
      const url = await getProviderLoginUrl(provider, "link");
      window.location.href = url;
    } catch (err) {
      console.error(err);
      toast({
        title: "Could not start linking",
        description: String(err),
        variant: "destructive",
      });
      setBusy(null);
    }
  };

  const unlink = async (provider: Provider) => {
    if (provider === primary) {
      toast({
        title: "Cannot unlink primary",
        description: "Log out to disconnect your primary account.",
        variant: "destructive",
      });
      return;
    }
    setBusy(provider);
    try {
      const next = await unlinkProvider(provider);
      setConnections(next);
      await refreshConnections();
      toast({ title: `Unlinked ${label(provider)}` });
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data
          ?.error ?? "Unlink failed";
      toast({ title: "Unlink failed", description: msg, variant: "destructive" });
    } finally {
      setBusy(null);
    }
  };

  return (
    <div className="rounded-lg border border-border/80 bg-secondary/20 p-4 space-y-3">
      <div>
        <h2 className="text-sm font-semibold text-foreground">Connected accounts</h2>
        <p className="text-xs text-muted-foreground mt-0.5">
          Link GitHub, GitLab, and Bitbucket to wipe or transfer across hosts.
        </p>
      </div>
      <div className="flex flex-wrap gap-2">
        {ALL.map((p) => {
          const linked = Boolean(connections[p]);
          const isPrimary = primary === p;
          return (
            <div
              key={p}
              className="flex items-center gap-2 rounded-md border border-border/60 bg-background/50 px-3 py-2"
            >
              <div className="min-w-0">
                <div className="flex items-center gap-1.5">
                  <span className="text-sm font-medium">{label(p)}</span>
                  {linked && (
                    <Badge variant="secondary" className="text-[10px]">
                      {isPrimary ? "primary" : "linked"}
                    </Badge>
                  )}
                </div>
                {linked && connections[p]?.login && (
                  <p className="text-xs text-muted-foreground truncate max-w-[10rem]">
                    @{connections[p]!.login}
                  </p>
                )}
              </div>
              {linked ? (
                <Button
                  size="sm"
                  variant="ghost"
                  disabled={isPrimary || busy === p}
                  onClick={() => unlink(p)}
                  title={isPrimary ? "Primary — log out to disconnect" : "Unlink"}
                >
                  {busy === p ? (
                    <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  ) : (
                    <Unlink className="h-3.5 w-3.5" />
                  )}
                </Button>
              ) : (
                <Button
                  size="sm"
                  variant="outline"
                  disabled={busy === p}
                  onClick={() => link(p)}
                >
                  {busy === p ? (
                    <Loader2 className="mr-1 h-3.5 w-3.5 animate-spin" />
                  ) : (
                    <Link2 className="mr-1 h-3.5 w-3.5" />
                  )}
                  Connect
                </Button>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};
