import { apiClient } from './client';
import type {
  LoginResponse,
  MCPServer,
  MCPTool,
  ToolInvocation,
  InvocationResponse,
  User,
} from '../types';

// Auth
export const login = async (email: string, password: string): Promise<LoginResponse> => {
  const { data } = await apiClient.post<LoginResponse>('/api/v1/auth/login', {
    email,
    password,
  });
  return data;
};

export const getCurrentUser = async (): Promise<User> => {
  const { data } = await apiClient.get<User>('/api/v1/auth/me');
  return data;
};

// Servers
export const listServers = async (): Promise<MCPServer[]> => {
  const { data } = await apiClient.get<{ data: MCPServer[] }>('/api/v1/servers');
  return data.data;
};

export const getServer = async (serverId: string): Promise<MCPServer> => {
  const { data } = await apiClient.get<MCPServer>(`/api/v1/servers/${serverId}`);
  return data;
};

export const createServer = async (server: Partial<MCPServer>): Promise<MCPServer> => {
  const { data } = await apiClient.post<MCPServer>('/api/v1/servers/', server);
  return data;
};

export const updateServer = async (
  serverId: string,
  updates: Partial<MCPServer>
): Promise<MCPServer> => {
  const { data } = await apiClient.patch<MCPServer>(`/api/v1/servers/${serverId}`, updates);
  return data;
};

export const deleteServer = async (serverId: string): Promise<void> => {
  await apiClient.delete(`/api/v1/servers/${serverId}`);
};

// Tools
export const listTools = async (serverId: string): Promise<MCPTool[]> => {
  const { data } = await apiClient.get<{ data: MCPTool[] }>(
    `/api/v1/servers/${serverId}/tools/`
  );
  return data.data;
};

export const getTool = async (serverId: string, toolId: string): Promise<MCPTool> => {
  const { data } = await apiClient.get<MCPTool>(
    `/api/v1/servers/${serverId}/tools/${toolId}`
  );
  return data;
};

export const createTool = async (
  serverId: string,
  tool: Partial<MCPTool>
): Promise<MCPTool> => {
  const { data } = await apiClient.post<MCPTool>(
    `/api/v1/servers/${serverId}/tools/`,
    tool
  );
  return data;
};

export const updateTool = async (
  serverId: string,
  toolId: string,
  updates: Partial<MCPTool>
): Promise<MCPTool> => {
  const { data } = await apiClient.patch<MCPTool>(
    `/api/v1/servers/${serverId}/tools/${toolId}`,
    updates
  );
  return data;
};

export const deleteTool = async (serverId: string, toolId: string): Promise<void> => {
  await apiClient.delete(`/api/v1/servers/${serverId}/tools/${toolId}`);
};

// Invocations
export const invokeTool = async (
  serverId: string,
  toolId: string,
  args: Record<string, any>
): Promise<InvocationResponse> => {
  const { data } = await apiClient.post<InvocationResponse>(
    `/api/v1/servers/${serverId}/tools/${toolId}/invoke`,
    { arguments: args }
  );
  return data;
};

export const listInvocations = async (params?: {
  serverId?: string;
  toolId?: string;
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<{ count: number; data: ToolInvocation[] }> => {
  const { data } = await apiClient.get('/api/v1/invocations', { params });
  return data;
};

export const getInvocation = async (invocationId: string): Promise<ToolInvocation> => {
  const { data } = await apiClient.get<ToolInvocation>(
    `/api/v1/invocations/${invocationId}`
  );
  return data;
};