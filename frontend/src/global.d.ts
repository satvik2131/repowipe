declare type User = {
  login: string;
  html_url: string;
  avatar_url: string;
  public_repos: number;
  total_private_repos?: number; //it will only return if authenticated
};

declare type Repos =
  | {
      id: number;
      name: string;
      description: string;
      language: string | null;
      html_url: string;
      stargazers_count: number;
      forks_count: number;
      updated_at: string; // ISO timestamp
      private: boolean;
    }
  | [];

declare type DeleteRepoData = {
  repos: string[];
  username: string;
};
