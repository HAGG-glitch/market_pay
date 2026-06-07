import type { Metadata } from "next";
import "./globals.css";
import { Providers } from "./providers";

export const metadata: Metadata = {
  title: "MarketPay - Fintech Dashboard",
  description: "MarketPay Payment Platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-full">
      <body className="h-full bg-white font-sans text-gray-900 antialiased">
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
