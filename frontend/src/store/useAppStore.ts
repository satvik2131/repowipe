import { create } from "zustand";
import { persist } from "zustand/middleware";
import { validateUser, logoutUser } from "@/api/auth.api";
import { listAllRepos, searchRepos, deleteRepos } from "@/api/repos.api";
import { AxiosResponse } from "axios";

type AppState = {
  isLoading: boolean;
  isError: string | null;
  isAuthenticated: boolean;
  user: User | null;
  allRepos: Repos[];
  searchedRepos: Repos[];
  page: number;
  setPage: (pg: number) => void;
  checkAuth: () => void;
  logout: () => Promise<void>;
  setIsAuthenticated: (auth: boolean) => void;
  setUser: (user: User | null) => void;
  fetchRepos: (filters?: Pick<RepoListFilters, "visibility" | "sort" | "direction">) => Promise<void>;
  findRepos: (searchRepoName: string, filters?: Partial<RepoListFilters>) => Promise<void>;
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
        const { isAuthenticated, user } = get();
        // Clear broken persisted sessions (auth flag without a user object)
        if (isAuthenticated && !user) {
          set({ isAuthenticated: false, user: null });
          localStorage.removeItem("auth");
          return;
        }
        if (isAuthenticated) {
          const isValid = await validateUser();
          if (!isValid) {
            set({ isAuthenticated: false, user: null });
            localStorage.removeItem("auth");
          }
        }
      },

      logout: async () => {
        try {
          await logoutUser();
        } catch (err) {
          console.error("Logout API failed (clearing local session anyway):", err);
        }
        set({
          isAuthenticated: false,
          user: null,
          allRepos: [],
          searchedRepos: [],
          page: 1,
          isLoading: false,
          isError: null,
        });
        localStorage.removeItem("auth");
      },

      findRepos: async (searchRepoName, filters) => {
        set({ isLoading: true, isError: null });
        try {
          const results = await searchRepos(
            get().user!.login,
            searchRepoName,
            filters
          );
          set({ isLoading: false, searchedRepos: Array.isArray(results) ? results : [] });
        } catch (err) {
          console.error("findRepos failed:", err);
          set({ isLoading: false, searchedRepos: [], isError: "Search failed" });
        }
      },

      fetchRepos: async (filters) => {
        set({ isLoading: true, isError: null });
        try {
          const resp = await listAllRepos(get().page, filters);
          set({
            isLoading: false,
            allRepos: Array.isArray(resp.data) ? resp.data : [],
          });
        } catch (err) {
          console.error("fetchRepos failed:", err);
          set({ isLoading: false, allRepos: [], isError: "Failed to load repos" });
        }
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
      onRehydrateStorage: () => (state) => {
        if (state?.isAuthenticated && !state.user) {
          state.setIsAuthenticated(false);
          state.setUser(null);
        }
      },
    }
  )
);
