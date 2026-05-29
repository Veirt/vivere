import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  filterCharacters,
  fetchCharacter,
  firstAvailableUrl,
  formatBirthday,
  formatBloodType,
  formatElement,
  formatHeight,
  formatJob,
  formatWeight,
  sortCharacters,
  type Character,
  type CharacterAssets,
  type CharacterProfile,
  type CharacterSkill,
  type CharacterSkills,
  type CharacterStats,
  type UniqueWeapon,
} from './characters.ts'

afterEach(() => {
  vi.restoreAllMocks()
})

const characters: Array<Character> = [
  character({
    id: 2,
    name: 'Beta',
    rarityFlags: 3,
    elementType: 1,
    profile: { description: 'moon singer' },
  }),
  character({ id: 1, name: 'Alpha', rarityFlags: 5, elementType: 2, memo: 'healer' }),
  character({ id: 3, name: 'Gamma', title: 'Dawn Blade', rarityFlags: 5, elementType: 1 }),
]

describe('character helpers', () => {
  it('filters across display fields, profile data, skills, weapons, stories, and voices', () => {
    const withRichData = character({
      id: 4,
      name: 'Delta',
      skills: { active: [{ ...skill(), id: 42, name: 'Petal Storm' }] },
      uniqueWeapon: { name: 'Silver Harp' },
      stories: [{ id: 1, episodeId: 1, level: 0, rarityFlags: 1, title: 'Hidden memory', text: 'silver rain', titleKey: 'story.title', textKey: 'story.text', rewardItems: [], audioJp: [], audioUs: [] }],
      voices: [{ id: 1, category: 1, unlockCondition: 0, unlockQuestId: 0, sortOrder: 1, subtitle: 'A quiet line', subtitleKey: 'voice.subtitle', buttonText: 'Greeting', buttonTextKey: 'voice.button', audioJp: '', audioUs: '' }],
    })
    const allCharacters = [...characters, withRichData]

    expect(filterCharacters(allCharacters, 'moon')).toEqual([characters[0]])
    expect(filterCharacters(allCharacters, 'healer')).toEqual([characters[1]])
    expect(filterCharacters(allCharacters, 'Dawn')).toEqual([characters[2]])
    expect(filterCharacters(allCharacters, 'Petal Storm')).toEqual([withRichData])
    expect(filterCharacters(allCharacters, 'Silver Harp')).toEqual([withRichData])
    expect(filterCharacters(allCharacters, 'silver rain')).toEqual([withRichData])
    expect(filterCharacters(allCharacters, 'quiet line')).toEqual([withRichData])
  })

  it('sorts by prototype-compatible keys', () => {
    expect(sortCharacters(characters, 'id').map((character) => character.id)).toEqual([1, 2, 3])
    expect(sortCharacters(characters, 'name').map((character) => character.name)).toEqual([
      'Alpha',
      'Beta',
      'Gamma',
    ])
    expect(sortCharacters(characters, 'rarity').map((character) => character.id)).toEqual([1, 3, 2])
    expect(sortCharacters(characters, 'element').map((character) => character.id)).toEqual([2, 3, 1])
    expect(sortCharacters([character({ id: 4, nameKey: 'Zeta' }), character({ id: 5, name: 'Alpha' })], 'name').map((character) => character.id)).toEqual([5, 4])
  })

  it('formats compact profile values', () => {
    expect(formatBirthday(101)).toBe('01/01')
    expect(formatBirthday(1225)).toBe('12/25')
    expect(formatBirthday(0)).toBe('')
    expect(formatHeight(162)).toBe('162 cm')
    expect(formatWeight(45)).toBe('45 kg')
    expect(formatBloodType(4)).toBe('AB')
    expect(formatBloodType(0)).toBe('')
  })

  it('formats enum-like character metadata with safe fallbacks', () => {
    expect(formatElement(1)).toBe('Azure')
    expect(formatElement(99)).toBe('Element 99')
    expect(formatJob(2)).toBe('Sniper')
    expect(formatJob(99)).toBe('Class 99')
  })

  it('selects the first safe available regional URL by preference', () => {
    expect(firstAvailableUrl({ jaJP: 'https://example.test/jp', enUS: 'https://example.test/us' }, ['enUS', 'jaJP'])).toBe('https://example.test/us')
    expect(firstAvailableUrl({ jaJP: 'https://example.test/jp' }, ['enUS', 'jaJP'])).toBe('https://example.test/jp')
    expect(firstAvailableUrl({ enUS: 'javascript:alert(1)', jaJP: 'notaurl' }, ['enUS', 'jaJP'])).toBe('')
    expect(firstAvailableUrl(null, ['enUS'])).toBe('')
  })

  it('fetches one character without downloading the full catalog and rewrites asset paths', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify(character({ id: 17, iconPath: '/assets/icon.png', assets: { iconLarge: '/assets/large.png' } })), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    )

    const result = await fetchCharacter(17, 'TextResourceEnUsMB', undefined, 'https://api.example.test')

    expect(fetchMock).toHaveBeenCalledWith(new URL('/api/characters/17?lang=TextResourceEnUsMB', 'https://api.example.test'), { signal: undefined })
    expect(result?.iconPath).toBe('https://api.example.test/assets/icon.png')
    expect(result?.assets.iconLarge).toBe('https://api.example.test/assets/large.png')
  })

  it('rewrites nested skill, weapon, story, and voiceline asset paths', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify(character({
        id: 18,
        skills: { active: [{ ...skill(), id: 42, icon: '/assets/skill.png' }] },
        uniqueWeapon: { icon: '/assets/weapon.png' },
        stories: [{ id: 1, episodeId: 1, level: 0, rarityFlags: 1, title: 'Memory', text: '', titleKey: '', textKey: '', rewardItems: [], audioJp: ['/assets/memory.wav'], audioUs: [] }],
        voices: [{ id: 1, category: 1, unlockCondition: 0, unlockQuestId: 0, sortOrder: 2, subtitle: 'Hi', subtitleKey: '', buttonText: 'Greeting', buttonTextKey: '', audioJp: '/assets/voice.wav', audioUs: '' }],
      })), { status: 200, headers: { 'Content-Type': 'application/json' } }),
    )

    const result = await fetchCharacter(18, 'TextResourceEnUsMB', undefined, 'https://api.example.test')

    expect(result?.skills.active[0]?.icon).toBe('https://api.example.test/assets/skill.png')
    expect(result?.uniqueWeapon.icon).toBe('https://api.example.test/assets/weapon.png')
    expect(result?.stories[0]?.audioJp[0]).toBe('https://api.example.test/assets/memory.wav')
    expect(result?.voices[0]?.audioJp).toBe('https://api.example.test/assets/voice.wav')
  })

  it('returns null for missing or invalid character ids', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('{"error":"missing"}', { status: 404 }))

    await expect(fetchCharacter(Number.NaN, 'TextResourceEnUsMB', undefined, 'https://api.example.test')).resolves.toBeNull()
    await expect(fetchCharacter(999, 'TextResourceEnUsMB', undefined, 'https://api.example.test')).resolves.toBeNull()
    expect(fetchMock).toHaveBeenCalledTimes(1)
  })
})

type CharacterFixtureOverrides = Partial<Omit<Character, 'profile' | 'stats' | 'assets' | 'skills' | 'uniqueWeapon'>> & {
  profile?: Partial<CharacterProfile>
  stats?: Partial<CharacterStats>
  assets?: Partial<CharacterAssets>
  skills?: Partial<CharacterSkills>
  uniqueWeapon?: Partial<UniqueWeapon>
}

function character(overrides: CharacterFixtureOverrides): Character {
  return {
    id: 0,
    memo: '',
    name: '',
    title: '',
    nameKey: '',
    name2Key: '',
    rarityFlags: 0,
    elementType: 0,
    jobFlags: 0,
    characterType: 0,
    startTimeFixJST: '',
    endTimeFixJST: '',
    iconPath: '',
    stories: [],
    voices: [],
    ...overrides,
    skills: {
      normal: skill(),
      active: [],
      passive: [],
      ...overrides.skills,
    },
    uniqueWeapon: {
      equipmentId: 0,
      name: '',
      nameKey: '',
      iconId: 0,
      icon: '',
      jobFlags: 0,
      rarityFlags: 0,
      level: 0,
      slotType: 0,
      compositeId: 0,
      exclusiveEffectId: 0,
      skillDescriptionId: 0,
      battleParameter: {},
      bonusParameters: [],
      descriptions: { ssr: '', ur: '', lr: '' },
      ...overrides.uniqueWeapon,
    },
    profile: {
      birthday: 0,
      height: 0,
      weight: 0,
      bloodType: 0,
      description: '',
      voiceJp: '',
      voiceUs: '',
      vocalJp: '',
      vocalUs: '',
      lamentJp: '',
      lamentUs: '',
      lyricsJp: '',
      lyricsUs: '',
      gachaMessage: '',
      movieJpUrls: {},
      movieUsUrls: {},
      ...overrides.profile,
    },
    stats: {
      speed: 0,
      normalSkillId: 0,
      activeSkillIds: [],
      passiveSkillIds: [],
      baseParameterCoefficient: {},
      baseParameterGrossCoefficient: 0,
      initialBattleParameter: {},
      ...overrides.stats,
    },
    assets: {
      iconSmall: '',
      iconMedium: '',
      iconLarge: '',
      iconWide: '',
      iconEmotion: '',
      elementIcon: '',
      jobIcon: '',
      modelTexture: '',
      effectTexture: '',
      songJp: '',
      songUs: '',
      songShortJp: '',
      songShortUs: '',
      ...overrides.assets,
    },
  }
}

function skill(): CharacterSkill {
  return {
    id: 0,
    type: 'active',
    name: '',
    nameKey: '',
    icon: '',
    conditions: '',
    initCooltime: 0,
    maxCooltime: 0,
    levels: [],
  }
}
