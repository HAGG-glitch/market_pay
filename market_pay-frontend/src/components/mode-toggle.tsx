"use client";

import { useModeStore } from "@/store/mode.store";

export function ModeToggle() {
  const { mode, setMode } = useModeStore();

  return (
    <div className="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm">
      <span className="text-gray-500">Mode:</span>
      <button
        type="button"
        onClick={() => setMode("demo")}
        className={`rounded px-2 py-0.5 font-medium ${
          mode === "demo"
            ? "bg-amber-100 text-amber-800"
            : "text-gray-500 hover:text-gray-700"
        }`}
      >
        Demo
      </button>
      <button
        type="button"
        onClick={() => setMode("live")}
        className={`rounded px-2 py-0.5 font-medium ${
          mode === "live"
            ? "bg-green-100 text-green-800"
            : "text-gray-500 hover:text-gray-700"
        }`}
      >
        Live
      </button>
      {mode === "demo" && (
        <span className="rounded bg-amber-50 px-1.5 py-0.5 text-xs text-amber-700">
          Demo Data
        </span>
      )}
    </div>
  );
}
