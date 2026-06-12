"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { onboardVendor } from "@/lib/api/vendor.service";
import apiClient from "@/lib/api/client";

export default function VendorOnboardPage() {
  const router = useRouter();
  const [step, setStep] = useState(1);
  const [form, setForm] = useState({
    name: "",
    phone: "",
    pin: "",
    nationalIdNumber: "",
    nationalIdType: "NATIONAL_ID",
    dateOfBirth: "",
    marketAssociationId: "",
    businessName: "",
    address: "",
  });
  const [error, setError] = useState("");

  const { data: markets = [] } = useQuery({
    queryKey: ["market-associations"],
    queryFn: async () => {
      const { data } = await apiClient.get("/vendors/market-associations");
      return data as { id: string; name: string; district: string }[];
    },
  });

  const mutation = useMutation({
    mutationFn: () =>
      onboardVendor({
        name: form.name,
        phone: form.phone,
        pin: form.pin,
        national_id_number: form.nationalIdNumber,
        national_id_type: form.nationalIdType,
        date_of_birth: new Date(form.dateOfBirth).toISOString(),
        market_association_id: form.marketAssociationId,
        business_name: form.businessName,
        address: form.address,
      }),
    onSuccess: () => {
      router.push("/dashboard/field-agent");
    },
    onError: (err: Error) => {
      setError(err.message || "Failed to create vendor");
    },
  });

  const handleChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setForm((prev) => ({ ...prev, [field]: e.target.value }));
    setError("");
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    mutation.mutate();
  };

  return (
    <div className="mx-auto max-w-lg space-y-6 animate-fade-in">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Vendor Onboarding</h1>
        <p className="text-gray-500">KYC submission form</p>
      </div>

      {error && (
        <div className="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>
            {step === 1 ? "Personal Information" : "Business Details & PIN"}
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
                  id="nationalIdNumber"
                  label="National ID Number"
                  value={form.nationalIdNumber}
                  onChange={handleChange("nationalIdNumber")}
                  placeholder="Enter national ID number"
                  required
                />
                <Select
                  id="nationalIdType"
                  label="ID Type"
                  value={form.nationalIdType}
                  onChange={handleChange("nationalIdType")}
                  options={[
                    { value: "NATIONAL_ID", label: "National ID" },
                    { value: "PASSPORT", label: "Passport" },
                    { value: "DRIVERS_LICENSE", label: "Driver's License" },
                    { value: "VOTER_ID", label: "Voter ID" },
                  ]}
                />
                <Input
                  id="dateOfBirth"
                  label="Date of Birth"
                  type="date"
                  value={form.dateOfBirth}
                  onChange={handleChange("dateOfBirth")}
                  required
                />
                <Select
                  id="marketAssociationId"
                  label="Market Association"
                  value={form.marketAssociationId}
                  onChange={handleChange("marketAssociationId")}
                  options={[
                    { value: "", label: "Select a market" },
                    ...markets.map((m) => ({
                      value: m.id,
                      label: `${m.name} - ${m.district}`,
                    })),
                  ]}
                  required
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
                  id="address"
                  label="Business Address"
                  value={form.address}
                  onChange={handleChange("address")}
                  placeholder="Enter business address"
                  required
                />
                <Input
                  id="pin"
                  label="4-digit PIN (for phone login)"
                  type="password"
                  maxLength={4}
                  value={form.pin}
                  onChange={handleChange("pin")}
                  placeholder="1234"
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
                  <Button
                    type="submit"
                    className="flex-1"
                    size="lg"
                    disabled={mutation.isPending}
                  >
                    {mutation.isPending ? "Submitting..." : "Submit KYC"}
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
