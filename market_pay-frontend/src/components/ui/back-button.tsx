"use client";

import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { Button } from "./button";

interface BackButtonProps {
  href?: string;
  label?: string;
}

export function BackButton({ href, label = "Back" }: BackButtonProps) {
  const router = useRouter();

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={() => (href ? router.push(href) : router.back())}
      aria-label={`Navigate back to ${label.toLowerCase()}`}
      className="gap-1.5"
    >
      <ArrowLeft size={16} aria-hidden="true" />
      {label}
    </Button>
  );
}
