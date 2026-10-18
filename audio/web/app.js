const loginPanel = document.querySelector('#login-panel')
const playerPanel = document.querySelector('#player-panel')
const libraryPanel = document.querySelector('#library-panel')
const loginForm = document.querySelector('#login-form')
const accessKeyInput = document.querySelector('#access-key')
const loginError = document.querySelector('#login-error')
const libraryStatus = document.querySelector('#library-status')
const audioList = document.querySelector('#audio-list')
const refreshButton = document.querySelector('#refresh-button')
const audioPlayer = document.querySelector('#audio-player')
const currentTitle = document.querySelector('#current-title')
const currentStatus = document.querySelector('#current-status')
const previousButton = document.querySelector('#previous-button')
const nextButton = document.querySelector('#next-button')

const lastAudioKey = 'audio-last-resource-id'
let audioItems = []
let currentAudioId = ''
let pendingResumeId = ''
let lastSavedAt = 0
let restoredLastAudio = false

function progressKey(id) {
  return `audio-progress:${id}`
}

function readProgress(id) {
  try {
    const raw = localStorage.getItem(progressKey(id))
    if (!raw) return null
    const progress = JSON.parse(raw)
    if (!Number.isFinite(progress.position) || progress.position < 0) return null
    return progress
  } catch (error) {
    console.warn('读取播放进度失败', error)
    return null
  }
}

function readLastAudioId() {
  try {
    return localStorage.getItem(lastAudioKey) || ''
  } catch (error) {
    console.warn('读取最后播放章节失败', error)
    return ''
  }
}

function saveLastAudioId(id) {
  try {
    localStorage.setItem(lastAudioKey, id)
  } catch (error) {
    console.warn('保存最后播放章节失败', error)
  }
}

function saveProgress() {
  if (!currentAudioId || !Number.isFinite(audioPlayer.currentTime)) return
  if (audioPlayer.ended || !Number.isFinite(audioPlayer.duration)) return

  const now = Date.now()
  if (now - lastSavedAt < 4000) return
  lastSavedAt = now
  try {
    localStorage.setItem(progressKey(currentAudioId), JSON.stringify({
      position: audioPlayer.currentTime,
      duration: audioPlayer.duration,
      updatedAt: new Date(now).toISOString(),
    }))
  } catch (error) {
    console.warn('保存播放进度失败', error)
  }
}

function saveProgressImmediately() {
  const previousSavedAt = lastSavedAt
  lastSavedAt = 0
  saveProgress()
  lastSavedAt = Math.max(previousSavedAt, lastSavedAt)
}

function formatTime(seconds) {
  if (!Number.isFinite(seconds) || seconds < 0) return '--:--'
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = Math.floor(seconds % 60).toString().padStart(2, '0')
  return `${minutes}:${remainingSeconds}`
}

function setVisible(element, visible) {
  element.classList.toggle('hidden', !visible)
}

function showLogin(message = '') {
  setVisible(loginPanel, true)
  setVisible(playerPanel, false)
  setVisible(libraryPanel, false)
  loginError.textContent = message
}

function showLibrary() {
  setVisible(loginPanel, false)
  setVisible(playerPanel, true)
  setVisible(libraryPanel, true)
}

async function request(url, options = {}) {
  const response = await fetch(url, options)
  if (response.status === 401) {
    let message = '访问会话已失效，请重新输入共享访问密钥。'
    try {
      const body = await response.json()
      if (body.error) message = body.error
    } catch (error) {
      console.warn('读取认证错误响应失败', error)
    }
    if (url !== '/api/access') showLogin(message)
    throw new Error(message)
  }
  if (!response.ok) {
    let message = `请求失败（${response.status}）`
    try {
      const body = await response.json()
      if (body.error) message = body.error
    } catch (error) {
      console.warn('读取错误响应失败', error)
    }
    throw new Error(message)
  }
  return response
}

async function login(event) {
  event.preventDefault()
  loginError.textContent = ''
  const key = accessKeyInput.value.trim()
  if (!key) return

  try {
    await request('/api/access', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ key }),
    })
    accessKeyInput.value = ''
    await loadLibrary()
  } catch (error) {
    loginError.textContent = error.message
  }
}

async function loadLibrary() {
  libraryStatus.textContent = '正在扫描音频库……'
  try {
    const response = await request('/api/library')
    const body = await response.json()
    audioItems = Array.isArray(body.items) ? body.items : []
    renderLibrary()
    showLibrary()
    if (!restoredLastAudio) {
      restoredLastAudio = true
      restoreLastAudio()
    }
  } catch (error) {
    if (error.message !== '访问会话已失效') {
      libraryStatus.textContent = error.message
    }
  }
}

function renderLibrary() {
  audioList.replaceChildren()
  if (audioItems.length === 0) {
    libraryStatus.textContent = '音频库为空，请将音频文件放入 server 配置的目录。'
    updateNavigationButtons()
    return
  }
  libraryStatus.textContent = `共 ${audioItems.length} 个音频资源`

  for (const item of audioItems) {
    const progress = readProgress(item.id)
    const button = document.createElement('button')
    button.type = 'button'
    button.className = 'audio-item'
    button.dataset.audioId = item.id

    const name = document.createElement('strong')
    name.textContent = item.name
    const path = document.createElement('span')
    path.className = 'audio-path'
    path.textContent = item.path
    const position = document.createElement('span')
    position.className = 'audio-position'
    position.textContent = progress ? `已播放 ${formatTime(progress.position)}` : '未播放'

    button.append(name, path, position)
    button.addEventListener('click', () => playAudio(item))
    audioList.append(button)
  }
  updateNavigationButtons()
}

function updateActiveItem() {
  for (const item of audioList.children) {
    item.classList.toggle('active', item.dataset.audioId === currentAudioId)
  }
}

function currentAudioIndex() {
  return audioItems.findIndex((item) => item.id === currentAudioId)
}

function updateNavigationButtons() {
  const index = currentAudioIndex()
  previousButton.disabled = index <= 0
  nextButton.disabled = index < 0 || index >= audioItems.length - 1
}

function restoreLastAudio() {
  const lastAudioId = readLastAudioId()
  if (!lastAudioId) return

  const item = audioItems.find((audio) => audio.id === lastAudioId)
  if (!item) {
    try {
      localStorage.removeItem(lastAudioKey)
    } catch (error) {
      console.warn('清理不存在的最后播放章节失败', error)
    }
    return
  }
  playAudio(item)
}

function playAudio(item) {
  saveProgressImmediately()
  currentAudioId = item.id
  saveLastAudioId(currentAudioId)
  pendingResumeId = item.id
  lastSavedAt = 0
  currentTitle.textContent = item.name
  currentStatus.textContent = item.path
  audioPlayer.src = `/api/audio/${encodeURIComponent(item.id)}/stream`
  audioPlayer.load()
  updateActiveItem()
  updateNavigationButtons()
}

function playAdjacent(offset) {
  const index = currentAudioIndex()
  const nextIndex = index + offset
  if (index < 0 || nextIndex < 0 || nextIndex >= audioItems.length) return
  playAudio(audioItems[nextIndex])
}

audioPlayer.addEventListener('loadedmetadata', () => {
  if (pendingResumeId !== currentAudioId) return
  const progress = readProgress(currentAudioId)
  if (progress && progress.position < audioPlayer.duration - 2) {
    audioPlayer.currentTime = progress.position
    currentStatus.textContent = `${currentTitle.textContent}，从 ${formatTime(progress.position)} 继续`
  }
  pendingResumeId = ''
  audioPlayer.play().catch(() => {
    currentStatus.textContent = '已载入音频，请点击播放按钮开始。'
  })
})

audioPlayer.addEventListener('timeupdate', saveProgress)
audioPlayer.addEventListener('pause', saveProgressImmediately)
audioPlayer.addEventListener('ended', () => {
  if (currentAudioId) localStorage.removeItem(progressKey(currentAudioId))
  const index = currentAudioIndex()
  renderLibrary()
  if (index >= 0 && index < audioItems.length - 1) {
    currentStatus.textContent = '播放完成，正在播放下一个。'
    playAudio(audioItems[index + 1])
    return
  }
  currentStatus.textContent = '播放完成，已到列表末尾。下次将从头开始。'
})

document.addEventListener('visibilitychange', () => {
  if (document.visibilityState === 'hidden') saveProgressImmediately()
})

window.addEventListener('beforeunload', saveProgressImmediately)
window.addEventListener('pagehide', saveProgressImmediately)
loginForm.addEventListener('submit', login)
refreshButton.addEventListener('click', loadLibrary)
previousButton.addEventListener('click', () => playAdjacent(-1))
nextButton.addEventListener('click', () => playAdjacent(1))

loadLibrary().catch((error) => {
  console.warn('初始化音频库失败', error)
})
