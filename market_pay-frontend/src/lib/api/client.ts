import axios from "axios";

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1",
  headers: { "Content-Type": "application/json" },
});

apiClient.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const mode = localStorage.getItem("marketpay_mode") || "demo";
    config.headers["X-MarketPay-Mode"] = mode;
    const token = localStorage.getItem("marketpay_token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => {
    const body = response.data;
    if (body && typeof body === "object" && "success" in body) {
      const { message, data, total, page, limit, total_pages, ...rest } = body;
      if (data !== undefined && total !== undefined) {
        response.data = { data, total, page, limit, total_pages };
      } else if (data !== undefined) {
        response.data = data;
      } else {
        response.data = rest;
      }
      if (message) {
        Object.defineProperty(response, "message", { value: message, writable: true });
      }
    }
    return response;
  },
  async (error) => {
    const original = error.config;
    if (error.response?.status === 401 && !original._retry) {
      original._retry = true;
      const refreshToken = localStorage.getItem("marketpay_refresh");
      if (refreshToken) {
        try {
          const { data } = await axios.post(
            `${apiClient.defaults.baseURL}/auth/refresh`,
            { refresh_token: refreshToken }
          );
          localStorage.setItem("marketpay_token", data.access_token);
          if (data.refresh_token) {
            localStorage.setItem("marketpay_refresh", data.refresh_token);
          }
          original.headers.Authorization = `Bearer ${data.access_token}`;
          return apiClient(original);
        } catch {
          localStorage.removeItem("marketpay_token");
          localStorage.removeItem("marketpay_refresh");
          localStorage.removeItem("marketpay_user");
          window.location.href = "/login";
        }
      }
    }

    const data = error.response?.data;
    if (data && typeof data === "object") {
      if (data.message) {
        error.message = data.message;
      }
      if ("success" in data) {
        error.response.data = data;
      }
    }

    return Promise.reject(error);
  }
);

export default apiClient;
