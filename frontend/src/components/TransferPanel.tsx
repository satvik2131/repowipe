import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { useAppStore } from "@/store/useAppStore";
import { ArrowRightLeft, Loader2, ExternalLink } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useShallow } from "zustand/react/shallow";
import { useToast } from "@/hooks/use-toast";

const label = (p: Provider) =>
  p === "github" ? "GitHub" : p === "gitlab" ? "GitLab" : "Bitbucket";

type Props = {
  selectedRepos: string[];
};

export const TransferPanel = ({ selectedRepos }: Props) => {
  const { toast } = useToast();
  const [destination, setDestination] = useState<Provider | "">("");
  const [submitting, setSubmitting] = useState(false);

  const {
    activeProvider,
    connections,
    transferJob,
    startTransferJob,
    pollTransferJob,
    clearTransferJob,
  } = useAppStore(
    useShallow((s) => ({
      activeProvider: s.activeProvider,
      connections: s.connections,
      transferJob: s.transferJob,
      startTransferJob: s.startTransferJob,
      pollTransferJob: s.pollTransferJob,
      clearTransferJob: s.clearTransferJob,
    }))
  );

  const destinations = useMemo(
    () =>
      (["github", "gitlab", "bitbucket"] as Provider[]).filter(
        (p) => p !== activeProvider && connections[p]
      ),
    [activeProvider, connections]
  );

  useEffect(() => {
    if (destinations.length === 1) setDestination(destinations[0]);
    else if (destination && !destinations.includes(destination as Provider)) {
      setDestination("");
    }
  }, [destinations, destination]);

  useEffect(() => {
    if (!transferJob?.id) return;
    if (transferJob.status === "completed" || transferJob.status === "failed") {
      return;
    }
    const t = setInterval(() => {
      pollTransferJob(transferJob.id).catch(console.error);
    }, 2000);
    return () => clearInterval(t);
  }, [transferJob?.id, transferJob?.status, pollTransferJob]);

  const onTransfer = async () => {
    if (!destination) {
      toast({
        title: "Pick a destination",
        description: "Link another provider first, then choose where to transfer.",
        variant: "destructive",
      });
      return;
    }
    if (selectedRepos.length === 0) {
      toast({
        title: "No repos selected",
        variant: "destructive",
      });
      return;
    }
    setSubmitting(true);
    try {
      await startTransferJob(destination, selectedRepos);
      toast({
        title: "Transfer started",
        description: `Copying ${selectedRepos.length} repo(s) to ${label(destination)}.`,
      });
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data
          ?.error ?? "Transfer failed to start";
      toast({ title: "Transfer error", description: msg, variant: "destructive" });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="space-y-3">
      <div className="flex flex-col sm:flex-row sm:items-center gap-2">
        <select
          className="h-9 rounded-md border border-input bg-secondary/60 px-3 text-sm"
          value={destination}
          onChange={(e) => setDestination(e.target.value as Provider | "")}
          disabled={destinations.length === 0}
        >
          <option value="">
            {destinations.length === 0
              ? "Link another provider to transfer"
              : "Transfer to…"}
          </option>
          {destinations.map((p) => (
            <option key={p} value={p}>
              {label(p)}
            </option>
          ))}
        </select>
        <Button
          size="sm"
          variant="secondary"
          disabled={
            submitting ||
            selectedRepos.length === 0 ||
            !destination ||
            destinations.length === 0
          }
          onClick={onTransfer}
        >
          {submitting ? (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <ArrowRightLeft className="mr-2 h-4 w-4" />
          )}
          Transfer ({selectedRepos.length})
        </Button>
      </div>

      {transferJob && (
        <div className="rounded-md border border-border/60 bg-background/60 p-3 text-sm space-y-2">
          <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2">
              <span className="font-medium">
                {label(transferJob.source)} → {label(transferJob.destination)}
              </span>
              <Badge variant="outline" className="capitalize">
                {transferJob.status}
              </Badge>
              {(transferJob.status === "queued" ||
                transferJob.status === "running") && (
                <Loader2 className="h-3.5 w-3.5 animate-spin text-primary" />
              )}
            </div>
            {(transferJob.status === "completed" ||
              transferJob.status === "failed") && (
              <Button size="sm" variant="ghost" onClick={clearTransferJob}>
                Dismiss
              </Button>
            )}
          </div>
          {transferJob.error && (
            <p className="text-destructive text-xs">{transferJob.error}</p>
          )}
          <ul className="space-y-1.5">
            {transferJob.repos.map((r) => (
              <li key={r.repo} className="text-xs text-muted-foreground">
                <span className="text-foreground font-medium">{r.repo}</span>{" "}
                <span className="capitalize">· {r.status}</span>
                {r.dest_url && (
                  <a
                    href={r.dest_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="ml-1 inline-flex items-center gap-0.5 text-primary hover:underline"
                  >
                    open <ExternalLink className="h-3 w-3" />
                  </a>
                )}
                {r.error && (
                  <span className="block text-destructive">{r.error}</span>
                )}
                {r.warnings?.map((w) => (
                  <span key={w} className="block text-amber-600 dark:text-amber-400">
                    {w}
                  </span>
                ))}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};
