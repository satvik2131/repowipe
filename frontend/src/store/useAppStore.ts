import { create } from "zustand";
import { persist } from "zustand/middleware";
import { validateUser } from "@/api/auth.api";
import { listAllRepos, searchRepos, deleteRepos } from "@/api/repos.api";
import { AxiosResponse } from "axios";

type AppState = {
  isLoading: boolean;
  isError: string | null;
  isAuthenticated: boolean;
  user: User | null;
  allRepos: Repos[] | null;
  searchedRepos: Repos[] | null;
  page: number;
  setPage: (pg: number) => void;
  checkAuth: () => void;
  setIsAuthenticated: (auth: boolean) => void;
  setUser: (user: User | null) => void;
  fetchRepos: () => void;
  findRepos: (searchRepoName: string) => void;
  deleteRepos: (repoData: DeleteRepoData) => Promise<AxiosResponse>;
};

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      isLoading: false,
      isError: null,
      isAuthenticated: false,
      user: null,
      allRepos: [],
      searchedRepos: [],
      page: 1,

      setPage: (pg) => set({ page: pg }),

      checkAuth: async () => {
        if (get().isAuthenticated) {
          const isValid = await validateUser();
          if (!isValid) {
            set({ isAuthenticated: false, user: null });
          }
        }
      },

      findRepos: async (searchRepoName) => {
        set({ isLoading: true });
        const results = await searchRepos(get().user!.login, searchRepoName);
        set({ isLoading: false, searchedRepos: results });
      },

      fetchRepos: async () => {
        set({ isLoading: true });
        const resp = await listAllRepos(get().page);
        set({ isLoading: false, allRepos: resp.data });
      },

      deleteRepos: async (repoData) => {
        return deleteRepos(repoData);
      },

      setIsAuthenticated: (auth) => set({ isAuthenticated: auth }),
      setUser: (user) => set({ user }),
    }),
    {
      name: "auth",
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        user: state.user,
      }),
    }
  )
);
