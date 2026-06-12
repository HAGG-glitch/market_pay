"use client";

import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { formatDate } from "@/lib/utils";
import { getAuditLogs } from "@/lib/api/audit.service";
import { useState } from "react";

const actionColors: Record<string, "success" | "warning" | "danger" | "info" | "default"> = {
  FREEZE: "danger",
  UNFREEZE: "success",
  CREATE: "success",
  UPDATE: "info",
  DELETE: "danger",
  APPROVE: "success",
  REJECT: "warning",
  DISBURSE: "info",
};

export default function AuditLogsPage() {
  const [actorFilter, setActorFilter] = useState("");
  const [resourceFilter, setResourceFilter] = useState("");

  const { data: logs = [], isLoading } = useQuery({
    queryKey: ["audit-logs", actorFilter, resourceFilter],
    queryFn: () =>
      getAuditLogs({
        actor_id: actorFilter || undefined,
        resource: resourceFilter || undefined,
      }),
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Audit Logs</h1>
        <p className="text-gray-500">Compliance audit trail</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <div className="flex-1">
              <Input
                id="actorFilter"
                label="Actor ID"
                value={actorFilter}
                onChange={(e) => setActorFilter(e.target.value)}
                placeholder="Filter by actor ID"
              />
            </div>
            <div className="flex-1">
              <Input
                id="resourceFilter"
                label="Resource"
                value={resourceFilter}
                onChange={(e) => setResourceFilter(e.target.value)}
                placeholder="Filter by resource (e.g. vendor, loan)"
              />
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Audit Trail ({logs.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-sm text-gray-500">Loading audit logs...</p>
          ) : logs.length === 0 ? (
            <p className="text-sm text-gray-500">No audit logs found.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-gray-50 text-left text-gray-500">
                    <th className="pb-3 pt-3 pl-4 font-medium">Date</th>
                    <th className="pb-3 pt-3 font-medium">Actor</th>
                    <th className="pb-3 pt-3 font-medium">Role</th>
                    <th className="pb-3 pt-3 font-medium">Action</th>
                    <th className="pb-3 pt-3 font-medium">Resource</th>
                    <th className="pb-3 pt-3 font-medium">Resource ID</th>
                    <th className="pb-3 pt-3 pr-4 font-medium">Details</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log, i) => (
                    <tr
                      key={log.id}
                      className={`border-b last:border-0 transition-colors hover:bg-gray-100 ${
                        i % 2 === 0 ? "bg-white" : "bg-gray-50/50"
                      }`}
                    >
                      <td className="py-3 pl-4 text-xs text-gray-500">
                        {formatDate(log.created_at)}
                      </td>
                      <td className="py-3 font-mono text-xs">
                        {log.actor_id.slice(0, 8)}...
                      </td>
                      <td className="py-3">
                        <Badge variant="default">{log.actor_role}</Badge>
                      </td>
                      <td className="py-3">
                        <Badge variant={actionColors[log.action] || "default"}>
                          {log.action}
                        </Badge>
                      </td>
                      <td className="py-3 text-gray-700">{log.resource}</td>
                      <td className="py-3 font-mono text-xs text-gray-500">
                        {log.resource_id.slice(0, 8)}...
                      </td>
                      <td className="py-3 pr-4 text-xs text-gray-500">
                        {log.old_state && (
                          <span className="text-red-600">{log.old_state}</span>
                        )}
                        {log.old_state && log.new_state && " → "}
                        {log.new_state && (
                          <span className="text-green-600">{log.new_state}</span>
                        )}
                        {!log.old_state && !log.new_state && "—"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
