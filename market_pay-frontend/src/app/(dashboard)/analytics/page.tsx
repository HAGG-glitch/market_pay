"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  LineChart,
  Line,
} from "recharts";

const loanData = [
  { month: "Jan", disbursed: 450000, repaid: 320000 },
  { month: "Feb", disbursed: 520000, repaid: 410000 },
  { month: "Mar", disbursed: 610000, repaid: 480000 },
  { month: "Apr", disbursed: 480000, repaid: 520000 },
  { month: "May", disbursed: 720000, repaid: 610000 },
  { month: "Jun", disbursed: 680000, repaid: 590000 },
];

const portfolioComposition = [
  { name: "Active", value: 65, color: "#486B6D" },
  { name: "Closed", value: 20, color: "#A98881" },
  { name: "Defaulted", value: 8, color: "#ef4444" },
  { name: "Pending", value: 7, color: "#eab308" },
];

const repaymentTrend = [
  { month: "Jan", rate: 92 },
  { month: "Feb", rate: 88 },
  { month: "Mar", rate: 94 },
  { month: "Apr", rate: 91 },
  { month: "May", rate: 96 },
  { month: "Jun", rate: 93 },
];

export default function AnalyticsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Analytics</h1>
        <p className="text-gray-500">Portfolio performance and insights</p>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Total Portfolio
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold">LE 3.46M</p>
            <p className="text-xs text-green-600">+12.3% from last month</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Repayment Rate
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-green-600">93%</p>
            <p className="text-xs text-green-600">+2.1% from last month</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-gray-500">
              Default Rate
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold text-red-600">8%</p>
            <p className="text-xs text-red-600">-0.5% from last month</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Disbursements vs Repayments</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={loanData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="month" fontSize={12} />
                <YAxis fontSize={12} />
                <Tooltip />
                <Bar
                  dataKey="disbursed"
                  fill="#486B6D"
                  radius={[4, 4, 0, 0]}
                  name="Disbursed"
                />
                <Bar
                  dataKey="repaid"
                  fill="#A98881"
                  radius={[4, 4, 0, 0]}
                  name="Repaid"
                />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Portfolio Composition</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={portfolioComposition}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  paddingAngle={4}
                  dataKey="value"
                >
                  {portfolioComposition.map((entry, i) => (
                    <Cell key={i} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
            <div className="mt-4 flex justify-center gap-4">
              {portfolioComposition.map((entry) => (
                <div key={entry.name} className="flex items-center gap-1.5">
                  <div
                    className="h-3 w-3 rounded-full"
                    style={{ backgroundColor: entry.color }}
                  />
                  <span className="text-xs text-gray-600">
                    {entry.name} ({entry.value}%)
                  </span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Repayment Rate Trend</CardTitle>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={repaymentTrend}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="month" fontSize={12} />
              <YAxis domain={[80, 100]} fontSize={12} />
              <Tooltip />
              <Line
                type="monotone"
                dataKey="rate"
                stroke="#486B6D"
                strokeWidth={2}
                dot={{ fill: "#486B6D", r: 4 }}
                name="Repayment Rate %"
              />
            </LineChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>
    </div>
  );
}
