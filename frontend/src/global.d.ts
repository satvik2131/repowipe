declare type User = {
  login: string;
  html_url: string;
  avatar_url: string;
  public_repos: number;
  total_private_repos?: number; //it will only return if authenticated
};

declare type Repos = {
  id: number;
  name: string;
  description: string | null;
  language: string | null;
  html_url: string;
  stargazers_count: number;
  forks_count: number;
  updated_at: string; // ISO timestamp
  private: boolean;
  fork?: boolean;
  archived?: boolean;
};

declare type RepoVisibility = "all" | "public" | "private";
declare type RepoKind = "all" | "sources" | "forks" | "archived";
declare type RepoSort = "updated" | "full_name" | "created" | "pushed" | "stars";

declare type RepoListFilters = {
  visibility: RepoVisibility;
  sort: RepoSort;
  direction: "asc" | "desc";
  language: string;
  kind: RepoKind;
};

declare type DeleteRepoData = {
  repos: string[];
  username: string;
};
