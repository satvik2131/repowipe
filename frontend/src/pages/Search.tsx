import { useEffect, useRef, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Header } from "@/components/Header";
import { Footer } from "@/components/Footer";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
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
} from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { useAppStore } from "@/store/useAppStore";
import { useShallow } from "zustand/react/shallow";

const Search = () => {
  const [searchTerm, setSearchTerm] = useState("");
  const [selectedRepos, setSelectedRepos] = useState<number[]>([]);
  const { toast } = useToast();
  const { fetchRepos, findRepos, allRepos, searchedRepos } = useAppStore(
    useShallow((state) => ({
      fetchRepos: state.fetchRepos,
      findRepos: state.findRepos,
      searchedRepos: state.searchedRepos,
      allRepos: state.allRepos,
      repoCount: state.user.public_repos + state.user.total_private_repos,
    }))
  );
  const timer = useRef<NodeJS.Timeout | null>(null);
  const repos: Repos[] | null =
    searchTerm.length !== 0 ? searchedRepos : allRepos;

  //Count the pages for number of repositories
  const paginationCount = Math.ceil(repoCount / 10);

  //Fetched all Repositories
  useEffect(() => {
    fetchRepos();
  }, []);

  console.log("all Repo--", allRepos);
  console.log("searched repo--", searchedRepos);
  console.log("search term--", searchTerm);

  //Search the repos and updates the repos state (with debouncing)
  useEffect(() => {
    if (searchTerm === "") return;

    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => {
      findRepos(searchTerm);
    }, 500);

    return () => {
      if (timer.current) clearTimeout(timer.current);
    };
  }, [searchTerm]);

  const handleSelectRepo = (repoId: number) => {
    setSelectedRepos((prev) =>
      prev.includes(repoId)
        ? prev.filter((id) => id !== repoId)
        : [...prev, repoId]
    );
  };

  const handleSelectAll = () => {
    if (selectedRepos.length === repos?.length) {
      setSelectedRepos([]);
    } else {
      setSelectedRepos(repos.map((repo) => repo.id));
    }
  };

  // const handleBulkDelete = () => {
  //   if (selectedRepos.length === 0) {
  //     toast({
  //       title: "No repos selected",
  //       description: "Please select at least one repository to delete.",
  //       variant: "destructive",
  //     });
  //     return;
  //   }

  //   toast({
  //     title: "Bulk deletion initiated",
  //     description: `${selectedRepos.length} repos will be deleted. This is a demo - no actual deletion occurred.`,
  //   });

  //   // In a real app, this would make API calls to delete repos
  //   setSelectedRepos([]);
  // };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  };

  return (
    <div className="min-h-screen flex flex-col bg-background font-space">
      <Header />

      <main className="flex-1 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header Section */}
          <div className="text-center mb-12">
            <h1 className="text-4xl font-bold text-foreground mb-4 font-space">
              Search & Manage repos
            </h1>
            <p className="text-lg text-muted-foreground max-w-2xl mx-auto font-space">
              Find and select repos to delete. Use the search to filter by name
              or description.
            </p>
          </div>

          {/* Search Section */}
          <Card className="p-6 mb-8 shadow-card">
            <div className="flex flex-col sm:flex-row gap-4 items-center">
              <div className="flex-1 relative">
                <SearchIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
                <Input
                  placeholder="Search repos by name or description..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 font-space"
                />
              </div>
              <Button className="w-full sm:w-auto font-space">
                <SearchIcon className="mr-2 h-4 w-4" />
                Search
              </Button>
            </div>
          </Card>

          {/* Actions Bar */}
          <div className="flex flex-col sm:flex-row justify-between items-center mb-6 gap-4">
            <div className="flex items-center gap-4">
              <Button
                variant="outline"
                onClick={handleSelectAll}
                className="font-space"
              >
                {selectedRepos.length === repos.length
                  ? "Deselect All"
                  : "Select All"}
              </Button>
              <span className="text-sm text-muted-foreground font-space">
                {selectedRepos.length} of {repos.length} selected
              </span>
            </div>
            <Button
              variant="destructive"
              disabled={selectedRepos.length === 0}
              className="font-space"
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete Selected ({selectedRepos.length})
            </Button>
          </div>

          {/* Repository List */}
          <div className="space-y-4">
            {repos.length === 0 ? (
              <Card className="p-8 text-center shadow-card">
                <Github className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-lg font-semibold text-foreground mb-2 font-space">
                  No repos found
                </h3>
                <p className="text-muted-foreground font-space">
                  Try adjusting your search terms or check your GitHub
                  connection.
                </p>
              </Card>
            ) : (
              repos.map((repo) => (
                <Card
                  key={repo.id}
                  className="p-6 shadow-card hover:shadow-elegant transition-smooth"
                >
                  <div className="flex items-start gap-4">
                    <Checkbox
                      checked={selectedRepos.includes(repo.id)}
                      onCheckedChange={() => handleSelectRepo(repo.id)}
                      className="mt-1"
                    />

                    <div className="flex-1 min-w-0">
                      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-2">
                        <div className="flex items-center gap-2">
                          <h3 className="text-lg font-semibold text-foreground font-space">
                            {repo.name}
                          </h3>
                          {repo.private && (
                            <Badge variant="secondary" className="text-xs">
                              Private
                            </Badge>
                          )}
                        </div>
                        <div className="flex items-center gap-4 text-sm text-muted-foreground">
                          <div className="flex items-center gap-1">
                            <Star className="h-4 w-4" />
                            {repo.stargazers_count}
                          </div>
                          <div className="flex items-center gap-1">
                            <GitFork className="h-4 w-4" />
                            {repo.forks_count}
                          </div>
                        </div>
                      </div>

                      <p className="text-muted-foreground mb-3 font-space">
                        {repo.description}
                      </p>

                      <div className="flex flex-wrap items-center gap-4 text-sm text-muted-foreground">
                        <div className="flex items-center gap-1">
                          <div className="w-3 h-3 rounded-full bg-primary" />
                          {repo.language}
                        </div>
                        <div className="flex items-center gap-1">
                          <Calendar className="h-4 w-4" />
                          Updated {formatDate(repo.updated_at)}
                        </div>
                      </div>
                    </div>
                  </div>
                </Card>
              ))
            )}
          </div>
        </div>
        <Pagination className="mt-5">
          <PaginationContent>
            <PaginationItem>
              <PaginationPrevious href="#" />
            </PaginationItem>
            {Array.from({ length: paginationCount }).map((_, count) => {
              return (
                <PaginationItem key={count + 1}>
                  <PaginationLink href="#">{count + 1}</PaginationLink>
                </PaginationItem>
              );
            })}
            <PaginationItem>
              <PaginationNext href="#" />
            </PaginationItem>
          </PaginationContent>
        </Pagination>
      </main>

      <Footer />
    </div>
  );
};

export default Search;
