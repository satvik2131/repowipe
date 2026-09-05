import { create } from "zustand";
import { persist } from "zustand/middleware";
import {
  validateUser,
  logoutUser,
  getConnections,
} from "@/api/auth.api";
import { listAllRepos, searchRepos, deleteRepos } from "@/api/repos.api";
import { startTransfer, getTransfer } from "@/api/transfers.api";
import { AxiosResponse } from "axios";

type AppState = {
  isLoading: boolean;
  isError: string | null;
  isAuthenticated: boolean;
  user: User | null;
  primary: Provider | null;
  connections: Partial<Record<Provider, User>>;
  activeProvider: Provider;
  allRepos: Repos[];
  searchedRepos: Repos[];
  page: number;
  transferJob: TransferJob | null;
  setPage: (pg: number) => void;
  checkAuth: () => void;
  logout: () => Promise<void>;
  setIsAuthenticated: (auth: boolean) => void;
  setUser: (user: User | null) => void;
  setConnections: (c: ConnectionsResponse) => void;
  setActiveProvider: (p: Provider) => void;
  refreshConnections: () => Promise<void>;
  fetchRepos: (
    filters?: Pick<RepoListFilters, "visibility" | "sort" | "direction">
  ) => Promise<void>;
  findRepos: (
    searchRepoName: string,
    filters?: Partial<RepoListFilters>
  ) => Promise<void>;
  deleteRepos: (repoData: DeleteRepoData) => Promise<AxiosResponse>;
  startTransferJob: (
    destination: Provider,
    repos: string[]
  ) => Promise<TransferJob>;
  pollTransferJob: (id: string) => Promise<TransferJob>;
  clearTransferJob: () => void;
};

const activeUsername = (
  connections: Partial<Record<Provider, User>>,
  active: Provider,
  fallback: User | null
) => connections[active]?.login ?? fallback?.login ?? "";

export const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      isLoading: false,
      isError: null,
      isAuthenticated: false,
      user: null,
      primary: null,
      connections: {},
      activeProvider: "github",
      allRepos: [],
      searchedRepos: [],
      page: 1,
      transferJob: null,

      setPage: (pg) => set({ page: pg }),

      checkAuth: async () => {
        const { isAuthenticated, user } = get();
        if (isAuthenticated && !user) {
          set({ isAuthenticated: false, user: null, connections: {} });
          localStorage.removeItem("auth");
          return;
        }
        if (isAuthenticated) {
          const isValid = await validateUser();
          if (!isValid) {
            set({
              isAuthenticated: false,
              user: null,
              connections: {},
              primary: null,
            });
            localStorage.removeItem("auth");
            return;
          }
          try {
            const conn = await getConnections();
            set({
              connections: conn.connections ?? {},
              primary: conn.primary,
              activeProvider: get().activeProvider || conn.primary,
            });
          } catch {
            /* ignore */
          }
        }
      },

      logout: async () => {
        try {
          await logoutUser();
        } catch (err) {
          console.error("Logout API failed:", err);
        }
        set({
          isAuthenticated: false,
          user: null,
          primary: null,
          connections: {},
          allRepos: [],
          searchedRepos: [],
          page: 1,
          transferJob: null,
          isLoading: false,
          isError: null,
        });
        localStorage.removeItem("auth");
      },

      setConnections: (c) =>
        set({
          connections: c.connections ?? {},
          primary: c.primary,
        }),

      setActiveProvider: (p) =>
        set({ activeProvider: p, page: 1, allRepos: [], searchedRepos: [] }),

      refreshConnections: async () => {
        try {
          const conn = await getConnections();
          set({
            connections: conn.connections ?? {},
            primary: conn.primary,
          });
        } catch (err) {
          console.error("refreshConnections failed:", err);
        }
      },

      findRepos: async (searchRepoName, filters) => {
        set({ isLoading: true, isError: null });
        try {
          const { activeProvider, connections, user } = get();
          const username = activeUsername(connections, activeProvider, user);
          const results = await searchRepos(
            activeProvider,
            username,
            searchRepoName,
            filters
          );
          set({
            isLoading: false,
            searchedRepos: Array.isArray(results) ? results : [],
          });
        } catch (err) {
          console.error("findRepos failed:", err);
          set({ isLoading: false, searchedRepos: [], isError: "Search failed" });
        }
      },

      fetchRepos: async (filters) => {
        set({ isLoading: true, isError: null });
        try {
          const resp = await listAllRepos(
            get().activeProvider,
            get().page,
            filters
          );
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
        return deleteRepos(get().activeProvider, repoData);
      },

      startTransferJob: async (destination, repos) => {
        const source = get().activeProvider;
        const job = await startTransfer({ source, destination, repos });
        set({ transferJob: job });
        return job;
      },

      pollTransferJob: async (id) => {
        const job = await getTransfer(id);
        set({ transferJob: job });
        return job;
      },

      clearTransferJob: () => set({ transferJob: null }),

      setIsAuthenticated: (auth) => set({ isAuthenticated: auth }),
      setUser: (user) => set({ user }),
    }),
    {
      name: "auth",
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        user: state.user,
        primary: state.primary,
        connections: state.connections,
        activeProvider: state.activeProvider,
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
