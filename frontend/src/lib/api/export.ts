import { ExportCSV } from '../../../wailsjs/go/http/ExportAPI';
import type { ApiResponse, ExportCSVRequest, ExportCSVResult } from './types';

function checkCode<T>(resp: ApiResponse<T>): T {
  if (resp.code !== 200) {
    throw new Error(resp.message || `Error code: ${resp.code}`);
  }
  return resp.data;
}

export const exportApi = {
  exportCSV: (req: ExportCSVRequest) =>
    ExportCSV(req as any).then(r => checkCode(r as unknown as ApiResponse<ExportCSVResult>)),
};
