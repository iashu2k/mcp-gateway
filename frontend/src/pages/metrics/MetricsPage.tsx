import { useQuery } from "@tanstack/react-query";
import { getMetrics } from "../../api/services";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
} from "recharts";
import { Activity } from "lucide-react";

const COLORS = ["#4f46e5", "#22c55e", "#ef4444", "#eab308", "#8b5cf6"];

export default function MetricsPage() {
  const { data: metrics, isLoading } = useQuery({
    queryKey: ["metrics"],
    queryFn: getMetrics,
    refetchInterval: 15000, // refresh every 15s
  });

  if (isLoading || !metrics) {
    return <div className="text-gray-500">Loading metrics...</div>;
  }

  // Aggregate invocations by tool
  const invocationsByTool = new Map<string, number>();
  const invocationsByStatus = new Map<string, number>();
  let totalInvocations = 0;
  let failedInvocations = 0;

  for (const m of metrics) {
    if (m.name === "mcp_gateway_invocations_total") {
      const key = `${m.labels.server}/${m.labels.tool}`;
      invocationsByTool.set(key, (invocationsByTool.get(key) || 0) + m.value);
      invocationsByStatus.set(
        m.labels.status,
        (invocationsByStatus.get(m.labels.status) || 0) + m.value,
      );
      totalInvocations += m.value;
      if (m.labels.status === "failed") failedInvocations += m.value;
    }
  }

  const toolChartData = Array.from(invocationsByTool.entries()).map(
    ([name, count]) => ({
      name,
      count,
    }),
  );

  const statusChartData = Array.from(invocationsByStatus.entries()).map(
    ([name, value]) => ({
      name,
      value,
    }),
  );

  const successRate =
    totalInvocations > 0
      ? (
          ((totalInvocations - failedInvocations) / totalInvocations) *
          100
        ).toFixed(1)
      : "100.0";

  const dbConns =
    metrics.find((m) => m.name === "mcp_gateway_database_connections_open")
      ?.value ?? 0;

  return (
    <div>
      <div className="flex items-center mb-6">
        <Activity className="h-8 w-8 text-indigo-600 mr-3" />
        <h1 className="text-3xl font-bold text-gray-900">Metrics Dashboard</h1>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-8">
        <div className="bg-white shadow rounded-lg p-6">
          <div className="text-sm font-medium text-gray-500">
            Total Invocations
          </div>
          <div className="mt-2 text-3xl font-bold text-gray-900">
            {totalInvocations}
          </div>
        </div>
        <div className="bg-white shadow rounded-lg p-6">
          <div className="text-sm font-medium text-gray-500">Success Rate</div>
          <div className="mt-2 text-3xl font-bold text-green-600">
            {successRate}%
          </div>
        </div>
        <div className="bg-white shadow rounded-lg p-6">
          <div className="text-sm font-medium text-gray-500">
            DB Connections
          </div>
          <div className="mt-2 text-3xl font-bold text-gray-900">{dbConns}</div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Invocations by tool */}
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            Invocations by Tool
          </h3>
          {toolChartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={toolChartData}>
                <XAxis dataKey="name" fontSize={12} />
                <YAxis allowDecimals={false} />
                <Tooltip />
                <Bar dataKey="count" fill="#4f46e5" />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="text-gray-500 text-center py-12">
              No invocation data yet
            </div>
          )}
        </div>

        {/* Status distribution */}
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">
            Status Distribution
          </h3>
          {statusChartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={statusChartData}
                  dataKey="value"
                  nameKey="name"
                  cx="50%"
                  cy="50%"
                  outerRadius={100}
                  label
                >
                  {statusChartData.map((_, i) => (
                    <Cell key={i} fill={COLORS[i % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <div className="text-gray-500 text-center py-12">
              No status data yet
            </div>
          )}
        </div>
      </div>

      <p className="mt-4 text-xs text-gray-400">
        Metrics refresh every 15 seconds from the Prometheus /metrics endpoint.
      </p>
    </div>
  );
}
