import { SelectSQLiteDatabaseFile } from '../../../wailsjs/go/http/FileDialogAPI';
import type { ApiResponse } from './types';

function checkCode<T>(resp: ApiResponse<T>): T {
  if (resp.code !== 200) {
    throw new Error(resp.message || `Error code: ${resp.code}`);
  }
  return resp.data;
}

export const fileDialogApi = {
  selectSQLiteDatabaseFile: () =>
    SelectSQLiteDatabaseFile().then(r => checkCode(r as unknown as ApiResponse<string>)),
};
