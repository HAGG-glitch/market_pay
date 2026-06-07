import Link from "next/link";

export default function NotFound() {
  return (
    <div className="flex h-screen items-center justify-center p-4">
      <div className="max-w-md text-center">
        <h1 className="mb-2 text-6xl font-bold text-[#486B6D]">404</h1>
        <p className="mb-6 text-gray-500">This page could not be found.</p>
        <Link
          href="/dashboard/super-admin"
          className="inline-flex h-10 items-center rounded-md bg-[#486B6D] px-6 text-sm font-medium text-white hover:bg-[#3a5a5c]"
        >
          Go to Dashboard
        </Link>
      </div>
    </div>
  );
}
