import request from '../request'
import type { ResponseData } from '../types'
import type { OperationLogListResponse } from '../types/log'

const EMPTY_LOG_PAGE: OperationLogListResponse = { list: [], total: 0 }

/** 分页获取当前用户的操作日志（首页操作 feed 数据源） */
export const getOperationLogs = async (page: number, pageSize: number, signal?: AbortSignal) => {
  const res = await request.get<ResponseData<OperationLogListResponse>>('/api/file/logs', {
    params: {
      page,
      page_size: pageSize,
    },
    signal,
  })
  if (res.data.code !== 0) {
    return EMPTY_LOG_PAGE
  }
  return res.data.data ?? EMPTY_LOG_PAGE
}
