// [输入]: Tauri Android 原生播放插件和现有 H5 元素。
// [输出]: server 配置、原生请求、播放控制和前台状态同步。
// [定位]: Android 适配层，普通网页不启用。
// [同步]: folder.md、../README.md。
if (window.__TAURI__) {
  const invoke = (command, args = {}) => window.__TAURI__.core.invoke(`plugin:playback|${command}`, args)
  const element = (id) => document.querySelector(`#${id}`)
  let state = { playing: false, positions: {} }
  let polling = false
  let changing = false
  let ready = false

  function showError(error) {
    const message = error instanceof Error ? error.message : String(error)
    element('native-error').textContent = message
    element('server-error').textContent = message
  }

  async function sync() {
    if (polling || changing || !ready || document.visibilityState === 'hidden') return
    polling = true
    try {
      state = await invoke('state')
      const item = audioItems.find((audio) => audio.id === state.id)
      currentAudioId = item ? item.id : ''
      if (item) {
        currentTitle.textContent = item.name
        currentStatus.textContent = item.path
      }
      element('native-toggle').textContent = state.playing ? '暂停' : '播放'
      element('native-toggle').disabled = !item
      const seek = element('native-seek')
      seek.max = String(state.duration)
      if (document.activeElement !== seek) seek.value = String(state.position)
      element('native-time').textContent = `${formatTime(state.position)} / ${formatTime(state.duration)}`
      element('native-error').textContent = state.error
      if (state.unauthorized) showLogin(state.error)
      for (const row of audioList.children) {
        const position = state.positions[row.dataset.audioId]
        if (Number.isFinite(position)) row.children[2].textContent = position > 0 ? `已播放 ${formatTime(position)}` : '未播放'
      }
      updateActiveItem()
      updateNavigationButtons()
    } finally {
      polling = false
    }
  }

  window.audioNative = {
    showError,
    progress(id) {
      const position = state.positions[id]
      return position > 0 ? { position } : null
    },
    async request(path, options = {}) {
      const key = options.body ? JSON.parse(options.body).key : undefined
      const response = await invoke('request', { path, key })
      return {
        status: response.status,
        ok: response.status >= 200 && response.status < 300,
        json: async () => JSON.parse(response.body),
      }
    },
    async play(items, id) {
      await invoke('play', { items: JSON.stringify(items), id })
      await sync()
    },
    async control(action, position) {
      await invoke('control', { action, position })
      await sync()
    },
    async adjacent(offset) {
      const index = audioItems.findIndex((item) => item.id === currentAudioId) + offset
      if (index < 0 || index >= audioItems.length) return
      if (!state.active) await this.play(audioItems, audioItems[index].id)
      else await this.control(offset < 0 ? 'previous' : 'next')
    },
    async restore() {
      await sync()
    },
    async boot() {
      setVisible(element('server-panel'), true)
      setVisible(loginPanel, false)
      setVisible(audioPlayer, false)
      setVisible(element('native-controls'), true)
      const config = await invoke('configure')
      element('server-address').value = config.address
      element('server-form').addEventListener('submit', async (event) => {
        event.preventDefault()
        if (changing) return
        changing = true
        try {
          await invoke('configure', { address: element('server-address').value.trim() })
          ready = true
          restoredLastAudio = false
          state = { playing: false, positions: {} }
          element('server-error').textContent = ''
          element('native-error').textContent = ''
          setVisible(loginPanel, false)
          setVisible(playerPanel, false)
          setVisible(libraryPanel, false)
          await loadLibrary()
        } catch (error) {
          showError(error)
        } finally {
          changing = false
        }
        await sync().catch(showError)
      })
      element('native-toggle').addEventListener('click', async () => {
        try {
          if (!state.active && currentAudioId) await window.audioNative.play(audioItems, currentAudioId)
          else await window.audioNative.control(state.playing ? 'pause' : 'resume')
        } catch (error) { showError(error) }
      })
      element('native-seek').addEventListener('change', () => {
        window.audioNative.control('seek', Number(element('native-seek').value)).catch(showError)
      })
      document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'visible') sync().catch(showError)
      })
      setInterval(() => sync().catch(showError), 1000)
      if (config.address) {
        ready = true
        await loadLibrary()
      }
    },
  }
}
