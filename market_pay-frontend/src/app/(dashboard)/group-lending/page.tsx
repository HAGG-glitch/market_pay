"use client";

import { Users } from "lucide-react";

export default function GroupLendingPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] text-center px-4">
      <div className="flex h-16 w-16 items-center justify-center rounded-full bg-amber-100 mb-6">
        <Users size={32} className="text-amber-600" />
      </div>
      <h1 className="text-3xl font-bold text-gray-900 mb-2">Group Lending</h1>
      <p className="text-lg text-gray-500 mb-2">Coming Soon</p>
      <p className="text-sm text-gray-400 max-w-md">
        We are working on a group lending feature that will allow vendors to form
        groups and access collective credit. Stay tuned!
      </p>
    </div>
  );
}
