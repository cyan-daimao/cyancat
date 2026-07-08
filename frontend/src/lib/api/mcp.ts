import { GetStatus, Start, Stop } from '../../../wailsjs/go/http/McpAPI';
import type { ApiResponse, McpServerStatusDTO, StartMcpServerRequest } from './types';

function checkCode<T>(resp: ApiResponse<T>): T {
  if (resp.code !== 200) {
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
