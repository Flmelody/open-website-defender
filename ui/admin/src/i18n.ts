import { createI18n } from 'vue-i18n'
import en from './locales/en'
import zh from './locales/zh'
import de from './locales/de'
import fr from './locales/fr'
import ja from './locales/ja'
import ru from './locales/ru'

const savedLocale = localStorage.getItem('locale') || 'en'

const i18n = createI18n({
  legacy: false,
  locale: savedLocale,
  fallbackLocale: 'en',
  messages: {
    en,
    zh,
    de,
    fr,
    ja,
    ru
  }
})

export default i18n
