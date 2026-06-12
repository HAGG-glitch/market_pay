import apiClient from "./client";
import type { PortfolioStats } from "@/types";

export interface DashboardSummary {
  total_vendors: number;
  active_loans: number;
  overdue_loans: number;
  portfolio_value: number;
  repayment_rate: number;
  default_rate: number;
}

export interface DisbursementVolume {
  month: string;
  volume: number;
}

export interface VendorDistribution {
  district: string;
  count: number;
}

export interface OfficerQueue {
  pending_review: number;
  under_review: number;
}

export async function getDashboardSummary() {
  const { data } = await apiClient.get<DashboardSummary>(
    "/reports/dashboard-summary"
  );
  return data;
}

export async function getPortfolioStats() {
  const { data } = await apiClient.get<PortfolioStats>("/reports/portfolio");
  return data;
}

export async function getRepaymentRates() {
  const { data } = await apiClient.get<{ repayment_rate_pct: number }>(
    "/reports/repayment-rate"
  );
  return { rate: data.repayment_rate_pct };
}

export async function getDefaultRates() {
  const { data } = await apiClient.get<{ default_rate_pct: number }>(
    "/reports/default-rate"
  );
  return { rate: data.default_rate_pct };
}

export async function getDisbursementVolume() {
  const { data } = await apiClient.get<
    { month: string; count: number; volume: number }[]
  >("/reports/disbursement-volume");
  return (data ?? []).map((d) => ({
    month: d.month,
    volume: d.volume,
  }));
}

export async function getVendorDistribution() {
  const { data } = await apiClient.get<
    { market_name: string; vendor_count: number }[]
  >("/reports/vendor-distribution");
  return (data ?? []).map((d) => ({
    district: d.market_name,
    count: d.vendor_count,
  }));
}

export async function getOfficerQueue() {
  const { data } = await apiClient.get<OfficerQueue>("/reports/officer-queue");
  return data;
}

export async function getPartnerSummary() {
  const { data } = await apiClient.get<
    {
      partner_id: string;
      loans_issued: number;
      total_disbursed: number;
      total_repaid: number;
      commission_owed: number;
    }[]
  >("/reports/partner-summary");
  const rows = data ?? [];
  return {
    total_loans: rows.reduce((s, r) => s + r.loans_issued, 0),
    total_disbursed: rows.reduce((s, r) => s + r.total_disbursed, 0),
    total_repaid: rows.reduce((s, r) => s + r.total_repaid, 0),
    commission: rows.reduce((s, r) => s + r.commission_owed, 0),
  };
}
