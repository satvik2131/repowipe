import axios, { AxiosResponse, HttpStatusCode } from "axios";

// Prefer same-origin `/api` in production (Vercel rewrites proxy to Koyeb).
// That avoids Brave/Shields and third-party-cookie blocks on cross-site API calls.
const apiURL =
  import.meta.env.VITE_API_URL ||
  (window.location.hostname === "localhost"
    ? "http://localhost:8080/api"
    : "/api");

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
