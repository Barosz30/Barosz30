const { contextBridge } = require('electron')

contextBridge.exposeInMainWorld('boxelApp', {
  isApp: true,
  serverUrl: 'ws://127.0.0.1:8787',
})
