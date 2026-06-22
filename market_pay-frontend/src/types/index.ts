export enum UserRole {
  SUPER_ADMIN = "SUPER_ADMIN",
  ADMIN = "ADMIN",
  LOAN_OFFICER = "LOAN_OFFICER",
  FIELD_AGENT = "FIELD_AGENT",
  VENDOR = "VENDOR",
  CUSTOMER = "CUSTOMER",
  MFI_PARTNER = "MFI_PARTNER",
}

export interface User {
  id: string;
  name: string;
  email: string;
  phone: string;
  role: UserRole;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
}

export interface Vendor {
  id: string;
  name: string;
  phone: string;
  kyc_status: string;
  group_id: string | null;
  credit_score: number;
}

export interface Loan {
  id: string;
  vendor_id: string;
  amount: number;
  interest_rate: number;
  status: LoanStatus;
  source: string;
  monime_reference?: string;
  created_at: string;
  disbursed_at?: string;
  reviewed_by?: string;
  review_note?: string;
  rejection_reason?: string;
  repayment_schedule: RepaymentSchedule[];
  funded_by: string | null;
}

export enum LoanStatus {
  DRAFT = "DRAFT",
  PENDING_REVIEW = "PENDING_REVIEW",
  UNDER_REVIEW = "UNDER_REVIEW",
  APPROVED = "APPROVED",
  REJECTED = "REJECTED",
  DISBURSED = "DISBURSED",
  ACTIVE = "ACTIVE",
  CLOSED = "CLOSED",
  DEFAULTED = "DEFAULTED",
}

export interface RepaymentSchedule {
  due_date: string;
  amount: number;
  paid: boolean;
}

export interface Payment {
  id: string;
  vendor_id: string;
  customer_id: string;
  amount: number;
  fee: number;
  status: PaymentStatus;
  created_at: string;
}

export enum PaymentStatus {
  PENDING = "PENDING",
  SUCCESS = "SUCCESS",
  FAILED = "FAILED",
}

export interface PortfolioStats {
  total_loans: number;
  total_disbursed: number;
  repayment_rate: number;
  default_rate: number;
  active_loans: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export interface NotificationItem {
  id: string;
  recipient_id: string;
  event_type: string;
  title: string;
  body: string;
  is_read: boolean;
  is_demo: boolean;
  metadata: Record<string, unknown>;
  created_at: string;
}
