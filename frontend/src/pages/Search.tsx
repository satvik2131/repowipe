import { useEffect, useMemo, useRef, useState } from "react";
import { Button, Input, Card, Badge, Checkbox } from "@/components/ui";
import { Header } from "@/components/Header";
import { Footer } from "@/components/Footer";
import { ConfirmDialog } from "@/components/ui/dialog";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import {
  Search as SearchIcon,
  Star,
  GitFork,
  Calendar,
  Trash2,
  Github,
  Filter,
  X,
  ArrowUpDown,
  ExternalLink,
  Loader2,
} from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { useAppStore } from "@/store/useAppStore";
import { useShallow } from "zustand/react/shallow";
import { COMMON_LANGUAGES, languageColor } from "@/lib/languageColors";
import { cn } from "@/lib/utils";

const DEFAULT_FILTERS: RepoListFilters = {
  visibility: "all",
  sort: "updated",
  direction: "desc",
  language: "",
  kind: "all",
};

const selectClass =
  "h-10 rounded-md border border-input bg-secondary/60 px-3 text-sm text-foreground font-space focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-50";

const Search = () => {
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedRepos, setSelectedRepos] = useState<string[]>([]);
  const [showDialog, setShowDialog] = useState(false);
  const [filters, setFilters] = useState<RepoListFilters>(DEFAULT_FILTERS);
  const { toast } = useToast();

  const {
    fetchRepos,
    findRepos,
    repoCount,
    allRepos,
    page,
    username,
    setPage,
    searchedRepos,
    deleteRepos,
    user,
    isLoading,
  } = useAppStore(
    useShallow((state) => ({
      fetchRepos: state.fetchRepos,
      findRepos: state.findRepos,
      setPage: state.setPage,
      deleteRepos: state.deleteRepos,
      searchedRepos: state.searchedRepos,
      allRepos: state.allRepos,
      user: state.user,
      isLoading: state.isLoading,
      repoCount: state.user
        ? state.user.public_repos + (state.user.total_private_repos ?? 0)
        : 0,
      page: state.page,
      username: state.user?.login ?? "",
    })),
  );

  const usesSearchApi =
    searchTerm.trim().length > 0 ||
    filters.language !== "" ||
    filters.kind !== "all" ||
    filters.sort === "stars";

  const listFilters = useMemo(
    () => ({
      visibility: filters.visibility,
      sort: filters.sort,
      direction: filters.direction,
    }),
    [filters.visibility, filters.sort, filters.direction],
  );

  const languages = useMemo(() => {
    const fromRepos = new Set<string>();
    for (const repo of [...allRepos, ...searchedRepos]) {
      if (repo.language) fromRepos.add(repo.language);
    }
    return Array.from(
      new Set([...COMMON_LANGUAGES, ...fromRepos]),
    ).sort((a, b) => a.localeCompare(b));
  }, [allRepos, searchedRepos]);

  // Always call hooks before any early return.
  useEffect(() => {
    if (!user) return;
    if (usesSearchApi) return;
    fetchRepos(listFilters);
  }, [user, page, listFilters, usesSearchApi]);

  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => {
    if (!user || !usesSearchApi) return;

    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => {
      findRepos(searchTerm.trim(), filters);
    }, 350);

    return () => {
      if (timer.current) clearTimeout(timer.current);
    };
  }, [user, searchTerm, filters, usesSearchApi]);

  // Clear selection when the visible list changes.
  useEffect(() => {
    setSelectedRepos([]);
  }, [page, searchTerm, filters]);

  if (!user) {
    return null;
  }

  const repos: Repos[] = usesSearchApi ? searchedRepos : allRepos;
  const paginationCount = Math.max(1, Math.ceil(repoCount / 10));
  const hasActiveFilters =
    filters.visibility !== "all" ||
    filters.kind !== "all" ||
    filters.language !== "" ||
    filters.sort !== "updated" ||
    filters.direction !== "desc" ||
    searchTerm.trim().length > 0;

  const updateFilter = <K extends keyof RepoListFilters>(
    key: K,
    value: RepoListFilters[K],
  ) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
    setPage(1);
  };

  const clearFilters = () => {
    setSearchTerm("");
    setFilters(DEFAULT_FILTERS);
    setPage(1);
  };

  const handleSelectRepo = (repoName: string) => {
    setSelectedRepos((prev) =>
      prev.includes(repoName)
        ? prev.filter((id) => id !== repoName)
        : [...prev, repoName],
    );
  };

  const handleSelectAll = () => {
    if (selectedRepos.length === repos.length) {
      setSelectedRepos([]);
    } else {
      setSelectedRepos(repos.map((repo) => repo.name));
    }
  };

  const handlePageChange = (currentPage: number) => {
    setPage(currentPage);
  };

  const handleBulkDelete = (e: React.MouseEvent<HTMLButtonElement>) => {
    e.preventDefault();
    if (selectedRepos.length === 0) {
      toast({
        title: "No repos selected",
        description: "Please select at least one repository to delete.",
        variant: "destructive",
      });
      return;
    }
    setShowDialog(true);
  };

  const onConfirm = () => {
    const deleteReposData: DeleteRepoData = {
      repos: selectedRepos,
      username: username,
    };
    deleteRepos(deleteReposData)
      .then((res) => {
        if (res.status === 200) {
          toast({
            title: "Repositories Deleted",
            description: `${selectedRepos.length} repositories deleted`,
            variant: "destructive",
          });
          setSelectedRepos([]);
          if (usesSearchApi) {
            findRepos(searchTerm.trim(), filters);
          } else {
            fetchRepos(listFilters);
          }
        }
      })
      .catch((err) => {
        toast({
          title:
            "Repository not found or you don’t have permission to delete it.",
          description: err?.response?.data
            ? String(err.response.data)
            : "Delete failed",
          variant: "destructive",
        });
      });
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  const runSearch = () => {
    if (usesSearchApi) {
      findRepos(searchTerm.trim(), filters);
    } else {
      fetchRepos(listFilters);
    }
  };

  return (
    <div className="min-h-screen flex flex-col bg-background font-space">
      <Header />

      <main className="flex-1 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="mb-8">
            <div className="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-3">
              <div>
                <h1 className="text-3xl sm:text-4xl font-bold text-foreground tracking-tight">
                  Search & Manage Repos
                </h1>
                <p className="mt-2 text-muted-foreground max-w-xl">
                  Filter, select, and bulk-delete repositories from your GitHub
                  account.
                </p>
              </div>
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Github className="h-4 w-4" />
                <span>
                  {repoCount} total · showing {repos.length}
                </span>
              </div>
            </div>
          </div>

          {/* Search + filters */}
          <Card className="mb-6 border-border/80 bg-secondary/20 shadow-card overflow-hidden">
            <div className="p-4 sm:p-5 space-y-4">
              <div className="flex flex-col sm:flex-row gap-3">
                <div className="flex-1 relative">
                  <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground h-4 w-4 pointer-events-none" />
                  <Input
                    placeholder="Search by name or description..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") runSearch();
                    }}
                    className="pl-10 bg-background/60 border-border"
                  />
                </div>
                <Button className="sm:w-auto shrink-0" onClick={runSearch}>
                  <SearchIcon className="mr-2 h-4 w-4" />
                  Search
                </Button>
              </div>

              <div className="flex items-center gap-2 text-xs uppercase tracking-wide text-muted-foreground">
                <Filter className="h-3.5 w-3.5" />
                Filters
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
                <label className="flex flex-col gap-1.5">
                  <span className="text-xs text-muted-foreground">Visibility</span>
                  <select
                    className={selectClass}
                    value={filters.visibility}
                    onChange={(e) =>
                      updateFilter(
                        "visibility",
                        e.target.value as RepoVisibility,
                      )
                    }
                  >
                    <option value="all">All</option>
                    <option value="public">Public</option>
                    <option value="private">Private</option>
                  </select>
                </label>

                <label className="flex flex-col gap-1.5">
                  <span className="text-xs text-muted-foreground">Type</span>
                  <select
                    className={selectClass}
                    value={filters.kind}
                    onChange={(e) =>
                      updateFilter("kind", e.target.value as RepoKind)
                    }
                  >
                    <option value="all">All</option>
                    <option value="sources">Sources</option>
                    <option value="forks">Forks</option>
                    <option value="archived">Archived</option>
                  </select>
                </label>

                <label className="flex flex-col gap-1.5">
                  <span className="text-xs text-muted-foreground">Language</span>
                  <select
                    className={selectClass}
                    value={filters.language}
                    onChange={(e) => updateFilter("language", e.target.value)}
                  >
                    <option value="">All languages</option>
                    {languages.map((lang) => (
                      <option key={lang} value={lang}>
                        {lang}
                      </option>
                    ))}
                  </select>
                </label>

                <label className="flex flex-col gap-1.5">
                  <span className="text-xs text-muted-foreground">Sort by</span>
                  <select
                    className={selectClass}
                    value={filters.sort}
                    onChange={(e) =>
                      updateFilter("sort", e.target.value as RepoSort)
                    }
                  >
                    <option value="updated">Last updated</option>
                    <option value="full_name">Name</option>
                    <option value="created">Created</option>
                    <option value="pushed">Last pushed</option>
                    <option value="stars">Stars</option>
                  </select>
                </label>

                <label className="flex flex-col gap-1.5">
                  <span className="text-xs text-muted-foreground">Order</span>
                  <select
                    className={selectClass}
                    value={filters.direction}
                    onChange={(e) =>
                      updateFilter(
                        "direction",
                        e.target.value as "asc" | "desc",
                      )
                    }
                    disabled={usesSearchApi}
                  >
                    <option value="desc">Descending</option>
                    <option value="asc">Ascending</option>
                  </select>
                </label>
              </div>

              {hasActiveFilters && (
                <div className="flex flex-wrap items-center gap-2 pt-1">
                  {searchTerm.trim() && (
                    <Badge variant="secondary" className="gap-1 font-normal">
                      Query: {searchTerm.trim()}
                    </Badge>
                  )}
                  {filters.visibility !== "all" && (
                    <Badge variant="secondary" className="gap-1 font-normal capitalize">
                      {filters.visibility}
                    </Badge>
                  )}
                  {filters.kind !== "all" && (
                    <Badge variant="secondary" className="gap-1 font-normal capitalize">
                      {filters.kind}
                    </Badge>
                  )}
                  {filters.language && (
                    <Badge variant="secondary" className="gap-1 font-normal">
                      <span
                        className="inline-block w-2 h-2 rounded-full mr-1"
                        style={{ backgroundColor: languageColor(filters.language) }}
                      />
                      {filters.language}
                    </Badge>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 px-2 text-muted-foreground"
                    onClick={clearFilters}
                  >
                    <X className="mr-1 h-3.5 w-3.5" />
                    Clear filters
                  </Button>
                </div>
              )}
            </div>
          </Card>

          {/* Actions bar */}
          <div className="sticky top-20 z-20 mb-4 rounded-lg border border-border/80 bg-background/90 backdrop-blur-md px-4 py-3 flex flex-col sm:flex-row justify-between items-stretch sm:items-center gap-3 shadow-card">
            <div className="flex items-center gap-3">
              <Button variant="outline" size="sm" onClick={handleSelectAll}>
                {repos.length > 0 && selectedRepos.length === repos.length
                  ? "Deselect All"
                  : "Select All"}
              </Button>
              <span className="text-sm text-muted-foreground">
                <span className="text-foreground font-medium">
                  {selectedRepos.length}
                </span>{" "}
                of {repos.length} selected
              </span>
              {isLoading && (
                <Loader2 className="h-4 w-4 animate-spin text-primary" />
              )}
            </div>

            <div className="flex items-center gap-2">
              <ConfirmDialog
                open={showDialog}
                onOpenChange={setShowDialog}
                onConfirm={onConfirm}
              />
              <Button
                variant="destructive"
                size="sm"
                onClick={handleBulkDelete}
                disabled={selectedRepos.length === 0}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete Selected ({selectedRepos.length})
              </Button>
            </div>
          </div>

          {/* Repository list */}
          <div className="space-y-2.5">
            {isLoading && repos.length === 0 ? (
              Array.from({ length: 5 }).map((_, i) => (
                <Card
                  key={i}
                  className="p-4 border-border/60 bg-secondary/10 animate-pulse"
                >
                  <div className="h-5 w-48 bg-muted rounded mb-3" />
                  <div className="h-4 w-full max-w-md bg-muted/70 rounded mb-2" />
                  <div className="h-3 w-40 bg-muted/50 rounded" />
                </Card>
              ))
            ) : repos.length === 0 ? (
              <Card className="p-10 text-center border-dashed border-border bg-secondary/10">
                <Github className="h-12 w-12 text-muted-foreground mx-auto mb-4 opacity-60" />
                <h3 className="text-lg font-semibold text-foreground mb-2">
                  No repos found
                </h3>
                <p className="text-muted-foreground mb-4 max-w-md mx-auto">
                  Try adjusting your filters or search terms.
                </p>
                {hasActiveFilters && (
                  <Button variant="outline" onClick={clearFilters}>
                    <X className="mr-2 h-4 w-4" />
                    Clear filters
                  </Button>
                )}
              </Card>
            ) : (
              repos.map((repo) => {
                const selected = selectedRepos.includes(repo.name);
                return (
                  <Card
                    key={repo.id}
                    className={cn(
                      "p-4 sm:p-5 border-border/70 bg-card/40 transition-all duration-200 hover:border-primary/40 hover:bg-secondary/30",
                      selected && "border-primary/50 bg-primary/5",
                    )}
                  >
                    <div className="flex items-start gap-3 sm:gap-4">
                      <Checkbox
                        checked={selected}
                        onCheckedChange={() => handleSelectRepo(repo.name)}
                        className="mt-1"
                        aria-label={`Select ${repo.name}`}
                      />

                      <div className="flex-1 min-w-0">
                        <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-2 mb-1.5">
                          <div className="flex flex-wrap items-center gap-2 min-w-0">
                            <a
                              href={repo.html_url}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-base sm:text-lg font-semibold text-foreground hover:text-primary transition-colors truncate inline-flex items-center gap-1.5"
                            >
                              {repo.name}
                              <ExternalLink className="h-3.5 w-3.5 opacity-50 shrink-0" />
                            </a>
                            <Badge
                              variant={repo.private ? "secondary" : "outline"}
                              className="text-[10px] uppercase tracking-wide"
                            >
                              {repo.private ? "Private" : "Public"}
                            </Badge>
                            {repo.fork && (
                              <Badge variant="outline" className="text-[10px]">
                                Fork
                              </Badge>
                            )}
                            {repo.archived && (
                              <Badge variant="destructive" className="text-[10px]">
                                Archived
                              </Badge>
                            )}
                          </div>
                          <div className="flex items-center gap-3 text-sm text-muted-foreground shrink-0">
                            <span className="inline-flex items-center gap-1">
                              <Star className="h-3.5 w-3.5" />
                              {repo.stargazers_count}
                            </span>
                            <span className="inline-flex items-center gap-1">
                              <GitFork className="h-3.5 w-3.5" />
                              {repo.forks_count}
                            </span>
                          </div>
                        </div>

                        {repo.description && (
                          <p className="text-sm text-muted-foreground mb-2.5 line-clamp-2">
                            {repo.description}
                          </p>
                        )}

                        <div className="flex flex-wrap items-center gap-x-4 gap-y-1.5 text-xs sm:text-sm text-muted-foreground">
                          {repo.language ? (
                            <span className="inline-flex items-center gap-1.5">
                              <span
                                className="w-2.5 h-2.5 rounded-full"
                                style={{
                                  backgroundColor: languageColor(repo.language),
                                }}
                              />
                              {repo.language}
                            </span>
                          ) : (
                            <span className="inline-flex items-center gap-1.5 text-muted-foreground/70">
                              No language
                            </span>
                          )}
                          <span className="inline-flex items-center gap-1.5">
                            <Calendar className="h-3.5 w-3.5" />
                            Updated {formatDate(repo.updated_at)}
                          </span>
                          {!usesSearchApi && filters.sort === "full_name" && (
                            <span className="inline-flex items-center gap-1 opacity-60">
                              <ArrowUpDown className="h-3.5 w-3.5" />
                              Sorted by name
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  </Card>
                );
              })
            )}
          </div>

          {!usesSearchApi && paginationCount > 1 && (
            <Pagination className="mt-8">
              <PaginationContent>
                <PaginationItem>
                  <PaginationPrevious
                    onClick={(e) => {
                      e.preventDefault();
                      if (page - 1 > 0) handlePageChange(page - 1);
                    }}
                    className={page <= 1 ? "pointer-events-none opacity-40" : "cursor-pointer"}
                  />
                </PaginationItem>
                {Array.from({ length: Math.min(paginationCount, 12) }).map(
                  (_, count) => {
                    const currPage = count + 1;
                    return (
                      <PaginationItem key={currPage}>
                        <PaginationLink
                          href="#"
                          isActive={currPage === page}
                          onClick={(e) => {
                            e.preventDefault();
                            handlePageChange(currPage);
                          }}
                        >
                          {currPage}
                        </PaginationLink>
                      </PaginationItem>
                    );
                  },
                )}
                <PaginationItem>
                  <PaginationNext
                    onClick={(e) => {
                      e.preventDefault();
                      if (page + 1 <= paginationCount) {
                        handlePageChange(page + 1);
                      }
                    }}
                    className={
                      page >= paginationCount
                        ? "pointer-events-none opacity-40"
                        : "cursor-pointer"
                    }
                  />
                </PaginationItem>
              </PaginationContent>
            </Pagination>
          )}
        </div>
      </main>
      <Footer />
    </div>
  );
};

export default Search;
