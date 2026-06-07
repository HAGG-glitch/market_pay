"use client";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex h-screen items-center justify-center p-4">
      <div className="max-w-md text-center">
        <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
          <span className="text-2xl font-bold text-red-600">!</span>
        </div>
        <h1 className="mb-2 text-xl font-bold text-gray-900">Something went wrong</h1>
        <p className="mb-6 text-sm text-gray-500">
          {error.message || "An unexpected error occurred."}
        </p>
        <button
          onClick={reset}
          className="inline-flex h-10 items-center rounded-md bg-[#486B6D] px-6 text-sm font-medium text-white hover:bg-[#3a5a5c]"
        >
          Try Again
        </button>
      </div>
    </div>
  );
}
