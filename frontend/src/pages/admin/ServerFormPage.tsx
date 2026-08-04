import { useState, useEffect } from "react";
import { useNavigate, useParams, Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getServer, createServer, updateServer } from "../../api/services";
import { ArrowLeft } from "lucide-react";
import type { MCPServer } from "../../types";

type ServerFormData = {
  name: string;
  description: string;
  baseUrl: string;
  transportType: MCPServer["transportType"];
  status: MCPServer["status"];
  ownerTeam: string;
};

export default function ServerFormPage() {
  const { serverId } = useParams();
  const isEdit = !!serverId;
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [form, setForm] = useState<ServerFormData>({
    name: "",
    description: "",
    baseUrl: "",
    transportType: "streamable_http",
    status: "active",
    ownerTeam: "",
  });
  const [error, setError] = useState("");

  const { data: existing } = useQuery({
    queryKey: ["server", serverId],
    queryFn: () => getServer(serverId!),
    enabled: isEdit,
  });

  useEffect(() => {
    if (existing) {
      setForm({
        name: existing.name,
        description: existing.description,
        baseUrl: existing.baseUrl,
        transportType: existing.transportType,
        status: existing.status,
        ownerTeam: existing.ownerTeam,
      });
    }
  }, [existing]);

  const saveMutation = useMutation({
    mutationFn: (data: ServerFormData) =>
      isEdit ? updateServer(serverId!, data) : createServer(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["servers"] });
      navigate("/admin/servers");
    },
    onError: (err: any) => {
      setError(err.response?.data?.message || "Failed to save server");
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    saveMutation.mutate(form);
  };

  const set =
    (field: keyof ServerFormData) =>
    (
      e: React.ChangeEvent<
        HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
      >,
    ) =>
      setForm({
        ...form,
        [field]: e.target.value as ServerFormData[typeof field],
      });

  return (
    <div className="max-w-2xl">
      <Link
        to="/admin/servers"
        className="inline-flex items-center text-sm text-indigo-600 hover:text-indigo-900 mb-4"
      >
        <ArrowLeft className="h-4 w-4 mr-1" />
        Back to servers
      </Link>

      <h1 className="text-3xl font-bold text-gray-900 mb-6">
        {isEdit ? "Edit Server" : "Add Server"}
      </h1>

      <form
        onSubmit={handleSubmit}
        className="bg-white shadow sm:rounded-lg p-6 space-y-4"
      >
        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
            {error}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Name *
          </label>
          <input
            type="text"
            required
            value={form.name}
            onChange={set("name")}
            className="w-full border border-gray-300 rounded-md px-3 py-2 focus:ring-indigo-500 focus:border-indigo-500"
            placeholder="github"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Description
          </label>
          <textarea
            value={form.description}
            onChange={set("description")}
            rows={2}
            className="w-full border border-gray-300 rounded-md px-3 py-2 focus:ring-indigo-500 focus:border-indigo-500"
            placeholder="GitHub REST API integration"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Base URL
          </label>
          <input
            type="text"
            value={form.baseUrl}
            onChange={set("baseUrl")}
            className="w-full border border-gray-300 rounded-md px-3 py-2 focus:ring-indigo-500 focus:border-indigo-500"
            placeholder="https://api.github.com"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Transport
            </label>
            <select
              value={form.transportType}
              onChange={set("transportType")}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            >
              <option value="streamable_http">streamable_http</option>
              <option value="sse">sse</option>
              <option value="stdio">stdio</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Status
            </label>
            <select
              value={form.status}
              onChange={set("status")}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            >
              <option value="active">active</option>
              <option value="inactive">inactive</option>
              <option value="unhealthy">unhealthy</option>
            </select>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Owner Team
          </label>
          <input
            type="text"
            value={form.ownerTeam}
            onChange={set("ownerTeam")}
            className="w-full border border-gray-300 rounded-md px-3 py-2 focus:ring-indigo-500 focus:border-indigo-500"
            placeholder="platform"
          />
        </div>

        <div className="flex justify-end space-x-3 pt-4">
          <Link
            to="/admin/servers"
            className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={saveMutation.isPending}
            className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
          >
            {saveMutation.isPending
              ? "Saving..."
              : isEdit
                ? "Update Server"
                : "Create Server"}
          </button>
        </div>
      </form>
    </div>
  );
}
