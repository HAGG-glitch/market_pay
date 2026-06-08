# MarketPay - Fintech Dashboard

A production-ready fintech dashboard and USSD companion web system for MarketPay.

## Tech Stack

- **Next.js 14+** (App Router)
- **TypeScript**
- **TailwindCSS v4**
- **TanStack Query** (React Query)
- **Axios** (API client)
- **Zustand** (global state)
- **Recharts** (analytics)
- **Lucide Icons**

## Design System

- Primary: `#486B6D`
- Accent: `#A98881`
- Font: Inter

## Getting Started

```bash
npm install
npm run dev
```

## User Roles

| Role | Dashboard |
|------|-----------|
| SUPER_ADMIN | Full portfolio overview |
| ADMIN | Loan portfolio management |
| LOAN_OFFICER | Loan approval queue |
| FIELD_AGENT | Vendor onboarding |
| VENDOR | Loan & repayment management |
| CUSTOMER | Payment portal |
| MFI_PARTNER | Portfolio overview |

## Project Structure

```
src/
├── app/            # App Router pages
│   ├── login/      # Authentication
│   ├── dashboard/  # Role-based dashboards
│   ├── loans/      # Loan management
│   ├── payments/   # Payment system
│   ├── vendors/    # Vendor onboarding
│   ├── group-lending/ # Group lending
│   └── analytics/  # Reports & charts
├── components/     # Reusable UI components
│   ├── ui/         # Base components (Button, Card, Input, etc.)
│   ├── layout/     # Sidebar, Navbar
│   └── dashboard/  # Dashboard widgets
├── lib/
│   ├── api/        # Service layer (Axios)
│   └── utils.ts    # Helpers
├── store/          # Zustand stores
├── hooks/          # React Query hooks
├── types/          # TypeScript definitions
└── providers/      # React Query provider
```

## Environment Variables

Create a `.env.local` file:

```env
NEXT_PUBLIC_API_URL=https://api.marketpay.local
```

## Deployment

Deploy to Vercel:

```bash
vercel
```
