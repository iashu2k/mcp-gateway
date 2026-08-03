export interface User {
  id: string;
  email: string;
  displayName: string;
  role: 'admin' | 'developer' | 'viewer';
}

export interface LoginResponse {
  accessToken: string;
  tokenType: string;
  expiresIn: number;
  user: User;
}

export interface MCPServer {
  id: string;
  name: string;
  description: string;
  baseUrl: string;
  transportType: string;
  status: 'active' | 'inactive' | 'unhealthy';
  ownerTeam: string;
  createdAt: string;
  updatedAt: string;
}

export interface MCPTool {
  id: string;
  serverId: string;
  name: string;
  title: string;
  description: string;
  inputSchema: Record<string, any>;
  riskLevel: 'low' | 'medium' | 'high';
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface ToolInvocation {
  id: string;
  serverId: string;
  toolId: string;
  userId: string;
  status: 'running' | 'succeeded' | 'failed' | 'denied';
  requestArguments: Record<string, any>;
  responsePayload?: Record<string, any>;
  errorCode?: string;
  errorMessage?: string;
  durationMs?: number;
  createdAt: string;
  completedAt?: string;
}

export interface InvocationResponse {
  invocationId: string;
  serverId: string;
  toolId: string;
  toolName: string;
  status: string;
  result?: Record<string, any>;
  durationMs: number;
  completedAt: string;
}