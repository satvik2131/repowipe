declare type Provider = "github" | "gitlab" | "bitbucket";

declare type User = {
  login: string;
  html_url: string;
  avatar_url: string;
  public_repos: number;
  total_private_repos?: number;
  provider?: Provider;
};

declare type Repos = {
  id: number;
  name: string;
  full_name?: string;
  description: string | null;
  language: string | null;
  html_url: string;
  stargazers_count: number;
  forks_count: number;
  updated_at: string;
  private: boolean;
  fork?: boolean;
  archived?: boolean;
  provider?: Provider;
  owner_login?: string;
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

declare type ConnectionsResponse = {
  primary: Provider;
  connections: Partial<Record<Provider, User>>;
};

declare type TransferRepoResult = {
  repo: string;
  status: "pending" | "running" | "succeeded" | "partial" | "failed";
  dest_url?: string;
  warnings?: string[];
  error?: string;
};

declare type TransferJob = {
  id: string;
  source: Provider;
  destination: Provider;
  status: "queued" | "running" | "completed" | "failed";
  repos: TransferRepoResult[];
  created_at: number;
  updated_at: number;
  error?: string;
};
