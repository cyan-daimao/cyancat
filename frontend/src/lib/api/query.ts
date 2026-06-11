import { Execute, Cancel, History } from '../../../wailsjs/go/http/QueryAPI';
import type { ApiResponse, ExecuteQueryRequest, QueryResultDTO, QueryHistoryRequest, PageResult, QueryHistoryDTO } from './types';

function checkCode<T>(resp: ApiResponse<T>): T {
  if (resp.code !== 200) {
    throw new Error(resp.message || `Error code: ${resp.code}`);
  }
  return resp.data;
}

export const queryApi = {
  execute: (req: ExecuteQueryRequest) =>
    Execute(req as any).then(r => checkCode(r as unknown as ApiResponse<QueryResultDTO>)),

  cancel: (connID: number) =>
    Cancel(connID).then(r => checkCode(r as unknown as ApiResponse<boolean>)),

  history: (req?: QueryHistoryRequest) =>
    History((req ?? {}) as any).then(r => checkCode(r as unknown as ApiResponse<PageResult<QueryHistoryDTO>>)),
};
