import { create } from "zustand";
import { persist } from "zustand/middleware";
import { validateUser, listAllRepos, searchRepos } from "@/api/apis_routes";
import { AxiosResponse } from "axios";

type AppState = {
  isLoading: boolean;
  isError: string;
  isAuthenticated: boolean;
  user: User | null;
  allRepos: Repos[] | null;
  searchedRepos: Repos[] | null;
  checkAuth: () => void;
  setIsAuthenticated: (auth: boolean) => void;
  setUser: (user: User | null) => void;
  fetchRepos: () => Promise<AxiosResponse<Repos>>;
  findRepos: (searchRepoName: string) => void;
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
      findRepos: async (searchRepoName) => {
        const searchedRepos = await searchRepos(
          get().user.login,
          searchRepoName
        );

        set({ searchedRepos: searchedRepos });
      },
      fetchRepos: async () => {
        set({ isLoading: true });
        const repositories = await listAllRepos();
        set({ isLoading: false, allRepos: repositories.data });
        return repositories;
      },
      checkAuth: async () => {
        if (get().isAuthenticated) {
          const data: boolean = await validateUser();
          if (!data) {
            set({ isAuthenticated: false });
            return;
          }
        }
      },
      setIsAuthenticated: (auth) => set({ isAuthenticated: auth }),
      setUser: (user) => set({ user }),
    }),
    {
      name: "auth", // key in localStorage
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        user: state.user,
      }), // only persist auth + user
    }
  )
);
