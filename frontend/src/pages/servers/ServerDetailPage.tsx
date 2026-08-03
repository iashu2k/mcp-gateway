import { useQuery } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import { getServer, listTools } from "../../api/services";
import { Server as ServerIcon, Wrench as Tool, ArrowLeft } from "lucide-react";

export default function ServerDetailPage() {
  const { serverId } = useParams<{ serverId: string }>();

  const { data: server, isLoading: serverLoading } = useQuery({
    queryKey: ["server", serverId],
    queryFn: () => getServer(serverId!),
    enabled: !!serverId,
  });

  const { data: tools, isLoading: toolsLoading } = useQuery({
    queryKey: ["tools", serverId],
    queryFn: () => listTools(serverId!),
    enabled: !!serverId,
  });

  if (serverLoading || toolsLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  if (!server) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        Server not found
      </div>
    );
  }

  return (
    <div>
      <Link
        to="/servers"
        className="inline-flex items-center text-sm text-indigo-600 hover:text-indigo-900 mb-4"
      >
        <ArrowLeft className="h-4 w-4 mr-1" />
        Back to servers
      </Link>

      <div className="bg-white shadow overflow-hidden sm:rounded-lg mb-6">
        <div className="px-4 py-5 sm:px-6">
          <div className="flex items-center">
            <ServerIcon className="h-8 w-8 text-indigo-600 mr-3" />
            <div>
              <h3 className="text-2xl font-bold text-gray-900">
                {server.name}
              </h3>
              <p className="mt-1 text-sm text-gray-500">{server.description}</p>
            </div>
          </div>
        </div>
        <div className="border-t border-gray-200 px-4 py-5 sm:px-6">
          <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-3">
            <div>
              <dt className="text-sm font-medium text-gray-500">Status</dt>
              <dd className="mt-1">
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    server.status === "active"
                      ? "bg-green-100 text-green-800"
                      : "bg-gray-100 text-gray-800"
                  }`}
                >
                  {server.status}
                </span>
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Base URL</dt>
              <dd className="mt-1 text-sm text-gray-900">{server.baseUrl}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Owner Team</dt>
              <dd className="mt-1 text-sm text-gray-900">{server.ownerTeam}</dd>
            </div>
          </dl>
        </div>
      </div>

      <div className="bg-white shadow overflow-hidden sm:rounded-lg">
        <div className="px-4 py-5 sm:px-6">
          <h3 className="text-lg font-medium text-gray-900">Tools</h3>
        </div>
        <ul className="divide-y divide-gray-200">
          {tools?.map((tool) => (
            <li key={tool.id}>
              <Link
                to={`/servers/${serverId}/tools/${tool.id}`}
                className="block hover:bg-gray-50 px-4 py-4 sm:px-6"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center">
                    <Tool className="h-5 w-5 text-gray-400 mr-3" />
                    <div>
                      <p className="text-sm font-medium text-indigo-600">
                        {tool.title || tool.name}
                      </p>
                      <p className="text-sm text-gray-500">
                        {tool.description}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
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
                    <span
                      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                        tool.enabled
                          ? "bg-green-100 text-green-800"
                          : "bg-gray-100 text-gray-800"
                      }`}
                    >
                      {tool.enabled ? "Enabled" : "Disabled"}
                    </span>
                  </div>
                </div>
              </Link>
            </li>
          ))}
        </ul>
        {tools?.length === 0 && (
          <div className="text-center py-12">
            <Tool className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">No tools</h3>
            <p className="mt-1 text-sm text-gray-500">
              This server has no registered tools.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
