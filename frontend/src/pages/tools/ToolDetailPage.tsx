import { useState } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import { getTool, invokeTool } from "../../api/services";
import { ArrowLeft, Play } from "lucide-react";

export default function ToolDetailPage() {
  const { serverId, toolId } = useParams<{
    serverId: string;
    toolId: string;
  }>();
  const [argsJson, setArgsJson] = useState("{}");
  const [result, setResult] = useState<any>(null);

  const { data: tool, isLoading } = useQuery({
    queryKey: ["tool", serverId, toolId],
    queryFn: () => getTool(serverId!, toolId!),
    enabled: !!serverId && !!toolId,
  });

  const invokeMutation = useMutation({
    mutationFn: (args: Record<string, any>) =>
      invokeTool(serverId!, toolId!, args),
    onSuccess: (data) => {
      setResult(data);
    },
    onError: (error: any) => {
      setResult({
        status: "error",
        error: error.response?.data?.message || "Invocation failed",
      });
    },
  });

  const handleInvoke = () => {
    try {
      const args = JSON.parse(argsJson);
      invokeMutation.mutate(args);
    } catch (err) {
      alert("Invalid JSON");
    }
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  if (!tool) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        Tool not found
      </div>
    );
  }

  return (
    <div>
      <Link
        to={`/servers/${serverId}`}
        className="inline-flex items-center text-sm text-indigo-600 hover:text-indigo-900 mb-4"
      >
        <ArrowLeft className="h-4 w-4 mr-1" />
        Back to server
      </Link>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Tool info */}
        <div className="bg-white shadow overflow-hidden sm:rounded-lg">
          <div className="px-4 py-5 sm:px-6">
            <h3 className="text-2xl font-bold text-gray-900">
              {tool.title || tool.name}
            </h3>
            <p className="mt-1 text-sm text-gray-500">{tool.description}</p>
          </div>
          <div className="border-t border-gray-200 px-4 py-5 sm:px-6">
            <dl className="space-y-4">
              <div>
                <dt className="text-sm font-medium text-gray-500">
                  Risk Level
                </dt>
                <dd className="mt-1">
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      tool.riskLevel === "low"
                        ? "bg-green-100 text-green-800"
                        : tool.riskLevel === "medium"
                          ? "bg-yellow-100 text-yellow-800"
                          : "bg-red-100 text-red-800"
                    }`}
                  >
                    {tool.riskLevel}
                  </span>
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Status</dt>
                <dd className="mt-1">
                  <span
                    className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      tool.enabled
                        ? "bg-green-100 text-green-800"
                        : "bg-gray-100 text-gray-800"
                    }`}
                  >
                    {tool.enabled ? "Enabled" : "Disabled"}
                  </span>
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">
                  Input Schema
                </dt>
                <dd className="mt-1">
                  <pre className="text-xs bg-gray-50 p-3 rounded overflow-auto max-h-64">
                    {JSON.stringify(tool.inputSchema, null, 2)}
                  </pre>
                </dd>
              </div>
            </dl>
          </div>
        </div>

        {/* Invoke sandbox */}
        <div className="bg-white shadow overflow-hidden sm:rounded-lg">
          <div className="px-4 py-5 sm:px-6">
            <h3 className="text-lg font-medium text-gray-900">Invoke Tool</h3>
          </div>
          <div className="border-t border-gray-200 px-4 py-5 sm:px-6">
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Arguments (JSON)
                </label>
                <textarea
                  value={argsJson}
                  onChange={(e) => setArgsJson(e.target.value)}
                  rows={8}
                  className="w-full font-mono text-sm border border-gray-300 rounded-md p-3 focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder='{"owner": "golang", "repo": "go"}'
                />
              </div>
              <button
                onClick={handleInvoke}
                disabled={invokeMutation.isPending || !tool.enabled}
                className="w-full inline-flex justify-center items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50"
              >
                <Play className="h-4 w-4 mr-2" />
                {invokeMutation.isPending ? "Invoking..." : "Invoke Tool"}
              </button>
            </div>

            {result && (
              <div className="mt-6">
                <h4 className="text-sm font-medium text-gray-700 mb-2">
                  Result
                </h4>
                <pre className="text-xs bg-gray-50 p-3 rounded overflow-auto max-h-96">
                  {JSON.stringify(result, null, 2)}
                </pre>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
