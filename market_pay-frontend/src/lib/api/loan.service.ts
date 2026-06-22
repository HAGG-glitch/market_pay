import apiClient from "./client";
import type { Loan, LoanStatus, PaginatedResponse } from "@/types";

interface BackendLoan {
  id: string;
  vendor_id: string;
  principal_amount: number;
  interest_rate: number;
  state: string;
  source: string;
  monime_reference?: string;
  created_at: string;
  disbursed_at?: string;
  reviewed_by?: string;
  review_note?: string;
  rejection_reason?: string;
  funded_by: string | null;
  schedules?: { due_date: string; total_due: number; is_paid: boolean }[];
}

function mapLoan(loan: BackendLoan): Loan {
  return {
    id: loan.id,
    vendor_id: loan.vendor_id,
    amount: loan.principal_amount,
    interest_rate: loan.interest_rate,
    status: loan.state as LoanStatus,
    source: loan.source,
    monime_reference: loan.monime_reference,
    created_at: loan.created_at,
    disbursed_at: loan.disbursed_at,
    reviewed_by: loan.reviewed_by,
    review_note: loan.review_note,
    rejection_reason: loan.rejection_reason,
    repayment_schedule: (loan.schedules || []).map((s) => ({
      due_date: s.due_date,
      amount: s.total_due,
      paid: s.is_paid,
    })),
    funded_by: loan.funded_by,
  };
}

export async function getLoans(params?: {
  page?: number;
  limit?: number;
  status?: string;
}) {
  const { data } = await apiClient.get<PaginatedResponse<BackendLoan>>(
    "/loans",
    {
      params: {
        page: params?.page,
        limit: params?.limit,
        state: params?.status,
      },
    }
  );
  return {
    ...data,
    data: data.data.map(mapLoan),
  };
}

export async function getLoan(id: string) {
  const { data } = await apiClient.get<BackendLoan>(`/loans/${id}`);
  return mapLoan(data);
}

export async function applyLoan(payload: {
  amount: number;
  interest_rate?: number;
  repayment_schedule?: { due_date: string; amount: number }[];
}) {
  const termWeeks = payload.repayment_schedule?.length || 4;
  const { data } = await apiClient.post<BackendLoan>("/loans", {
    loan_type: "STARTER_LOAN",
    amount: payload.amount,
    term_weeks: termWeeks,
    frequency: "BIWEEKLY",
    funded_by: "MFI_PARTNER",
  });
  return mapLoan(data);
}

export async function updateLoanStatus(id: string, status: string) {
  if (status === "APPROVED") {
    const { data } = await apiClient.put<BackendLoan>(`/loans/${id}/approve`, {
      note: "",
    });
    return mapLoan(data);
  }
  if (status === "REJECTED") {
    const { data } = await apiClient.put<BackendLoan>(`/loans/${id}/reject`, {
      reason: "Rejected via dashboard",
    });
    return mapLoan(data);
  }
  throw new Error(`Unsupported loan status transition: ${status}`);
}

export async function disburseLoan(id: string, monimeReference: string) {
  const { data } = await apiClient.put<BackendLoan>(`/loans/${id}/disburse`, {
    monime_reference: monimeReference,
  });
  return mapLoan(data);
}
