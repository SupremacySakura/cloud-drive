import request from '../request'
import type { ResponseData } from '../types'
import { calculateHash } from '../../utils/hash'
import type { FileListItem, InitUploadFileRequest, InitUploadFileResponse, MakeDirectoryRequest } from '../types/file'
import type { UploadFileConfig } from '../../types/file'

export const uploadFile = async (file: File, fileConfig: UploadFileConfig, onProgress: (progress: number) => void) => {
    onProgress(0)
    const total_chunks = Math.ceil(file.size / (1024 * 1024))
    // init
    const initData: InitUploadFileRequest = {
        file_name: file.name,
        file_size: file.size,
        file_hash: await calculateHash(file),
        chunk_size: 1024 * 1024,
        total_chunks,
        file_type: fileConfig.file_type,
        folder_id: fileConfig.folder_id || 0,
    }
    const initRes = await request.post<ResponseData<InitUploadFileResponse>>('/api/file/init', initData)
    if (initRes.data.code !== 0) {
        onProgress(0)
        return
    }
    if (initRes.data.data.status !== 'uploading') {
        onProgress(100)
        return
    }
    const uploaded_chunks = new Set(initRes.data.data.uploaded_chunks)
    // 上传的时候跳过已经上传的chunk
    for (let chunk_index = 0; chunk_index < total_chunks; chunk_index++) {
        if (uploaded_chunks.has(chunk_index)) {
            onProgress((chunk_index + 1) / total_chunks * 100)
            continue
        }
        // 上传chunk
        const chunk = file.slice(chunk_index * 1024 * 1024, (chunk_index + 1) * 1024 * 1024)
        const uploadData = new FormData()
        uploadData.append('task_id', initRes.data.data.task_id.toString())
        uploadData.append('chunk_index', chunk_index.toString())
        uploadData.append('chunk_data', chunk)
        uploadData.append('chunk_hash', await calculateHash(chunk))
        const chunkRes = await request.post<ResponseData<{}>>('/api/file/chunk', uploadData)
        if (chunkRes.data.code !== 0) {
            onProgress(0)
            return
        }
        // 更新进度
        onProgress((chunk_index + 1) / total_chunks * 100)
    }
    // 合并chunk
    const mergeRes = await request.post<ResponseData<{}>>('/api/file/merge', {
        task_id: initRes.data.data.task_id,
    })
    if (mergeRes.data.code !== 0) {
        onProgress(0)
        return
    }
    onProgress(100)
}

export const getListByFolderIDAndUserID = async (folderID: number, page: number, pageSize: number) => {
    const res = await request.get<ResponseData<FileListItem[]>>('/api/file/list', {
        params: {
            folder_id: folderID,
            page,
            page_size: pageSize,
        },
    })
    if (res.data.code !== 0) {
        return []
    }
    return res.data.data
}

export const getListCountByFolderIDAndUserID = async (folderID: number) => {
    const res = await request.get<ResponseData<number>>('/api/file/list/count', {
        params: {
            folder_id: folderID,
        },
    })
    if (res.data.code !== 0) {
        return 0
    }
    return res.data.data
}

export const makeDirectory = async (data: MakeDirectoryRequest) => {
    const res = await request.post<ResponseData<number>>('/api/file/mkdir', data)
    if (res.data.code !== 0) {
        return 0
    }
    return res.data.data
}
