/**
 * 计算文件的 SHA-256 哈希值
 * @param file 文件对象或 Blob 对象
 * @returns 文件的 SHA-256 哈希值
 */
export const calculateHash = async (file: File | Blob): Promise<string> => {
    const arrayBuffer = await file.arrayBuffer()
    const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer)
    const hashArray = Array.from(new Uint8Array(hashBuffer))
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
    return hashHex
}

/**
 * 创建一个唯一的 ID
 * @returns 唯一的 ID 字符串
 */
export const createId = () => {
    if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID()
    return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}