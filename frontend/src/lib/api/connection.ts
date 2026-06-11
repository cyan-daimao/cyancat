import { List, Page, GetByID, Create, Update, Delete, Test, Open, Close } from '../../../wailsjs/go/http/ConnectionAPI';
import type {
  ApiResponse, ConnectionDTO, CreateConnectionRequest, UpdateConnectionRequest,
  TestConnectionRequest, TestConnectionResultDTO, ListConnectionRequest,
  PageConnectionRequest, PageResult
} from './types';

function checkCode<T>(resp: ApiResponse<T>): T {
  if (resp.code !== 200) {
    throw new Error(resp.message || `Error code: ${resp.code}`);
  }
  return resp.data;
}

export const connectionApi = {
  list: (req?: ListConnectionRequest) =>
    List((req ?? {}) as any).then(r => checkCode(r as unknown as ApiResponse<ConnectionDTO[]>)),

  page: (req?: PageConnectionRequest) =>
    Page((req ?? {}) as any).then(r => checkCode(r as unknown as ApiResponse<PageResult<ConnectionDTO>>)),

  getById: (id: number) =>
    GetByID(id).then(r => checkCode(r as unknown as ApiResponse<ConnectionDTO>)),

  create: (req: CreateConnectionRequest) =>
    Create(req as any).then(r => checkCode(r as unknown as ApiResponse<ConnectionDTO>)),

  update: (id: number, req: UpdateConnectionRequest) =>
    Update(id, req as any).then(r => checkCode(r as unknown as ApiResponse<ConnectionDTO>)),

  delete: (id: number) =>
    Delete(id).then(r => checkCode(r as unknown as ApiResponse<boolean>)),

  test: (req: TestConnectionRequest) =>
    Test(req as any).then(r => checkCode(r as unknown as ApiResponse<TestConnectionResultDTO>)),

  open: (id: number) =>
    Open(id).then(r => checkCode(r as unknown as ApiResponse<ConnectionDTO>)),

  close: (id: number) =>
    Close(id).then(r => checkCode(r as unknown as ApiResponse<boolean>)),
};
