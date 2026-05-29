import type { Character, CharacterStory, CharacterVoice } from './characters.ts'

export type AudioSource = {
  label: string
  src: string
}

export function preferredAutoplayAudio(character: Character) {
  return character.assets.songShortJp || character.assets.songJp || character.stories[0]?.audioJp[0] || character.voices[0]?.audioJp || ''
}

export function storyAudioSources(story: CharacterStory): Array<AudioSource> {
  return [
    ...story.audioJp.map((src, index) => ({ label: `JP segment ${index + 1}`, src })),
    ...story.audioUs.map((src, index) => ({ label: `US segment ${index + 1}`, src })),
  ]
}

export function voiceAudioSources(voice: CharacterVoice): Array<AudioSource> {
  return [
    voice.audioJp ? { label: 'JP voice', src: voice.audioJp } : null,
    voice.audioUs ? { label: 'US voice', src: voice.audioUs } : null,
  ].filter((source): source is AudioSource => Boolean(source))
}
