"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function VendorOnboardPage() {
  const router = useRouter();
  const [step, setStep] = useState(1);
  const [form, setForm] = useState({
    name: "",
    phone: "",
    email: "",
    businessName: "",
    businessAddress: "",
  });

  const handleChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [field]: e.target.value }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    router.push("/dashboard/field-agent");
  };

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Vendor Onboarding</h1>
        <p className="text-gray-500">KYC submission form</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>
            {step === 1 ? "Personal Information" : "Business Details"}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {step === 1 ? (
              <>
                <Input
                  id="name"
                  label="Full Name"
                  value={form.name}
                  onChange={handleChange("name")}
                  placeholder="Enter vendor name"
                  required
                />
                <Input
                  id="phone"
                  label="Phone Number"
                  type="tel"
                  value={form.phone}
                  onChange={handleChange("phone")}
                  placeholder="+234 800 000 0000"
                  required
                />
                <Input
                  id="email"
                  label="Email Address"
                  type="email"
                  value={form.email}
                  onChange={handleChange("email")}
                  placeholder="vendor@example.com"
                />
                <Button
                  type="button"
                  className="w-full"
                  size="lg"
                  onClick={() => setStep(2)}
                >
                  Next
                </Button>
              </>
            ) : (
              <>
                <Input
                  id="businessName"
                  label="Business Name"
                  value={form.businessName}
                  onChange={handleChange("businessName")}
                  placeholder="Enter business name"
                  required
                />
                <Input
                  id="businessAddress"
                  label="Business Address"
                  value={form.businessAddress}
                  onChange={handleChange("businessAddress")}
                  placeholder="Enter business address"
                  required
                />
                <div className="flex gap-3">
                  <Button
                    type="button"
                    variant="outline"
                    className="flex-1"
                    size="lg"
                    onClick={() => setStep(1)}
                  >
                    Back
                  </Button>
                  <Button type="submit" className="flex-1" size="lg">
                    Submit KYC
                  </Button>
                </div>
              </>
            )}
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
