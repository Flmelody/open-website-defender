import { DataBoard, Moon, Sunny, User, Lock, Document, Location, Key, Setting, Link, Connection } from '@element-plus/icons-vue'
import type { Component } from 'vue'

export interface MenuItem {
  index: string
  icon: Component
  label: string
}

export const menuItems: MenuItem[] = [
  {
    index: '/dashboard',
    icon: DataBoard,
    label: 'menu.dashboard'
  },
  {
    index: '/waf-rules',
    icon: Lock,
    label: 'menu.waf_rules'
  },
  {
    index: '/access-logs',
    icon: Document,
    label: 'menu.access_logs'
  },
  {
    index: '/ip-white-list',
    icon: Sunny,
    label: 'menu.ip_white_list'
  },
  {
    index: '/ip-black-list',
    icon: Moon,
    label: 'menu.ip_black_list'
  },
  {
    index: '/geo-block',
    icon: Location,
    label: 'menu.geo_block'
  },
  {
    index: '/authorized-domains',
    icon: Link,
    label: 'menu.authorized_domains'
  },
  {
    index: '/licenses',
    icon: Key,
    label: 'menu.licenses'
  },
  {
    index: '/oauth-clients',
    icon: Connection,
    label: 'menu.oauth_clients'
  },
  {
    index: '/users',
    icon: User,
    label: 'menu.users_db'
  },
  {
    index: '/system-settings',
    icon: Setting,
    label: 'menu.system_settings'
  }
]
