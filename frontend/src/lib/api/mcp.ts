import { GetStatus, Start, Stop } from '../../../wailsjs/go/http/McpAPI';
import type { ApiResponse, McpServerStatusDTO, StartMcpServerRequest } from './types';

export class McpPortConflictError extends Error {
  code: number;
  data: McpServerStatusDTO | null;

  constructor(message: string, data: McpServerStatusDTO | null) {
    super(message);
    this.name = 'McpPortConflictError';
    this.code = 409;
    this.data = data;
  }
}

function checkCode<T>(resp: ApiResponse<T>): T {
  if (resp.code !== 200) {
    if (resp.code === 409) {
      throw new McpPortConflictError(resp.message || '端口冲突', (resp.data as unknown as McpServerStatusDTO) || null);
    }
    throw new Error(resp.message || `Error code: ${resp.code}`);
  }
  return resp.data;
}

export const mcpApi = {
  getStatus: (connID: number) =>
    GetStatus(connID).then(r => checkCode(r as unknown as ApiResponse<McpServerStatusDTO>)),

  start: (req: StartMcpServerRequest) =>
    Start(req as any).then(r => checkCode(r as unknown as ApiResponse<McpServerStatusDTO>)),

  stop: (connID: number) =>
    Stop(connID).then(r => checkCode(r as unknown as ApiResponse<boolean>)),
};
