// 仅用于设计预览：所有任务和数据均在内存中，不调用真实服务。
const files = [
    { id: 'intro', name: 'RAG 学习笔记.md', size: '12.4 KB', state: 'new', title: 'RAG 学习笔记' },
    { id: 'react', name: 'React Server Components.md', size: '8.2 KB', state: 'ready', title: 'React Server Components' },
    { id: 'queue', name: 'Go 并发模式.markdown', size: '6.8 KB', state: 'pending', title: 'Go 并发模式' },
    { id: 'empty', name: '待整理笔记.md', size: '0 KB', state: 'failed', title: '待整理笔记', error: '未读取到有效正文。原文件为空，请更新文件后重试。' },
    { id: 'pdf', name: '系统设计参考.pdf', size: '2.1 MB', state: 'unsupported' },
    { id: 'image', name: '架构草图.png', size: '420 KB', state: 'unsupported' },
]
const labels = { new: '未加入', ready: '已加入', pending: '排队中', running: '处理中', failed: '加入失败', unsupported: '暂不支持' }
const $ = (id) => document.getElementById(id)
const escapeHtml = (value) => String(value).replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char])
const badge = (file) => `<span class="status ${file.state}">${labels[file.state]}</span>`
let chosenFile = null
let returnFocus = null
let toastTimer

function showToast(message) {
    clearTimeout(toastTimer)
    $('toast-message').textContent = message
    $('toast-wrap').hidden = false
    toastTimer = setTimeout(() => { $('toast-wrap').hidden = true }, 3500)
}

function action(file) {
    const base = `class="text-action" data-id="${file.id}"`
    if (file.state === 'new') return `<button ${base} data-action="add">加入知识库</button>`
    if (file.state === 'ready') return `<a class="text-action" href="#detail/${file.id}">查看知识</a>`
    if (file.state === 'failed') return `<button ${base} data-action="retry">重试加入</button>`
    if (file.state === 'unsupported') return '<button class="text-action" disabled>仅支持 Markdown</button>'
    return `<button ${base} data-action="task">查看进度</button>`
}

function renderFiles() {
    const query = $('file-search').value.trim().toLowerCase()
    const filtered = files.filter((file) => file.name.toLowerCase().includes(query))
    $('file-rows').innerHTML = filtered.map((file) => `<tr>
    <td><div class="filename"><span class="md-icon ${file.state === 'unsupported' ? 'other' : ''}">${file.id === 'pdf' ? 'PDF' : file.id === 'image' ? 'PNG' : 'MD'}</span><div><strong>${escapeHtml(file.name)}</strong><small>学习资料</small></div></div></td>
    <td class="t-secondary">${file.size}</td><td>${badge(file)}</td><td>${action(file)}</td>
  </tr>`).join('')
    $('file-empty').hidden = filtered.length > 0
}

function renderLibrary() {
    const query = $('knowledge-search').value.trim().toLowerCase()
    const filter = $('knowledge-filter').value
    const items = files.filter((file) => !['new', 'unsupported'].includes(file.state) && file.title.toLowerCase().includes(query) && (filter === 'all' || file.state === filter))
    $('knowledge-list').innerHTML = items.map((file) => `<div class="knowledge-row"><div>
    ${badge(file)}<h2>${escapeHtml(file.title)}</h2><p>${file.state === 'ready' ? '正文已提取并保存，可打开阅读。原始 Markdown 文件保持不变。' : file.state === 'failed' ? escapeHtml(file.error) : '正在准备知识正文，完成后即可阅读。你可以继续浏览其他内容。'}</p>
    <div class="knowledge-meta"><span>Markdown 文件</span><span>${escapeHtml(file.name)}</span><span>10月19日</span></div>
    </div><div>${action(file)}</div></div>`).join('')
    const isEmpty = items.length === 0
    $('library-empty').hidden = !isEmpty
    if (isEmpty) {
        $('library-empty').querySelector('h2').textContent = filter === 'empty' || (!query && filter === 'all') ? '让第一份资料成为知识' : '没有匹配的知识'
        $('library-empty').querySelector('p').textContent = filter === 'empty' || (!query && filter === 'all') ? '选择已经上传的 Markdown 文件，保留原件，同时获得清晰可读的知识正文。' : '试试其他关键词或切换处理状态。'
    }
}

const content = {
    intro: `<h2>理解 RAG：让回答有据可查</h2><p>检索增强生成（RAG）将资料检索与大语言模型结合。回答之前，系统先找到相关内容，再把这些内容作为回答依据。</p><h3>一个清晰的处理流程</h3><ul><li>保留原始文档，提取可读正文。</li><li>将正文切分成适合检索的片段。</li><li>根据问题检索相关片段。</li><li>基于证据回答，并标明来源。</li></ul><h3>为什么要保留来源？</h3><p>知识库不是把文件换一个地方存放。它让资料成为可查找、可核对的知识；原件则始终用于回溯。</p><blockquote>先让知识能够可靠地保存和阅读，再逐步加入智能检索。</blockquote>`,
    react: `<h2>React Server Components</h2><p>Server Components 允许部分组件在服务端执行。它们可以靠近数据源读取数据，并减少发送给浏览器的 JavaScript。</p><h3>服务端与客户端的职责</h3><p>需要状态、事件处理或浏览器 API 的交互部分，可以使用 Client Components。并不是所有组件都应该放到客户端。</p><pre>'use client'\n\nexport function Counter() {\n  // 在客户端处理交互状态\n}</pre><h3>理解边界</h3><ul><li>服务端组件负责读取数据与组织内容。</li><li>客户端组件负责交互与浏览器能力。</li><li>在边界之间传递可序列化的数据。</li></ul><blockquote>这份笔记用于理解组件执行位置，不等同于完整的框架迁移指南。</blockquote>`,
    queue: `<h2>Go 并发模式</h2><p>使用 goroutine 执行并发工作，通过 channel 协调数据流。并发任务需要明确的退出条件，避免泄漏。</p><h3>任务取消</h3><p>将 <code>context.Context</code> 传入任务，在取消时及时结束处理。</p><h3>有限并发</h3><p>使用固定数量的 worker 控制资源占用，不为每个输入无限创建 goroutine。</p>`,
}

function navigate() {
    const route = decodeURIComponent(location.hash.slice(1)) || 'files'
    const [page, id] = route.split('/')
    const activePage = ['files', 'library', 'detail', 'states'].includes(page) ? page : 'files'
    if (activePage === 'files' && $('knowledge-filter').value === 'empty') $('knowledge-filter').value = 'all'
    for (const key of ['files', 'library', 'detail', 'states']) $(`view-${key}`).hidden = key !== activePage
    document.querySelectorAll('[data-page]').forEach((link) => {
        const selected = link.dataset.page === (activePage === 'detail' ? 'library' : activePage)
        link.classList.toggle('active', selected)
        if (selected) link.setAttribute('aria-current', 'page')
        else link.removeAttribute('aria-current')
    })
    if (activePage === 'detail') {
        $('progress-panel').hidden = true
        const file = files.find((item) => item.id === id && item.state === 'ready')
        if (!file) { location.hash = 'library'; return }
        $('detail-title').textContent = file.title
        $('detail-crumb').textContent = file.title
        $('detail-file').textContent = file.name
        $('detail-content').innerHTML = content[file.id] || content.intro
    }
    renderFiles()
    renderLibrary()
    window.scrollTo(0, 0)
}

function renderPanel(file) {
    $('progress-panel').hidden = false
    $('progress-filename').textContent = file.name
    $('progress-title').textContent = labels[file.state]
    const finished = file.state === 'ready'
    const failed = file.state === 'failed'
    const step = file.step ?? (finished ? 3 : 0)
    $('progress-steps').innerHTML = ['读取原文件', '提取 Markdown 正文', '创建知识条目'].map((label, index) => `<div class="progress-step ${finished || step > index ? 'done' : !failed && file.state === 'running' && step === index ? 'current' : ''}"><span class="step-dot">${finished || step > index ? '✓' : index + 1}</span>${label}</div>`).join('')
    $('progress-message').textContent = failed ? file.error : finished ? '正文已入库。原文件保持不变，可以打开知识详情。' : file.state === 'pending' ? '已提交，等待处理。关闭此面板不会取消任务。' : '正在处理演示任务。原文件保持不变。'
    $('panel-actions').innerHTML = finished ? `<a class="btn btn-tint" href="#detail/${file.id}">查看知识</a>` : failed ? `<button class="btn btn-secondary" data-action="retry" data-id="${file.id}">重试加入</button>` : `<button class="btn btn-secondary" data-action="simulate" data-id="${file.id}">推进演示下一步</button>`
}

function advance(file) {
    if (file.state === 'ready') return
    if (file.state === 'pending') { file.state = 'running'; file.step = 0 }
    else if (file.id === 'empty') { file.state = 'failed'; file.step = 0 }
    else { file.step = (file.step || 0) + 1; if (file.step >= 3) file.state = 'ready' }
    renderFiles()
    renderLibrary()
    renderPanel(file)
    if (file.state === 'ready') showToast('已加入知识库，原文件保持不变')
}

document.addEventListener('click', (event) => {
    const trigger = event.target.closest('[data-action]')
    if (!trigger) return
    const file = files.find((item) => item.id === trigger.dataset.id)
    if (!file) return
    if (trigger.dataset.action === 'add') {
        if (file.state !== 'new') { renderPanel(file); return }
        chosenFile = file
        returnFocus = trigger
        $('dialog-filename').textContent = file.name
        $('dialog-filemeta').textContent = `${file.size} · 学习资料 · Markdown 文件`
        $('knowledge-title').value = file.title
        $('title-error').hidden = true
        $('knowledge-title').removeAttribute('aria-invalid')
        $('add-dialog').showModal()
        $('knowledge-title').focus()
    } else if (trigger.dataset.action === 'retry') {
        file.state = 'pending'; file.step = 0
        renderFiles(); renderLibrary(); renderPanel(file)
        showToast('已重新提交，请查看处理进度')
    } else if (trigger.dataset.action === 'simulate') advance(file)
    else renderPanel(file)
})

$('add-form').addEventListener('submit', (event) => {
    event.preventDefault()
    const title = $('knowledge-title').value.trim()
    if (!title || Array.from(title).length > 200) {
        $('title-error').hidden = false
        $('knowledge-title').setAttribute('aria-invalid', 'true')
        $('knowledge-title').focus()
        return
    }
    if (!chosenFile || chosenFile.state !== 'new') return
    chosenFile.title = title
    chosenFile.state = 'pending'
    chosenFile.step = 0
    $('add-dialog').close()
    renderFiles(); renderLibrary(); renderPanel(chosenFile)
    showToast('已提交加入任务，可继续浏览文件')
})
for (const id of ['close-dialog', 'cancel-dialog']) $(id).addEventListener('click', () => $('add-dialog').close())
$('add-dialog').addEventListener('close', () => {
    if (returnFocus?.isConnected) returnFocus.focus()
    else document.querySelector('.nav-link.active')?.focus()
})
$('close-panel').addEventListener('click', () => { $('progress-panel').hidden = true; document.querySelector('.nav-link.active')?.focus() })
$('file-search').addEventListener('input', renderFiles)
$('knowledge-search').addEventListener('input', renderLibrary)
$('knowledge-filter').addEventListener('change', renderLibrary)
$('preview-empty').addEventListener('click', () => { $('knowledge-filter').value = 'empty'; $('progress-panel').hidden = true; location.hash = 'library' })
window.addEventListener('hashchange', navigate)
navigate()
