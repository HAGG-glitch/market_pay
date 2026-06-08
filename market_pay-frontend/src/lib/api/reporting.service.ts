import apiClient from "./client";
import type { PortfolioStats } from "@/types";

export async function getPortfolioStats() {
  const { data } = await apiClient.get<PortfolioStats>("/reporting/portfolio");
  return data;
}

export async function getRepaymentRates() {
  const { data } = await apiClient.get<{ rate: number }>(
    "/reporting/repayment-rates"
  );
  return data;
}

export async function getDefaultRates() {
  const { data } = await apiClient.get<{ rate: number }>(
    "/reporting/default-rates"
  );
  return data;
}
