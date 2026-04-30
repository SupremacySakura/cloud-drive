<script setup lang="ts">
import { computed, ref } from 'vue'
import { downloadByPickupCode } from '../services/apis/file'
import { formatBytes, iconForListItem } from '../utils/file'
import { Icon } from '@iconify/vue'

const codeDigits = ref<string[]>(Array.from({ length: 6 }, () => ''))
const inputRefs = ref<(HTMLInputElement | null)[]>(Array.from({ length: 6 }, () => null))
const isLoading = ref(false)
const errorMessage = ref('')
const successState = ref(false)

const downloadedFile = ref({
  name: '',
  size: '',
  typeIcon: '',
  iconBg: '',
  iconFg: '',
  extLabel: '',
})

const getIconForContentType = (filename: string, contentType: string) => {
  if (contentType === 'application/zip' || filename.endsWith('.zip')) {
    return { ...iconForListItem({ type: 'folder', file_type: '' }), ext: 'ZIP' }
  }

  // 简易映射成 FileType 给工具函数使用
  let fileType = 'other'
  if (contentType.startsWith('video/')) fileType = 'video'
  else if (contentType.startsWith('image/')) fileType = 'image'
  else if (contentType.startsWith('audio/')) fileType = 'audio'
  else if (
    contentType.includes('pdf') ||
    contentType.includes('word') ||
    contentType.includes('excel') ||
    contentType.includes('powerpoint') ||
    contentType.startsWith('text/')
  )
    fileType = 'document'

  const iconMeta = iconForListItem({ type: 'file', file_type: fileType })
  const ext = filename.split('.').pop()?.toUpperCase() || 'FILE'

  return { ...iconMeta, ext: ext.substring(0, 4) }
}

const normalizedCode = computed(() => codeDigits.value.join('').toUpperCase())
const canSubmit = computed(() => normalizedCode.value.length === 6 && !isLoading.value)

const setInputRef = (el: HTMLInputElement | null, index: number) => {
  inputRefs.value[index] = el
}

const focusInput = (index: number) => {
  const target = inputRefs.value[index]
  if (target) {
    target.focus()
    target.select()
  }
}

const updateDigit = (index: number, value: string) => {
  const next = value
    .replace(/[^a-zA-Z0-9]/g, '')
    .slice(-1)
    .toUpperCase()
  codeDigits.value[index] = next
}

const handleInput = (index: number, event: Event) => {
  const target = event.target as HTMLInputElement
  updateDigit(index, target.value)
  target.value = codeDigits.value[index]
  if (codeDigits.value[index] && index < 5) {
    focusInput(index + 1)
  }
}

const handleKeydown = (index: number, event: KeyboardEvent) => {
  if (event.key === 'Backspace' && !codeDigits.value[index] && index > 0) {
    focusInput(index - 1)
  }
  if (event.key === 'ArrowLeft' && index > 0) {
    event.preventDefault()
    focusInput(index - 1)
  }
  if (event.key === 'ArrowRight' && index < 5) {
    event.preventDefault()
    focusInput(index + 1)
  }
}

const handlePaste = (event: ClipboardEvent) => {
  event.preventDefault()
  const text = event.clipboardData?.getData('text') ?? ''
  const chars = text
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, '')
    .slice(0, 6)
    .split('')
  if (chars.length === 0) return
  codeDigits.value = Array.from({ length: 6 }, (_, idx) => chars[idx] ?? '')
  focusInput(Math.min(chars.length, 6) - 1)
}

const resetFeedback = () => {
  errorMessage.value = ''
  successState.value = false
}

const mapErrorMessage = (message: string) => {
  if (message.includes('参数')) return '取件码格式不正确，请输入 6 位字母或数字。'
  if (message.includes('不存在')) return '取件码不存在或对应资源已删除。'
  if (message.includes('服务器')) return '服务器处理失败，请稍后重试。'
  if (message.includes('下载')) return message
  return '取件失败，请检查取件码后重试。'
}

const handleRetrieve = async () => {
  if (!canSubmit.value) return
  resetFeedback()
  isLoading.value = true
  try {
    const { fileName, fileSize, contentType } = await downloadByPickupCode(normalizedCode.value)
    const iconInfo = getIconForContentType(fileName, contentType)

    downloadedFile.value = {
      name: fileName,
      size: formatBytes(fileSize),
      typeIcon: iconInfo.icon,
      iconBg: iconInfo.bg,
      iconFg: iconInfo.fg,
      extLabel: iconInfo.ext,
    }

    successState.value = true
  } catch (error) {
    const message = error instanceof Error ? error.message : '取件失败，请稍后重试。'
    errorMessage.value = mapErrorMessage(message)
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div
    class="bg-[#f6f8f7] dark:bg-[#10221b] text-slate-900 dark:text-slate-100 min-h-screen flex flex-col font-sans relative z-0"
  >
    <main class="flex-1 flex flex-col items-center justify-center p-6 sm:p-12 relative z-10">
      <!-- Main Container -->
      <div class="w-full max-w-[480px] space-y-8">
        <!-- Entry Card -->
        <div
          class="bg-white dark:bg-slate-900/50 p-8 rounded-xl border border-primary/10 shadow-xl shadow-primary/5"
        >
          <div class="text-center mb-8">
            <h1 class="text-3xl font-black text-slate-900 dark:text-slate-100 mb-2">文件取件</h1>
            <p class="text-slate-500 dark:text-slate-400">请输入您的 6 位提取码获取文件</p>
          </div>
          <div class="space-y-6">
            <div class="flex justify-between gap-2 sm:gap-4">
              <input
                v-for="(_, index) in codeDigits"
                :key="index"
                :ref="el => setInputRef(el as HTMLInputElement | null, index)"
                :aria-label="`取件码第${index + 1}位`"
                class="w-full aspect-square text-center text-2xl font-bold rounded-lg border-2 border-slate-200 dark:border-slate-700 bg-transparent focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-all uppercase"
                maxlength="1"
                inputmode="text"
                placeholder="·"
                type="text"
                :autofocus="index === 0"
                :value="codeDigits[index]"
                @input="handleInput(index, $event)"
                @keydown="handleKeydown(index, $event)"
                @paste="handlePaste"
              />
            </div>
            <button
              aria-label="提取文件"
              class="w-full bg-primary hover:bg-primary/90 disabled:opacity-60 disabled:cursor-not-allowed text-white font-bold h-14 rounded-lg flex items-center justify-center gap-2 transition-all shadow-lg shadow-primary/20 focus:ring-2 focus:ring-primary/50 focus:outline-none"
              :disabled="!canSubmit"
              @click="handleRetrieve"
            >
              <span class="material-symbols-outlined">key</span>
              {{ isLoading ? '提取中...' : '提取文件' }}
            </button>
          </div>
          <p v-if="errorMessage" class="mt-4 text-center text-sm text-red-500">
            {{ errorMessage }}
          </p>
          <div
            class="mt-6 flex items-center justify-center gap-2 text-xs text-slate-400 dark:text-slate-500 uppercase tracking-widest font-semibold"
          >
            <span class="material-symbols-outlined text-sm">lock_clock</span>
            仅限有效期内提取
          </div>
        </div>

        <!-- Success State Section -->
        <div v-if="successState" class="space-y-4">
          <div class="flex items-center gap-2 px-2">
            <span class="material-symbols-outlined text-primary text-sm">check_circle</span>
            <h3 class="text-sm font-bold uppercase tracking-wider text-slate-500">已提取文件</h3>
          </div>
          <div
            class="bg-white dark:bg-slate-900/50 p-6 rounded-xl border border-primary/20 shadow-lg flex flex-col sm:flex-row items-center gap-6"
          >
            <div
              class="h-20 w-16 rounded-lg border flex items-center justify-center flex-shrink-0 relative overflow-hidden group"
              :class="[downloadedFile.iconBg, downloadedFile.iconFg, 'border-current/20']"
            >
              <div
                class="absolute inset-0 bg-current opacity-5 group-hover:opacity-10 transition-opacity"
              ></div>
              <Icon :icon="downloadedFile.typeIcon" class="text-4xl" />
              <div
                class="absolute bottom-1 right-1 bg-current text-[8px] text-white px-1 rounded font-bold"
                :class="downloadedFile.iconFg.replace('text-', 'bg-')"
              >
                {{ downloadedFile.extLabel }}
              </div>
            </div>
            <div class="flex-1 text-center sm:text-left overflow-hidden">
              <h4
                class="text-slate-900 dark:text-slate-100 font-bold truncate text-lg"
                :title="downloadedFile.name"
              >
                {{ downloadedFile.name }}
              </h4>
              <p class="text-slate-500 dark:text-slate-400 text-sm">
                {{ downloadedFile.size }} • 下载成功
              </p>
            </div>
            <button
              aria-label="重新下载文件"
              class="w-full sm:w-auto px-6 h-12 bg-slate-900 dark:bg-white dark:text-slate-900 text-white rounded-lg font-bold flex items-center justify-center gap-2 hover:scale-[1.02] transition-transform focus:ring-2 focus:ring-slate-400 focus:outline-none"
              @click="handleRetrieve"
            >
              <span class="material-symbols-outlined text-xl">download</span>
              重新下载
            </button>
          </div>
        </div>

        <!-- Security Footer -->
        <div class="text-center py-4 border-t border-primary/5">
          <p class="text-xs text-slate-400 dark:text-slate-500 leading-relaxed">
            传输过程全程加密，提取码过期后将无法下载文件。
            <br />
          </p>
        </div>
      </div>
    </main>

    <!-- Background Decoration -->
    <div class="fixed inset-0 pointer-events-none z-[-1] overflow-hidden">
      <div
        class="absolute top-[-10%] right-[-10%] w-[40%] h-[40%] bg-primary/5 rounded-full blur-[120px]"
      ></div>
      <div
        class="absolute bottom-[-10%] left-[-10%] w-[40%] h-[40%] bg-primary/5 rounded-full blur-[120px]"
      ></div>
    </div>
  </div>
</template>
