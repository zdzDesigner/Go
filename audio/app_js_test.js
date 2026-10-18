const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const test = require('node:test')
const vm = require('node:vm')

const appSource = fs.readFileSync(path.join(__dirname, 'web/app.js'), 'utf8')

class Element {
  constructor(id = '') {
    this.id = id
    this.children = []
    this.dataset = {}
    this.classList = { toggle: () => {} }
    this.listeners = {}
    this.textContent = id === 'current-title' ? '尚未选择音频' : ''
    this.value = ''
    this.disabled = false
    this.currentTime = 0
    this.duration = 100
    this.ended = false
    this.src = ''
  }

  addEventListener(name, handler) {
    this.listeners[name] = handler
  }

  append(...children) {
    this.children.push(...children)
  }

  replaceChildren(...children) {
    this.children = children
  }

  load() {}

  play() {
    return Promise.resolve()
  }
}

async function boot(storageEntries = []) {
  const ids = [
    'login-panel', 'player-panel', 'library-panel', 'login-form', 'access-key',
    'login-error', 'library-status', 'audio-list', 'refresh-button', 'audio-player',
    'current-title', 'current-status', 'previous-button', 'next-button',
  ]
  const elements = Object.fromEntries(ids.map((id) => [id, new Element(id)]))
  const storage = new Map(storageEntries)
  const context = {
    console,
    Date,
    JSON,
    Number,
    Promise,
    Error,
    localStorage: {
      getItem: (key) => storage.get(key) ?? null,
      setItem: (key, value) => storage.set(key, value),
      removeItem: (key) => storage.delete(key),
    },
    document: {
      querySelector: (selector) => elements[selector.slice(1)],
      createElement: (tag) => new Element(tag),
      addEventListener: () => {},
      visibilityState: 'visible',
    },
    window: { addEventListener: () => {} },
    fetch: async () => ({
      ok: true,
      status: 200,
      json: async () => ({
        items: [{ id: 'chapter-1', name: '第一章', path: '第一章.mp3' }],
      }),
    }),
    setTimeout,
  }

  vm.runInNewContext(appSource, context)
  await new Promise((resolve) => setImmediate(resolve))
  return { elements, storage }
}

test('选择章节后保存最后章节和播放进度', async () => {
  const { elements, storage } = await boot()
  elements['audio-list'].children[0].listeners.click()
  elements['audio-player'].currentTime = 42
  elements['audio-player'].listeners.timeupdate()

  assert.equal(storage.get('audio-last-resource-id'), 'chapter-1')
  const progress = JSON.parse(storage.get('audio-progress:chapter-1'))
  assert.equal(progress.position, 42)
  assert.equal(progress.duration, 100)
  assert.equal(typeof progress.updatedAt, 'string')
})

test('重新打开 H5 后恢复最后章节和播放进度', async () => {
  const { elements } = await boot([
    ['audio-last-resource-id', 'chapter-1'],
    ['audio-progress:chapter-1', JSON.stringify({ position: 42, duration: 100 })],
  ])
  elements['audio-player'].listeners.loadedmetadata()
  await new Promise((resolve) => setImmediate(resolve))

  assert.equal(elements['audio-player'].src, '/api/audio/chapter-1/stream')
  assert.equal(elements['current-title'].textContent, '第一章')
  assert.equal(elements['audio-player'].currentTime, 42)
})
