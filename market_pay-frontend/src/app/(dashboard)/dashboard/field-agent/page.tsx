"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useState } from "react";
import { useRouter } from "next/navigation";

export default function FieldAgentDashboard() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [phone, setPhone] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    router.push("/vendors/onboard");
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Field Agent Portal</h1>
        <p className="text-gray-500">Onboard vendors and manage groups</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Quick Vendor Onboard</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              id="name"
              label="Vendor Name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter vendor name"
              required
            />
            <Input
              id="phone"
              label="Phone Number"
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="+234 800 000 0000"
              required
            />
            <Button type="submit" className="w-full" size="lg" aria-label="Start vendor onboarding process">
              Start Onboarding
            </Button>
          </form>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2">
        <Card
          className="cursor-pointer transition-all duration-200 hover:shadow-lg hover:-translate-y-0.5"
          onClick={() => router.push("/group-lending")}
          role="button"
          tabIndex={0}
          aria-label="Go to group management"
          onKeyDown={(e) => e.key === "Enter" && router.push("/group-lending")}
        >
          <CardHeader>
            <CardTitle>Group Management</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-gray-500">
              Create and manage vendor lending groups
            </p>
          </CardContent>
        </Card>
        <Card
          className="cursor-pointer transition-all duration-200 hover:shadow-lg hover:-translate-y-0.5"
          role="button"
          tabIndex={0}
          aria-label="Review KYC submissions"
        >
          <CardHeader>
            <CardTitle>KYC Submissions</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-gray-500">
              Review pending KYC verification requests
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
