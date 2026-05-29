export const languages = [
  { value: 'TextResourceEnUsMB', label: 'English' },
  { value: 'TextResourceJaJpMB', label: 'Japanese' },
  { value: 'TextResourceZhTwMB', label: 'Chinese (TW)' },
  { value: 'TextResourceKoKrMB', label: 'Korean' },
] as const

export type Language = (typeof languages)[number]['value']

export const DEFAULT_LANGUAGE: Language = 'TextResourceEnUsMB'

export function isLanguage(value: unknown): value is Language {
  return typeof value === 'string' && languages.some((language) => language.value === value)
}

export const sortOptions = [
  { value: 'id', label: 'Sort by id' },
  { value: 'name', label: 'Sort by name' },
  { value: 'rarity', label: 'Sort by rarity' },
  { value: 'element', label: 'Sort by element' },
] as const

export type SortKey = (typeof sortOptions)[number]['value']

export function isSortKey(value: unknown): value is SortKey {
  return typeof value === 'string' && sortOptions.some((option) => option.value === value)
}

export type CharacterProfile = {
  birthday: number
  height: number
  weight: number
  bloodType: number
  description: string
  voiceJp: string
  voiceUs: string
  vocalJp: string
  vocalUs: string
  lamentJp: string
  lamentUs: string
  lyricsJp: string
  lyricsUs: string
  gachaMessage: string
  movieJpUrls: Record<string, string>
  movieUsUrls: Record<string, string>
}

export type CharacterSkillLevel = {
  orderNumber: number
  characterLevel: number
  equipmentRarityFlags: number
  description: string
  descriptionKey: string
  subSkillIds: Array<number>
}

export type CharacterSkill = {
  id: number
  type: 'active' | 'passive'
  name: string
  nameKey: string
  icon: string
  conditions: string
  initCooltime: number
  maxCooltime: number
  levels: Array<CharacterSkillLevel>
}

export type CharacterSkills = {
  normal: CharacterSkill
  active: Array<CharacterSkill>
  passive: Array<CharacterSkill>
}

export type ParameterChange = {
  category: 'base' | 'battle' | string
  type: number
  changeType: number
  value: number
}

export type UniqueWeapon = {
  equipmentId: number
  name: string
  nameKey: string
  iconId: number
  icon: string
  jobFlags: number
  rarityFlags: number
  level: number
  slotType: number
  compositeId: number
  exclusiveEffectId: number
  skillDescriptionId: number
  battleParameter: Record<string, number>
  bonusParameters: Array<ParameterChange>
  descriptions: {
    ssr: string
    ur: string
    lr: string
  }
}

export type CharacterStats = {
  speed: number
  normalSkillId: number
  activeSkillIds: Array<number>
  passiveSkillIds: Array<number>
  baseParameterCoefficient: Record<string, number>
  baseParameterGrossCoefficient: number
  initialBattleParameter: Record<string, number>
}

export type CharacterAssets = {
  iconSmall: string
  iconMedium: string
  iconLarge: string
  iconWide: string
  iconEmotion: string
  elementIcon: string
  jobIcon: string
  modelTexture: string
  effectTexture: string
  songJp: string
  songUs: string
  songShortJp: string
  songShortUs: string
}

export type CharacterStory = {
  id: number
  episodeId: number
  level: number
  rarityFlags: number
  title: string
  text: string
  titleKey: string
  textKey: string
  rewardItems: Array<{ itemId: number; itemType: number; itemCount: number }>
  audioJp: Array<string>
  audioUs: Array<string>
}

export type CharacterVoice = {
  id: number
  category: number
  unlockCondition: number
  unlockQuestId: number
  sortOrder: number
  subtitle: string
  subtitleKey: string
  buttonText: string
  buttonTextKey: string
  audioJp: string
  audioUs: string
}

export type Character = {
  id: number
  memo: string
  name: string
  title: string
  nameKey: string
  name2Key: string
  rarityFlags: number
  elementType: number
  jobFlags: number
  characterType: number
  startTimeFixJST: string
  endTimeFixJST: string
  profile: CharacterProfile
  stats: CharacterStats
  assets: CharacterAssets
  skills: CharacterSkills
  uniqueWeapon: UniqueWeapon
  stories: Array<CharacterStory>
  voices: Array<CharacterVoice>
  iconPath: string
}

const DEFAULT_API_ORIGIN = 'http://localhost:8080'

const assetFields = [
  'iconSmall',
  'iconMedium',
  'iconLarge',
  'iconWide',
  'iconEmotion',
  'elementIcon',
  'jobIcon',
  'modelTexture',
  'effectTexture',
  'songJp',
  'songUs',
  'songShortJp',
  'songShortUs',
] as const

const elementLabels: Record<number, string> = {
  1: 'Azure',
  2: 'Crimson',
  3: 'Emerald',
  4: 'Amber',
  5: 'Radiance',
  6: 'Chaos',
}

const jobLabels: Record<number, string> = {
  1: 'Warrior',
  2: 'Sniper',
  4: 'Sorcerer',
}

const bloodTypeLabels: Record<number, string> = {
  1: 'A',
  2: 'B',
  3: 'O',
  4: 'AB',
}

export function getApiOrigin(origin?: string) {
  return origin || import.meta.env.VITE_API_ORIGIN || DEFAULT_API_ORIGIN
}

function withAssetOrigin(character: Character, apiOrigin: string): Character {
  const assets = { ...character.assets }
  for (const field of assetFields) {
    assets[field] = withOrigin(assets[field], apiOrigin)
  }

  return {
    ...character,
    iconPath: withOrigin(character.iconPath, apiOrigin),
    assets,
    skills: {
      normal: withSkillAssetOrigin(character.skills.normal, apiOrigin),
      active: character.skills.active.map((skill) => withSkillAssetOrigin(skill, apiOrigin)),
      passive: character.skills.passive.map((skill) => withSkillAssetOrigin(skill, apiOrigin)),
    },
    uniqueWeapon: {
      ...character.uniqueWeapon,
      icon: withOrigin(character.uniqueWeapon.icon, apiOrigin),
    },
    stories: character.stories.map((story) => ({
      ...story,
      audioJp: story.audioJp.map((path) => withOrigin(path, apiOrigin)),
      audioUs: story.audioUs.map((path) => withOrigin(path, apiOrigin)),
    })),
    voices: character.voices.map((voice) => ({
      ...voice,
      audioJp: withOrigin(voice.audioJp, apiOrigin),
      audioUs: withOrigin(voice.audioUs, apiOrigin),
    })),
  }
}

function withSkillAssetOrigin(skill: CharacterSkill, apiOrigin: string): CharacterSkill {
  return { ...skill, icon: withOrigin(skill.icon, apiOrigin) }
}

function withOrigin(path: string, origin: string) {
  if (!path || /^[a-z][a-z\d+.-]*:/i.test(path)) {
    return path
  }
  return new URL(path, origin).toString()
}

export async function fetchCharacters(language: Language, signal?: AbortSignal, origin?: string): Promise<Array<Character>> {
  const apiOrigin = getApiOrigin(origin)
  const path = `/api/characters?lang=${encodeURIComponent(language)}`
  const response = await fetch(new URL(path, apiOrigin), { signal })
  if (!response.ok) {
    throw new Error((await response.text()) || `Failed to load characters (${response.status})`)
  }
  const characters = (await response.json()) as Array<Character>
  return characters.map((character) => withAssetOrigin(character, apiOrigin))
}

export async function fetchCharacter(characterId: number, language: Language, signal?: AbortSignal, origin?: string) {
  if (!Number.isInteger(characterId)) {
    return null
  }

  const apiOrigin = getApiOrigin(origin)
  const path = `/api/characters/${characterId}?lang=${encodeURIComponent(language)}`
  const response = await fetch(new URL(path, apiOrigin), { signal })
  if (response.status === 404) {
    return null
  }
  if (!response.ok) {
    throw new Error((await response.text()) || `Failed to load character (${response.status})`)
  }
  return withAssetOrigin((await response.json()) as Character, apiOrigin)
}

export function filterCharacters(characters: Array<Character>, query: string) {
  const normalizedQuery = query.trim().toLowerCase()
  if (!normalizedQuery) {
    return characters
  }
  return characters.filter((character) =>
    JSON.stringify([
      character.name,
      character.title,
      character.memo,
      character.profile,
      character.skills,
      character.uniqueWeapon,
      character.stories,
      character.voices,
    ])
      .toLowerCase()
      .includes(normalizedQuery),
  )
}

export function sortCharacters(characters: Array<Character>, sort: SortKey) {
  return [...characters].sort((a, b) => {
    if (sort === 'name') {
      return displayName(a).localeCompare(displayName(b))
    }
    if (sort === 'rarity') {
      return b.rarityFlags - a.rarityFlags || a.id - b.id
    }
    if (sort === 'element') {
      return a.elementType - b.elementType || a.id - b.id
    }
    return a.id - b.id
  })
}

export function formatBirthday(value: number) {
  if (!value) {
    return ''
  }
  const raw = String(value).padStart(4, '0')
  return `${raw.slice(0, -2)}/${raw.slice(-2)}`
}

export function formatHeight(value: number) {
  return value ? `${value} cm` : ''
}

export function formatWeight(value: number) {
  return value ? `${value} kg` : ''
}

export function formatElement(value: number) {
  return elementLabels[value] ?? `Element ${value}`
}

export function formatJob(value: number) {
  return jobLabels[value] ?? `Class ${value}`
}

export function formatBloodType(value: number) {
  return bloodTypeLabels[value] ?? (value ? `Type ${value}` : '')
}

export function displayName(character: Pick<Character, 'name' | 'nameKey'>) {
  return character.name || character.nameKey
}

export function firstAvailableUrl(urls: Record<string, string> | null | undefined, preferredKeys: Array<string>) {
  if (!urls) {
    return ''
  }
  for (const key of preferredKeys) {
    const value = safeHttpUrl(urls[key])
    if (value) {
      return value
    }
  }
  for (const value of Object.values(urls)) {
    const safeValue = safeHttpUrl(value)
    if (safeValue) {
      return safeValue
    }
  }
  return ''
}

function safeHttpUrl(value: string | undefined) {
  if (!value) {
    return ''
  }
  try {
    const url = new URL(value)
    return url.protocol === 'https:' || url.protocol === 'http:' ? url.toString() : ''
  } catch {
    return ''
  }
}
