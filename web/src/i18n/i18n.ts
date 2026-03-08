import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

import enPlayer from './en/player.json'
import enAdmin from './en/admin.json'
import ruPlayer from './ru/player.json'
import ruAdmin from './ru/admin.json'

i18n.use(initReactI18next).init({
  resources: {
    en: { player: enPlayer, admin: enAdmin },
    ru: { player: ruPlayer, admin: ruAdmin },
  },
  lng: 'ru',
  fallbackLng: 'ru',
  ns: ['player', 'admin'],
  defaultNS: 'player',
  interpolation: { escapeValue: false },
})

export default i18n
