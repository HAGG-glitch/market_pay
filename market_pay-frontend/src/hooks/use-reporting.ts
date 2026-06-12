"use client";

import { useQuery } from "@tanstack/react-query";
import {
  getDashboardSummary,
  getPortfolioStats,
  getDisbursementVolume,
  getVendorDistribution,
  getOfficerQueue,
  getPartnerSummary,
  getRepaymentRates,
  getDefaultRates,
} from "@/lib/api/reporting.service";

export function useDashboardSummary() {
  return useQuery({ queryKey: ["dashboard-summary"], queryFn: getDashboardSummary });
}

export function usePortfolioStats() {
  return useQuery({ queryKey: ["portfolio"], queryFn: getPortfolioStats });
}

export function useDisbursementVolume() {
  return useQuery({ queryKey: ["disbursement-volume"], queryFn: getDisbursementVolume });
}

export function useVendorDistribution() {
  return useQuery({ queryKey: ["vendor-distribution"], queryFn: getVendorDistribution });
}

export function useOfficerQueue() {
  return useQuery({ queryKey: ["officer-queue"], queryFn: getOfficerQueue });
}

export function usePartnerSummary() {
  return useQuery({ queryKey: ["partner-summary"], queryFn: getPartnerSummary });
}

export function useRepaymentRate() {
  return useQuery({ queryKey: ["repayment-rate"], queryFn: getRepaymentRates });
}

export function useDefaultRate() {
  return useQuery({ queryKey: ["default-rate"], queryFn: getDefaultRates });
}
