import { app, BrowserWindow, Menu } from 'electron'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))

function createWindow() {
  const win = new BrowserWindow({
    width: 1440,
    height: 900,
    minWidth: 960,
    minHeight: 640,
    title: 'BOXEL',
    backgroundColor: '#1a1228',
    autoHideMenuBar: true,
    webPreferences: {
      preload: path.join(here, 'preload.cjs'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  })

  const url = process.env.BOXEL_URL || 'http://127.0.0.1:5173/'
  void win.loadURL(url)
}

const template = [
  {
    label: 'BOXEL',
    submenu: [
      { label: 'Nowe okno (drugi gracz)', click: () => createWindow() },
      { type: 'separator' },
      { role: 'quit' },
    ],
  },
]

app.whenReady().then(() => {
  Menu.setApplicationMenu(Menu.buildFromTemplate(template))
  createWindow()
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})
