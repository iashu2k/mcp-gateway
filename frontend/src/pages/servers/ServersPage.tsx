import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { listServers } from "../../api/services";
import { Server as ServerIcon, Plus } from "lucide-react";
import { useAuth } from "../../hooks/useAuth";

export default function ServersPage() {
  const { user } = useAuth();
  const {
    data: servers,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["servers"],
    queryFn: listServers,
  });

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500">Loading servers...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        Failed to load servers
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-900">MCP Servers</h1>
        {user?.role === "admin" && (
          <Link
            to="/admin/servers/new"
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
          >
            <Plus className="h-4 w-4 mr-1" />
            Add Server
          </Link>
        )}
      </div>

      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {servers?.map((server) => (
          <Link
            key={server.id}
            to={`/servers/${server.id}`}
            className="bg-white overflow-hidden shadow rounded-lg hover:shadow-md transition-shadow"
          >
            <div className="p-6">
              <div className="flex items-center">
                <ServerIcon className="h-8 w-8 text-indigo-600" />
                <div className="ml-4 flex-1">
                  <h3 className="text-lg font-medium text-gray-900">
                    {server.name}
                  </h3>
                  <p className="text-sm text-gray-500">{server.description}</p>
                </div>
              </div>
              <div className="mt-4 flex items-center justify-between">
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    server.status === "active"
                      ? "bg-green-100 text-green-800"
                      : "bg-gray-100 text-gray-800"
                  }`}
                >
                  {server.status}
                </span>
                <span className="text-sm text-gray-500">
                  {server.ownerTeam}
                </span>
              </div>
            </div>
          </Link>
        ))}
      </div>

      {servers?.length === 0 && (
        <div className="text-center py-12">
          <ServerIcon className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No servers</h3>
          <p className="mt-1 text-sm text-gray-500">
            Get started by creating a new MCP server.
          </p>
        </div>
      )}
    </div>
  );
}
