<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { computed, reactive, ref } from 'vue'
import { uploadFile } from '../services/apis/file'
import { detectFileType, formatBytes, iconForFile } from '../utils/file'
import { createId } from '../utils/hash'
import { useUserStore } from '../stores/user'
import LoginRequiredPlaceholder from '../components/bussiness/LoginRequiredPlaceholder.vue'

const userStore = useUserStore()

type QueueItemStatus = 'pending' | 'hashing' | 'uploading' | 'merging' | 'success' | 'failed' | 'canceled'

type QueueItem = {
    id: string
    file: File
    status: QueueItemStatus
    percent: number
    message: string | null
    canceled: boolean
}

const isDragging = ref(false)
const fileInputRef = ref<HTMLInputElement | null>(null)
const items = ref<QueueItem[]>([])
const processingCount = computed(() => items.value.filter((i) => ['pending', 'hashing', 'uploading', 'merging'].includes(i.status)).length)
const clampPercent = (value: number) => Math.min(100, Math.max(0, Math.floor(value)))

const removeItem = (id: string) => {
    items.value = items.value.filter((x) => x.id !== id)
}

const cancelItem = (item: QueueItem) => {
    item.canceled = true
    item.status = 'canceled'
    item.message = '已取消'
}

const retryItem = async (item: QueueItem) => {
    item.status = 'pending'
    item.percent = 0
    item.message = null
    item.canceled = false
    await startUpload(item)
}

const clearCompleted = () => {
    items.value = items.value.filter((x) => x.status !== 'success')
}

const openFileDialog = () => {
    fileInputRef.value?.click()
}

const pushFiles = async (files: File[]) => {
    for (const file of files) {
        const item = reactive<QueueItem>({
            id: createId(),
            file,
            status: 'pending',
            percent: 0,
            message: null,
            canceled: false,
        })
        items.value.unshift(item)
        await startUpload(item)
    }
}

const onFileInputChange = async (e: Event) => {
    const el = e.target as HTMLInputElement
    const fileList = el.files
    if (!fileList?.length) return
    const files = Array.from(fileList)
    el.value = ''
    await pushFiles(files)
}

const onDragOver = (e: DragEvent) => {
    e.preventDefault()
    isDragging.value = true
}

const onDragLeave = () => {
    isDragging.value = false
}

const onDrop = async (e: DragEvent) => {
    e.preventDefault()
    isDragging.value = false
    const fileList = e.dataTransfer?.files
    if (!fileList?.length) return
    await pushFiles(Array.from(fileList))
}

const startUpload = async (item: QueueItem) => {
    try {
        item.status = 'uploading'
        item.percent = 0
        item.message = '上传中...'
        const fileType = detectFileType(item.file)
        await uploadFile(
            item.file,
            {
                file_type: fileType,
                folder_id: 0,
            },
            (progress) => {
                if (item.canceled) return
                item.percent = clampPercent(progress)
                item.message = progress >= 100 ? '上传完成' : '上传中...'
            },
        )

        if (item.canceled) return
        item.status = 'success'
        item.percent = 100
        item.message = '上传完成'
    } catch (e: any) {
        if (item.canceled) return
        item.status = 'failed'
        item.message = e?.message || '上传失败'
    }
}

const badgeText = (item: QueueItem) => {
    if (item.status === 'success') return 'Success'
    if (item.status === 'failed') return 'Failed'
    if (item.status === 'canceled') return 'Canceled'
    if (item.status === 'merging') return 'Merging'
    if (item.status === 'hashing') return 'Hashing'
    if (item.status === 'pending') return 'Pending'
    return `${clampPercent(item.percent)}%`
}

const badgeClass = (item: QueueItem) => {
    if (item.status === 'failed') return 'text-red-500'
    if (item.status === 'success') return 'text-primary'
    if (item.status === 'canceled') return 'text-slate-400'
    return 'text-primary'
}
</script>

<template>
    <LoginRequiredPlaceholder v-if="!userStore.isLoggedIn" />
    <div v-else
        class="flex-1 flex flex-col min-w-0 bg-background-light dark:bg-background-dark font-display text-slate-900 dark:text-slate-100">
        <main class="flex-1 overflow-y-auto p-8 space-y-8">
            <section>
                <input ref="fileInputRef" type="file" class="hidden" multiple @change="onFileInputChange" />
                <div class="w-full border-2 border-dashed border-primary/30 rounded-xl bg-primary/5 hover:bg-primary/[0.08] transition-all group flex flex-col items-center justify-center py-16 px-4 cursor-pointer"
                    :class="isDragging ? 'ring-2 ring-primary/40 bg-primary/[0.10]' : ''" @click="openFileDialog"
                    @dragover="onDragOver" @dragleave="onDragLeave" @drop="onDrop">
                    <div
                        class="size-16 rounded-full bg-white dark:bg-background-dark shadow-sm flex items-center justify-center text-primary mb-4 group-hover:scale-110 transition-transform">
                        <Icon class="text-4xl" icon="material-symbols:upload-file" />
                    </div>
                    <h3 class="text-xl font-bold text-slate-800 dark:text-slate-100 mb-2">拖拽文件到此处上传</h3>
                    <p class="text-slate-500 dark:text-slate-400 text-sm mb-6 max-w-sm text-center">
                        支持图片、视频、文档与压缩包等常见格式。
                    </p>
                    <button
                        class="bg-white dark:bg-slate-800 border border-primary/30 text-primary font-bold px-8 py-2.5 rounded-lg hover:bg-primary hover:text-white transition-all shadow-sm"
                        type="button">
                        选择文件
                    </button>
                </div>
            </section>

            <section class="space-y-4">
                <div class="flex items-center justify-between">
                    <h3 class="text-lg font-bold text-slate-800 dark:text-slate-100">上传队列</h3>
                    <div class="flex items-center gap-4">
                        <span class="text-sm text-slate-500">{{ processingCount }} 项处理中</span>
                        <button class="text-sm font-bold text-primary hover:underline" type="button"
                            @click="clearCompleted">
                            清理已完成
                        </button>
                    </div>
                </div>

                <div class="bg-white dark:bg-slate-900 border border-primary/10 rounded-xl overflow-hidden shadow-sm">
                    <div v-if="items.length === 0" class="p-8 text-center text-slate-500 dark:text-slate-400">
                        暂无上传任务
                    </div>

                    <div v-for="item in items" :key="item.id"
                        class="flex items-center p-4 border-b border-primary/5 last:border-0 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
                        <div class="size-12 rounded-lg flex items-center justify-center shrink-0"
                            :class="`${iconForFile(item.file).bg} ${iconForFile(item.file).fg}`">
                            <Icon :icon="iconForFile(item.file).icon" />
                        </div>

                        <div class="ml-4 flex-1 min-w-0">
                            <div class="flex items-center justify-between mb-1">
                                <p class="text-sm font-bold text-slate-800 dark:text-slate-100 truncate">
                                    {{ item.file.name }}
                                </p>
                                <span class="text-xs font-bold uppercase tracking-wider" :class="badgeClass(item)">
                                    {{ badgeText(item) }}
                                </span>
                            </div>

                            <div class="flex items-center gap-3">
                                <div class="flex-1 bg-primary/10 h-1.5 rounded-full overflow-hidden">
                                    <div class="h-1.5 rounded-full"
                                        :class="item.status === 'failed' ? 'bg-red-500' : 'bg-primary'"
                                        :style="{ width: `${clampPercent(item.percent)}%` }"></div>
                                </div>
                                <span class="text-xs text-slate-500 whitespace-nowrap">
                                    {{ formatBytes(item.file.size) }} • {{ item.message || '等待中' }}
                                </span>
                            </div>
                        </div>

                        <div class="ml-6 flex items-center gap-2">
                            <button v-if="item.status === 'failed'"
                                class="flex items-center gap-1.5 bg-primary/10 hover:bg-primary/20 text-primary font-bold px-3 py-1.5 rounded-lg text-xs transition-colors"
                                type="button" @click="retryItem(item)">
                                <Icon class="text-sm" icon="material-symbols:replay" />
                                重试
                            </button>

                            <div v-if="item.status === 'success'" class="p-2 text-primary">
                                <Icon class="text-xl" icon="material-symbols:check-circle" />
                            </div>

                            <button
                                v-if="item.status === 'uploading' || item.status === 'hashing' || item.status === 'merging'"
                                class="p-2 text-slate-400 hover:text-red-500 rounded-lg transition-colors" type="button"
                                @click="cancelItem(item)">
                                <Icon class="text-xl" icon="material-symbols:close" />
                            </button>

                            <button v-else class="p-2 text-slate-400 hover:text-red-500 rounded-lg transition-colors"
                                type="button" @click="removeItem(item.id)">
                                <Icon class="text-xl" icon="material-symbols:close" />
                            </button>
                        </div>
                    </div>
                </div>
            </section>
        </main>
    </div>
</template>

<style lang="sass" scoped></style>
