export enum FileType {
  Image = 'image',
  Video = 'video',
  Audio = 'audio',
  Document = 'document',
  Other = 'other',
}

export type UploadFileConfig = {
  file_type: FileType
  folder_id?: number
}
