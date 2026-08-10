export type ServiceStatus = 'online' | 'offline' | 'unknown'

export interface ServiceItem {
  id: string
  name: string
  description: string
  icon: string
  tags: string[]
  internalUrl: string
  externalUrl: string
  status: ServiceStatus
}

export interface ServiceGroup {
  id: string
  name: string
  icon: string
  collapsed?: boolean
  items: ServiceItem[]
}

export const mockGroups: ServiceGroup[] = [
  {
    id: 'media',
    name: 'Media',
    icon: 'mdi:movie-open-outline',
    items: [
      {
        id: 'plex',
        name: 'Plex',
        description: 'Media Server',
        icon: 'mdi:plex',
        tags: ['movie', 'video', 'media'],
        internalUrl: 'http://192.168.1.10:32400',
        externalUrl: 'https://plex.example.com',
        status: 'online',
      },
      {
        id: 'jellyfin',
        name: 'Jellyfin',
        description: 'Free Media System',
        icon: 'mdi:jellyfish-outline',
        tags: ['movie', 'video', 'media'],
        internalUrl: 'http://192.168.1.10:8096',
        externalUrl: 'https://jellyfin.example.com',
        status: 'online',
      },
      {
        id: 'navidrome',
        name: 'Navidrome',
        description: 'Music Streaming',
        icon: 'mdi:music-circle-outline',
        tags: ['music', 'audio', 'media'],
        internalUrl: 'http://192.168.1.10:4533',
        externalUrl: 'https://music.example.com',
        status: 'unknown',
      },
    ],
  },
  {
    id: 'knowledge',
    name: 'Knowledge',
    icon: 'mdi:book-open-page-variant-outline',
    items: [
      {
        id: 'noteverse',
        name: 'NoteVerse',
        description: '知识管理 · 文元',
        icon: 'mdi:notebook-outline',
        tags: ['notes', 'knowledge'],
        internalUrl: 'http://192.168.1.10:3001',
        externalUrl: 'https://notes.example.com',
        status: 'online',
      },
      {
        id: 'bookverse',
        name: 'BookVerse',
        description: '个人书库',
        icon: 'mdi:book-open-outline',
        tags: ['books', 'reading'],
        internalUrl: 'http://192.168.1.10:3002',
        externalUrl: 'https://books.example.com',
        status: 'online',
      },
    ],
  },
  {
    id: 'system',
    name: 'System',
    icon: 'mdi:server-outline',
    items: [
      {
        id: 'dsm',
        name: 'DSM',
        description: 'NAS Management',
        icon: 'mdi:nas',
        tags: ['nas', 'system'],
        internalUrl: 'https://192.168.1.10:5001',
        externalUrl: 'https://nas.example.com',
        status: 'online',
      },
      {
        id: 'docker',
        name: 'Portainer',
        description: 'Container Console',
        icon: 'mdi:docker',
        tags: ['docker', 'system'],
        internalUrl: 'http://192.168.1.10:9000',
        externalUrl: 'https://docker.example.com',
        status: 'offline',
      },
      {
        id: 'files',
        name: 'FileBrowser',
        description: '文件管理',
        icon: 'mdi:folder-outline',
        tags: ['files', 'system'],
        internalUrl: 'http://192.168.1.10:8081',
        externalUrl: 'https://files.example.com',
        status: 'online',
      },
    ],
  },
]
