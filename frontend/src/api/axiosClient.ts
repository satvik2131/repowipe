import axios, { AxiosResponse, HttpStatusCode } from "axios";

const apiURL =
  import.meta.env.VITE_API_URL ||
  (window.location.hostname === "localhost"
    ? "http://localhost:8080/api"
    : undefined);

if (!apiURL) {
  throw new Error(
    "VITE_API_URL is not set. Configure it in the Vercel build environment to point at the Koyeb backend (e.g. https://<app>.koyeb.app/api).",
  );
}

const axiosClient = axios.create({
  baseURL: apiURL,
  withCredentials: true, // send session cookie cross-site
});

// Optional: add interceptors
axiosClient.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error("API Error:", error.response?.data || error.message);

    if (error.response?.status === HttpStatusCode.Unauthorized) {
      localStorage.removeItem("auth");
      window.location.href = "/";
    }
    //Handling error responses
    return Promise.reject(error);
  },
);

export default axiosClient;
