import { Moon, Sunny, User } from '@element-plus/icons-vue'
import type { Component } from 'vue'

export interface MenuItem {
  index: string
  icon: Component
  label: string
}

export const menuItems: MenuItem[] = [
  {
    index: '/users',
    icon: User,
    label: 'menu.users_db'
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
  }
]

