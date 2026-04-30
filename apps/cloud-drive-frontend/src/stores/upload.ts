import { defineStore } from 'pinia'

type UploadTaskStatus =
  | 'pending'
  | 'hashing'
  | 'uploading'
  | 'merging'
  | 'success'
  | 'failed'
  | 'canceled'
export type UploadTask = {
  id: string
  file: File
  targetFolderId: number
  relativePath: string
  status: UploadTaskStatus
  percent: number
  message: string | null
  canceled: boolean
}

export const useUploadStore = defineStore('upload', {
  state: () => ({
    uploadTasks: [] as UploadTask[],
    isUploadPanelOpen: false,
    isUploading: false,
    fileInputRef: null as HTMLInputElement | null,
    folderInputRef: null as HTMLInputElement | null,
  }),
  actions: {
    addTasks(tasks: UploadTask[]) {
      this.uploadTasks.unshift(...tasks)
    },
    setPanelOpen(open: boolean) {
      this.isUploadPanelOpen = open
    },
    setUploading(flag: boolean) {
      this.isUploading = flag
    },
    setFileInputRef(ref: HTMLInputElement | null) {
      this.fileInputRef = ref
    },
    setFolderInputRef(ref: HTMLInputElement | null) {
      this.folderInputRef = ref
    },
  },
})
