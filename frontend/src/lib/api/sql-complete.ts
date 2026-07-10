import { Complete } from '../../../wailsjs/go/http/SqlCompleteAPI';
import type { ApiResponse, SqlCompleteCandidate, SqlCompleteRequest } from './types';

function checkCode<T>(resp: ApiResponse<T>): T {
  if (resp.code !== 200) {
    throw new Error(resp.message || `Error code: ${resp.code}`);
  }
  return resp.data;
}

export const sqlCompleteApi = {
  complete: (req: SqlCompleteRequest) =>
    Complete(req as any).then(r => checkCode(r as unknown as ApiResponse<SqlCompleteCandidate[]>)),
};
