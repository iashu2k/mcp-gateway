import { useState, useEffect } from "react";
import { useNavigate, useParams, Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getTool, createTool, updateTool } from "../../api/services";
import { ArrowLeft } from "lucide-react";

const DEFAULT_SCHEMA = JSON.stringify(
  {
    type: "object",
    properties: {},
    additionalProperties: false,
  },
  null,
  2,
);

export default function ToolFormPage() {
  const { serverId, toolId } = useParams();
  const isEdit = !!toolId;
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [form, setForm] = useState({
    name: "",
    title: "",
    description: "",
    riskLevel: "low",
    enabled: true,
  });
  const [schemaJson, setSchemaJson] = useState(DEFAULT_SCHEMA);
  const [error, setError] = useState("");

  const { data: existing } = useQuery({
    queryKey: ["tool", serverId, toolId],
    queryFn: () => getTool(serverId!, toolId!),
    enabled: isEdit,
  });

  useEffect(() => {
    if (existing) {
      setForm({
        name: existing.name,
        title: existing.title || "",
        description: existing.description,
        riskLevel: existing.riskLevel,
        enabled: existing.enabled,
      });
      setSchemaJson(JSON.stringify(existing.inputSchema, null, 2));
    }
  }, [existing]);

  const saveMutation = useMutation({
    mutationFn: (data: any) =>
      isEdit
        ? updateTool(serverId!, toolId!, data)
        : createTool(serverId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tools", serverId] });
      navigate(`/servers/${serverId}`);
    },
    onError: (err: any) => {
      setError(err.response?.data?.message || "Failed to save tool");
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    let inputSchema;
    try {
      inputSchema = JSON.parse(schemaJson);
    } catch {
      setError("Input schema is not valid JSON");
      return;
    }

    saveMutation.mutate({ ...form, inputSchema });
  };

  return (
    <div className="max-w-2xl">
      <Link
        to={`/servers/${serverId}`}
        className="inline-flex items-center text-sm text-indigo-600 hover:text-indigo-900 mb-4"
      >
        <ArrowLeft className="h-4 w-4 mr-1" />
        Back to server
      </Link>

      <h1 className="text-3xl font-bold text-gray-900 mb-6">
        {isEdit ? "Edit Tool" : "Add Tool"}
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

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Name *
            </label>
            <input
              type="text"
              required
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
              placeholder="list_issues"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Title
            </label>
            <input
              type="text"
              value={form.title}
              onChange={(e) => setForm({ ...form, title: e.target.value })}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
              placeholder="List GitHub Issues"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Description
          </label>
          <textarea
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            rows={2}
            className="w-full border border-gray-300 rounded-md px-3 py-2"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Risk Level
            </label>
            <select
              value={form.riskLevel}
              onChange={(e) =>
                setForm({ ...form, riskLevel: e.target.value as any })
              }
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            >
              <option value="low">low</option>
              <option value="medium">medium</option>
              <option value="high">high</option>
            </select>
          </div>
          <div className="flex items-end pb-2">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={form.enabled}
                onChange={(e) =>
                  setForm({ ...form, enabled: e.target.checked })
                }
                className="h-4 w-4 text-indigo-600 border-gray-300 rounded"
              />
              <span className="ml-2 text-sm text-gray-700">Enabled</span>
            </label>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Input Schema (JSON Schema)
          </label>
          <textarea
            value={schemaJson}
            onChange={(e) => setSchemaJson(e.target.value)}
            rows={12}
            className="w-full font-mono text-xs border border-gray-300 rounded-md p-3"
            spellCheck={false}
          />
        </div>

        <div className="flex justify-end space-x-3 pt-4">
          <Link
            to={`/servers/${serverId}`}
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
                ? "Update Tool"
                : "Create Tool"}
          </button>
        </div>
      </form>
    </div>
  );
}
