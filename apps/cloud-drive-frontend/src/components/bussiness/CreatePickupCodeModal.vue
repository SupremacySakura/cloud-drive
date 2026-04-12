<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { createPickupCode, getListByFolderIDAndUserID } from '../../services/apis/file'
import type { PickupCodeType } from '../../services/types/file'
import type { FileListItem } from '../../services/types/file'

const emit = defineEmits<{
    (e: 'close'): void
    (e: 'success'): void
}>()

const form = ref({
    file_id: null as number | null,
    folder_id: null as number | null,
    type: 'file' as PickupCodeType,
    max_downloads: 10,
    expire_days: 7,
})

const fileList = ref<FileListItem[]>([])
const loading = ref(false)
const submitting = ref(false)
const selectedItemId = ref<number | null>(null)
const currentFolderId = ref(0)
const pathStack = ref<{ id: number; name: string }[]>([{ id: 0, name: '根目录' }])
const toast = ref<{ type: 'success' | 'error'; text: string } | null>(null)

const minDownloads = 1
const maxDownloads = 100
const expireDayOptions = [1, 7, 14, 30]

const expireTime = computed(() => {
    const date = new Date()
    date.setDate(date.getDate() + form.value.expire_days)
    return date.toISOString()
})

const canSubmit = computed(() => {
    return selectedItemId.value !== null && form.value.max_downloads >= minDownloads && form.value.max_downloads <= maxDownloads
})

const displayFileList = computed(() => {
    if (currentFolderId.value === 0) {
        return fileList.value
    }
    const parentItem: FileListItem = {
        id: -1,
        name: '..',
        type: 'folder',
        file_type: '',
        size: 0,
        updated_at: '',
    }
    return [parentItem, ...fileList.value]
})

const clearSelection = () => {
    selectedItemId.value = null
    form.value.file_id = null
    form.value.folder_id = null
}

const fetchFileList = async (folderId = 0) => {
    loading.value = true
    try {
        currentFolderId.value = folderId
        fileList.value = await getListByFolderIDAndUserID(folderId, 1, 100)
        clearSelection()
    } finally {
        loading.value = false
    }
}

const handleSelectItem = (item: FileListItem) => {
    if (item.id === -1) return
    selectedItemId.value = item.id
    if (item.type === 'folder') {
        form.value.folder_id = item.id
        form.value.file_id = null
        form.value.type = 'folder'
    } else {
        form.value.file_id = item.id
        form.value.folder_id = null
        form.value.type = 'file'
    }
}

const handleGoParent = async () => {
    if (pathStack.value.length <= 1) return
    pathStack.value = pathStack.value.slice(0, pathStack.value.length - 1)
    const parent = pathStack.value[pathStack.value.length - 1]
    await fetchFileList(parent.id)
}

const handleEnterFolder = async (item: FileListItem) => {
    if (item.id === -1) {
        await handleGoParent()
        return
    }
    if (item.type !== 'folder') return
    pathStack.value.push({ id: item.id, name: item.name })
    await fetchFileList(item.id)
}

const handleNavigatePath = async (index: number) => {
    const target = pathStack.value[index]
    if (!target) return
    pathStack.value = pathStack.value.slice(0, index + 1)
    await fetchFileList(target.id)
}

const handleSubmit = async () => {
    if (!canSubmit.value || submitting.value) return

    submitting.value = true
    try {
        const result = await createPickupCode({
            code: '',
            file_id: form.value.file_id,
            folder_id: form.value.folder_id,
            type: form.value.type,
            max_downloads: form.value.max_downloads,
            expire_time: expireTime.value,
        })

        if (result !== null) {
            toast.value = { type: 'success', text: '创建取件码成功' }
            setTimeout(() => {
                emit('success')
            }, 700)
            return
        }
        toast.value = { type: 'error', text: '创建取件码失败，请重试' }
    } catch {
        toast.value = { type: 'error', text: '创建取件码失败，请重试' }
    } finally {
        submitting.value = false
    }
}

onMounted(() => {
    fetchFileList(0)
})
</script>

<template>
    <div class="bg-white dark:bg-slate-900 rounded-xl shadow-2xl w-full max-w-xl mx-4 overflow-hidden">
        <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
            <h3 class="text-lg font-bold text-slate-900 dark:text-white">创建取件码</h3>
            <button class="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors"
                type="button" @click="emit('close')">
                <Icon icon="material-symbols:close-rounded" class="text-xl" />
            </button>
        </div>

        <div class="p-6 space-y-6">
            <div v-if="toast" class="px-4 py-3 rounded-lg text-sm font-medium border" :class="toast.type === 'success'
                ? 'bg-green-50 text-green-700 border-green-200 dark:bg-green-900/20 dark:text-green-300 dark:border-green-800'
                : 'bg-red-50 text-red-700 border-red-200 dark:bg-red-900/20 dark:text-red-300 dark:border-red-800'">
                {{ toast.text }}
            </div>
            <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">选择文件或文件夹
                </label>
                <div class="mb-2 flex items-center gap-1 text-xs text-slate-500 overflow-x-auto whitespace-nowrap">
                    <button v-for="(path, index) in pathStack" :key="`${path.id}-${index}`" type="button"
                        class="inline-flex items-center gap-1 hover:text-primary transition-colors"
                        @click="handleNavigatePath(index)">
                        <Icon v-if="index !== 0" icon="material-symbols:chevron-right-rounded" class="text-sm" />
                        <span>{{ path.name }}</span>
                    </button>
                </div>
                <div v-if="loading" class="text-center py-8 text-slate-500">
                    正在加载文件...
                </div>
                <div v-else-if="displayFileList.length === 0" class="text-center py-8 text-slate-500">
                    暂无可选文件，请先上传文件。
                </div>
                <div v-else class="max-h-48 overflow-y-auto border border-slate-200 dark:border-slate-700 rounded-lg">
                    <div v-for="item in displayFileList" :key="item.id"
                        class="px-4 py-3 flex items-center gap-3 cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
                        :class="{ 'bg-primary/10 border-l-2 border-l-primary': selectedItemId === item.id && item.id !== -1 }"
                        @click="handleSelectItem(item)" @dblclick="handleEnterFolder(item)">
                        <Icon
                            :icon="item.id === -1 ? 'material-symbols:drive-folder-upload-rounded' : item.type === 'folder' ? 'material-symbols:folder-outline-rounded' : 'material-symbols:description-outline-rounded'"
                            class="text-xl" :class="item.type === 'folder' ? 'text-amber-500' : 'text-slate-400'" />
                        <div class="flex-1 min-w-0">
                            <p class="text-sm font-medium text-slate-900 dark:text-white truncate">{{ item.name }}</p>
                            <p class="text-xs text-slate-500">{{ item.id === -1 ? '返回上一级' : item.type ===
                                'folder' ? '文件夹' : '文件' }}</p>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Max Downloads -->
            <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">最大下载次数
                </label>
                <input type="number" v-model.number="form.max_downloads" :min="minDownloads" :max="maxDownloads"
                    class="w-full px-4 py-2 border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent" />
                <p class="mt-1 text-xs text-slate-500">范围：{{ minDownloads }} - {{ maxDownloads }}</p>
            </div>

            <!-- Expiration -->
            <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">有效期
                </label>
                <div class="flex gap-2">
                    <button v-for="days in expireDayOptions" :key="days" type="button"
                        class="flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-all"
                        :class="form.expire_days === days
                            ? 'bg-primary text-white'
                            : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700'" @click="form.expire_days = days">
                        {{ days }} 天
                    </button>
                </div>
                <p class="mt-2 text-xs text-slate-500">到期时间：{{ new Date(expireTime).toLocaleDateString() }}</p>
            </div>
        </div>

        <div class="px-6 py-4 border-t border-slate-200 dark:border-slate-800 flex justify-end gap-3">
            <button type="button"
                class="px-4 py-2 rounded-lg text-sm font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                @click="emit('close')">
                取消
            </button>
            <button type="button"
                class="px-6 py-2 rounded-lg text-sm font-medium bg-primary text-white hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                :disabled="!canSubmit || submitting" @click="handleSubmit">
                {{ submitting ? '创建中...' : '生成取件码' }}
            </button>
        </div>
    </div>
</template>
