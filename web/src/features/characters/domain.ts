import { Array, Effect, Option, Order, Record, Schema, String as EffectString, pipe } from "effect";
import MiniSearch from "minisearch";

export const languages = [
  { value: "TextResourceEnUsMB", label: "English" },
  { value: "TextResourceJaJpMB", label: "Japanese" },
  { value: "TextResourceZhTwMB", label: "Chinese (TW)" },
  { value: "TextResourceKoKrMB", label: "Korean" },
] as const;

export const Language = Schema.Literals([
  "TextResourceEnUsMB",
  "TextResourceJaJpMB",
  "TextResourceZhTwMB",
  "TextResourceKoKrMB",
]);

export type Language = typeof Language.Type;

export const DEFAULT_LANGUAGE: Language = "TextResourceEnUsMB";

export const isLanguage = Schema.is(Language);

export const sortOptions = [
  { value: "id", label: "Sort by id" },
  { value: "name", label: "Sort by name" },
  { value: "rarity", label: "Sort by rarity" },
  { value: "element", label: "Sort by element" },
] as const;

export const SortKey = Schema.Literals(["id", "name", "rarity", "element"]);

export type SortKey = typeof SortKey.Type;

export const isSortKey = Schema.is(SortKey);

function searchStringWithFallback<A extends string>(
  schema: Schema.Codec<A, A>,
  fallback: NoInfer<A>,
) {
  return schema.pipe(
    Schema.catchDecoding(() => Effect.succeedSome(fallback)),
    Schema.withDecodingDefault(Effect.succeed(fallback)),
  );
}

export const CharacterIndexSearchSchema = Schema.Struct({
  lang: searchStringWithFallback(Language, DEFAULT_LANGUAGE),
  q: searchStringWithFallback(Schema.String, ""),
  sort: searchStringWithFallback(SortKey, "id"),
});

export type CharacterIndexSearch = typeof CharacterIndexSearchSchema.Type;

export const validateCharacterIndexSearch = Schema.toStandardSchemaV1(CharacterIndexSearchSchema);

export const CharacterDetailSearchSchema = Schema.Struct({
  lang: searchStringWithFallback(Language, DEFAULT_LANGUAGE),
});

export type CharacterDetailSearch = typeof CharacterDetailSearchSchema.Type;

export const validateCharacterDetailSearch = Schema.toStandardSchemaV1(CharacterDetailSearchSchema);

export class CharacterNotFoundError extends Schema.TaggedError<
  CharacterNotFoundError,
  { readonly brand: unique symbol }
>()("CharacterNotFoundError", {
  characterId: Schema.String,
}) {}

export class CharacterApiError extends Schema.TaggedError<
  CharacterApiError,
  { readonly brand: unique symbol }
>()("CharacterApiError", {
  endpoint: Schema.String,
  reason: Schema.String,
  status: Schema.optionalKey(Schema.Int),
}) {}

export class CharacterDecodeError extends Schema.TaggedError<
  CharacterDecodeError,
  { readonly brand: unique symbol }
>()("CharacterDecodeError", {
  endpoint: Schema.String,
  reason: Schema.String,
}) {}

export type CharacterFetchError = CharacterNotFoundError | CharacterApiError | CharacterDecodeError;

export const CharacterProfile = Schema.Struct({
  birthday: Schema.Finite,
  height: Schema.Finite,
  weight: Schema.Finite,
  bloodType: Schema.Finite,
  description: Schema.String,
  voiceJp: Schema.String,
  voiceUs: Schema.String,
  vocalJp: Schema.String,
  vocalUs: Schema.String,
  lamentJp: Schema.String,
  lamentUs: Schema.String,
  lyricsJp: Schema.String,
  lyricsUs: Schema.String,
  gachaMessage: Schema.String,
  movieJpUrls: Schema.Record(Schema.String, Schema.String),
  movieUsUrls: Schema.Record(Schema.String, Schema.String),
});

export type CharacterProfile = typeof CharacterProfile.Type;

export const CharacterSkillLevel = Schema.Struct({
  orderNumber: Schema.Finite,
  characterLevel: Schema.Finite,
  equipmentRarityFlags: Schema.Finite,
  description: Schema.String,
  descriptionKey: Schema.String,
  subSkillIds: Schema.Array(Schema.Finite),
});

export type CharacterSkillLevel = typeof CharacterSkillLevel.Type;

export const CharacterSkill = Schema.Struct({
  id: Schema.Finite,
  type: Schema.Literals(["active", "passive"]),
  name: Schema.String,
  nameKey: Schema.String,
  icon: Schema.String,
  conditions: Schema.String,
  initCooltime: Schema.Finite,
  maxCooltime: Schema.Finite,
  levels: Schema.Array(CharacterSkillLevel),
});

export type CharacterSkill = typeof CharacterSkill.Type;

export const CharacterSkills = Schema.Struct({
  normal: CharacterSkill,
  active: Schema.Array(CharacterSkill),
  passive: Schema.Array(CharacterSkill),
});

export type CharacterSkills = typeof CharacterSkills.Type;

export const ParameterChange = Schema.Struct({
  category: Schema.String,
  type: Schema.Finite,
  changeType: Schema.Finite,
  value: Schema.Finite,
});

export type ParameterChange = typeof ParameterChange.Type;

export const UniqueWeaponDescriptions = Schema.Struct({
  ssr: Schema.String,
  ur: Schema.String,
  lr: Schema.String,
});

export type UniqueWeaponDescriptions = typeof UniqueWeaponDescriptions.Type;

export const UniqueWeapon = Schema.Struct({
  equipmentId: Schema.Finite,
  name: Schema.String,
  nameKey: Schema.String,
  iconId: Schema.Finite,
  icon: Schema.String,
  jobFlags: Schema.Finite,
  rarityFlags: Schema.Finite,
  level: Schema.Finite,
  slotType: Schema.Finite,
  compositeId: Schema.Finite,
  exclusiveEffectId: Schema.Finite,
  skillDescriptionId: Schema.Finite,
  battleParameter: Schema.Record(Schema.String, Schema.Finite),
  bonusParameters: Schema.Array(ParameterChange),
  descriptions: UniqueWeaponDescriptions,
});

export type UniqueWeapon = typeof UniqueWeapon.Type;

export const CharacterStats = Schema.Struct({
  speed: Schema.Finite,
  normalSkillId: Schema.Finite,
  activeSkillIds: Schema.Array(Schema.Finite),
  passiveSkillIds: Schema.Array(Schema.Finite),
  baseParameterCoefficient: Schema.Record(Schema.String, Schema.Finite),
  baseParameterGrossCoefficient: Schema.Finite,
  initialBattleParameter: Schema.Record(Schema.String, Schema.Finite),
});

export type CharacterStats = typeof CharacterStats.Type;

export const CharacterAssets = Schema.Struct({
  iconSmall: Schema.String,
  iconMedium: Schema.String,
  iconLarge: Schema.String,
  iconWide: Schema.String,
  iconEmotion: Schema.String,
  elementIcon: Schema.String,
  jobIcon: Schema.String,
  modelTexture: Schema.String,
  effectTexture: Schema.String,
  songJp: Schema.String,
  songUs: Schema.String,
  songShortJp: Schema.String,
  songShortUs: Schema.String,
});

export type CharacterAssets = typeof CharacterAssets.Type;

export const RewardItem = Schema.Struct({
  itemId: Schema.Finite,
  itemType: Schema.Finite,
  itemCount: Schema.Finite,
});

export type RewardItem = typeof RewardItem.Type;

export const CharacterStory = Schema.Struct({
  id: Schema.Finite,
  episodeId: Schema.Finite,
  level: Schema.Finite,
  rarityFlags: Schema.Finite,
  title: Schema.String,
  text: Schema.String,
  titleKey: Schema.String,
  textKey: Schema.String,
  rewardItems: Schema.Array(RewardItem),
  audioJp: Schema.Array(Schema.String),
  audioUs: Schema.Array(Schema.String),
});

export type CharacterStory = typeof CharacterStory.Type;

export const CharacterVoice = Schema.Struct({
  id: Schema.Finite,
  category: Schema.Finite,
  unlockCondition: Schema.Finite,
  unlockQuestId: Schema.Finite,
  sortOrder: Schema.Finite,
  subtitle: Schema.String,
  subtitleKey: Schema.String,
  buttonText: Schema.String,
  buttonTextKey: Schema.String,
  audioJp: Schema.String,
  audioUs: Schema.String,
});

export type CharacterVoice = typeof CharacterVoice.Type;

export const Character = Schema.Struct({
  id: Schema.Finite,
  memo: Schema.String,
  name: Schema.String,
  title: Schema.String,
  nameKey: Schema.String,
  name2Key: Schema.String,
  rarityFlags: Schema.Finite,
  elementType: Schema.Finite,
  jobFlags: Schema.Finite,
  characterType: Schema.Finite,
  startTimeFixJST: Schema.String,
  endTimeFixJST: Schema.String,
  profile: CharacterProfile,
  stats: CharacterStats,
  assets: CharacterAssets,
  skills: CharacterSkills,
  uniqueWeapon: UniqueWeapon,
  stories: Schema.Array(CharacterStory),
  voices: Schema.Array(CharacterVoice),
  iconPath: Schema.String,
});

export type Character = typeof Character.Type;

export const CharacterList = Schema.Array(Character);

export const CharacterListFromJson = Schema.fromJsonString(CharacterList);

export const CharacterFromJson = Schema.fromJsonString(Character);

const DEFAULT_API_ORIGIN = "http://localhost:8080";

const elementLabels: Readonly<Record<number, string>> = {
  1: "Azure",
  2: "Crimson",
  3: "Emerald",
  4: "Amber",
  5: "Radiance",
  6: "Chaos",
};

const jobLabels: Readonly<Record<number, string>> = {
  1: "Warrior",
  2: "Sniper",
  4: "Sorcerer",
};

const bloodTypeLabels: Readonly<Record<number, string>> = {
  1: "A",
  2: "B",
  3: "O",
  4: "AB",
};

export function getApiOrigin(origin?: string): string {
  if (origin !== undefined && origin.length > 0) {
    return origin;
  }

  const envOrigin: unknown = import.meta.env.VITE_API_ORIGIN;
  if (typeof envOrigin === "string" && envOrigin.length > 0) {
    return envOrigin;
  }

  return DEFAULT_API_ORIGIN;
}

function withOrigin(path: string, origin: string): string {
  if (path.length === 0 || /^[a-z][a-z\d+.-]*:/i.test(path)) {
    return path;
  }

  return new URL(path, origin).toString();
}

function withSkillAssetOrigin(skill: CharacterSkill, apiOrigin: string): CharacterSkill {
  return {
    ...skill,
    icon: withOrigin(skill.icon, apiOrigin),
  };
}

export function withAssetOrigin(character: Character, apiOrigin: string): Character {
  return {
    ...character,
    iconPath: withOrigin(character.iconPath, apiOrigin),
    assets: Record.map(character.assets, (path) => withOrigin(path, apiOrigin)),
    skills: {
      normal: withSkillAssetOrigin(character.skills.normal, apiOrigin),
      active: Array.map(character.skills.active, (skill) => withSkillAssetOrigin(skill, apiOrigin)),
      passive: Array.map(character.skills.passive, (skill) =>
        withSkillAssetOrigin(skill, apiOrigin),
      ),
    },
    uniqueWeapon: {
      ...character.uniqueWeapon,
      icon: withOrigin(character.uniqueWeapon.icon, apiOrigin),
    },
    stories: Array.map(character.stories, (story) => ({
      ...story,
      audioJp: Array.map(story.audioJp, (path) => withOrigin(path, apiOrigin)),
      audioUs: Array.map(story.audioUs, (path) => withOrigin(path, apiOrigin)),
    })),
    voices: Array.map(character.voices, (voice) => ({
      ...voice,
      audioJp: withOrigin(voice.audioJp, apiOrigin),
      audioUs: withOrigin(voice.audioUs, apiOrigin),
    })),
  };
}

type CharacterSearchDocument = {
  readonly id: number;
  readonly idText: string;
  readonly name: string;
  readonly title: string;
  readonly element: string;
  readonly job: string;
  readonly memo: string;
  readonly description: string;
  readonly voice: string;
  readonly lament: string;
  readonly weapon: string;
  readonly skills: string;
  readonly stories: string;
  readonly voices: string;
  readonly keys: string;
};

const SEARCH_FIELDS = [
  "idText",
  "name",
  "title",
  "element",
  "job",
  "memo",
  "description",
  "voice",
  "lament",
  "weapon",
  "skills",
  "stories",
  "voices",
  "keys",
];

const SEARCH_BOOST = {
  name: 5,
  title: 3,
  idText: 3,
  element: 2,
  job: 2,
  memo: 2,
  voice: 2,
  lament: 2,
  weapon: 2,
  skills: 2,
  description: 1.5,
  stories: 1,
  voices: 1,
  keys: 0.5,
} as const;

const CJK_PATTERN = /[\p{Script=Han}\p{Script=Hiragana}\p{Script=Katakana}\p{Script=Hangul}]/u;

const searchCorpusCache = new WeakMap<Character, string>();
const searchIndexCache = new WeakMap<
  ReadonlyArray<Character>,
  MiniSearch<CharacterSearchDocument>
>();

function collectSkillSearchParts(skill: CharacterSkill, parts: Array<string>): void {
  if (skill.name.length > 0) parts.push(skill.name);
  if (skill.nameKey.length > 0) parts.push(skill.nameKey);

  for (const level of skill.levels) {
    if (level.description.length > 0) parts.push(level.description);
  }
}

function joinNonEmpty(parts: ReadonlyArray<string>): string {
  return pipe(
    parts,
    Array.filter((part) => part.length > 0),
    Array.join(" "),
  );
}

function toSearchDocument(character: Character, docIndex: number): CharacterSearchDocument {
  const skillParts: Array<string> = [];
  collectSkillSearchParts(character.skills.normal, skillParts);

  for (const skill of character.skills.active) {
    collectSkillSearchParts(skill, skillParts);
  }

  for (const skill of character.skills.passive) {
    collectSkillSearchParts(skill, skillParts);
  }

  const storyParts: Array<string> = [];
  for (const story of character.stories) {
    if (story.title.length > 0) storyParts.push(story.title);
    if (story.text.length > 0) storyParts.push(story.text);
  }

  const voiceParts: Array<string> = [];
  for (const voice of character.voices) {
    if (voice.buttonText.length > 0) voiceParts.push(voice.buttonText);
    if (voice.subtitle.length > 0) voiceParts.push(voice.subtitle);
  }

  return {
    id: docIndex,
    idText: String(character.id),
    name: character.name,
    title: character.title,
    element: formatElement(character.elementType),
    job: formatJob(character.jobFlags),
    memo: character.memo,
    description: character.profile.description,
    voice: joinNonEmpty([
      character.profile.voiceJp,
      character.profile.voiceUs,
      character.profile.vocalJp,
      character.profile.vocalUs,
    ]),
    lament: joinNonEmpty([
      character.profile.lamentJp,
      character.profile.lamentUs,
      character.profile.lyricsJp,
      character.profile.lyricsUs,
      character.profile.gachaMessage,
    ]),
    weapon: joinNonEmpty([
      character.uniqueWeapon.name,
      character.uniqueWeapon.descriptions.ssr,
      character.uniqueWeapon.descriptions.ur,
      character.uniqueWeapon.descriptions.lr,
    ]),
    skills: joinNonEmpty(skillParts),
    stories: joinNonEmpty(storyParts),
    voices: joinNonEmpty(voiceParts),
    keys: joinNonEmpty([character.nameKey, character.name2Key, character.uniqueWeapon.nameKey]),
  };
}

function getSearchIndex(characters: ReadonlyArray<Character>): MiniSearch<CharacterSearchDocument> {
  const cached = searchIndexCache.get(characters);
  if (cached !== undefined) {
    return cached;
  }

  const index = new MiniSearch<CharacterSearchDocument>({
    fields: SEARCH_FIELDS,
  });
  const documents = Array.map(characters, (character, docIndex) =>
    toSearchDocument(character, docIndex),
  );
  index.addAll(documents);
  searchIndexCache.set(characters, index);

  return index;
}

function buildSearchCorpus(character: Character): string {
  const cached = searchCorpusCache.get(character);
  if (cached !== undefined) {
    return cached;
  }

  const doc = toSearchDocument(character, 0);
  const corpus = pipe(
    [
      doc.idText,
      doc.name,
      doc.title,
      doc.element,
      doc.job,
      doc.memo,
      doc.description,
      doc.voice,
      doc.lament,
      doc.weapon,
      doc.skills,
      doc.stories,
      doc.voices,
      doc.keys,
    ],
    Array.filter((part) => part.length > 0),
    Array.join("\n"),
    EffectString.toLowerCase,
  );
  searchCorpusCache.set(character, corpus);

  return corpus;
}

export function filterCharacters(
  characters: ReadonlyArray<Character>,
  query: string,
): ReadonlyArray<Character> {
  const trimmedQuery = EffectString.trim(query);
  if (trimmedQuery.length === 0) {
    return characters;
  }

  if (CJK_PATTERN.test(trimmedQuery)) {
    const normalizedQuery = EffectString.toLowerCase(trimmedQuery);
    return Array.filter(characters, (character) =>
      EffectString.includes(normalizedQuery)(buildSearchCorpus(character)),
    );
  }

  const index = getSearchIndex(characters);
  const results = index.search(trimmedQuery, {
    prefix: true,
    fuzzy: (term) => (term.length >= 4 ? 0.2 : false),
    combineWith: "AND",
    boost: SEARCH_BOOST,
  });

  const matched: Array<Character> = [];
  for (const hit of results) {
    const character = characters[Number(hit.id)];
    if (character !== undefined) {
      matched.push(character);
    }
  }

  return matched;
}

const byIdOrder: Order.Order<Character> = Order.mapInput(Order.Number, (character) => character.id);

const byNameOrder: Order.Order<Character> = Order.combine(
  Order.make<Character>((self, that) =>
    EffectString.localeCompare(displayName(that))(displayName(self)),
  ),
  byIdOrder,
);

const byRarityOrder: Order.Order<Character> = Order.combine(
  Order.flip(Order.mapInput(Order.Number, (character) => character.rarityFlags)),
  byIdOrder,
);

const byElementOrder: Order.Order<Character> = Order.combine(
  Order.mapInput(Order.Number, (character) => character.elementType),
  byIdOrder,
);

const sortOrders: Readonly<Record<SortKey, Order.Order<Character>>> = {
  id: byIdOrder,
  name: byNameOrder,
  rarity: byRarityOrder,
  element: byElementOrder,
};

export function sortCharacters(
  characters: ReadonlyArray<Character>,
  sort: SortKey,
): ReadonlyArray<Character> {
  return Array.sort(characters, sortOrders[sort]);
}

export function formatBirthday(value: number): string {
  if (value === 0) {
    return "";
  }

  const raw = EffectString.padStart(4, "0")(String(value));

  return `${EffectString.slice(0, -2)(raw)}/${EffectString.slice(-2)(raw)}`;
}

export function formatHeight(value: number): string {
  return value !== 0 ? `${value} cm` : "";
}

export function formatWeight(value: number): string {
  return value !== 0 ? `${value} kg` : "";
}

export function formatElement(value: number): string {
  return elementLabels[value] ?? `Element ${value}`;
}

export function formatJob(value: number): string {
  return jobLabels[value] ?? `Class ${value}`;
}

export function formatBloodType(value: number): string {
  return bloodTypeLabels[value] ?? (value !== 0 ? `Type ${value}` : "");
}

export function displayName(character: Pick<Character, "name" | "nameKey">): string {
  return character.name.length > 0 ? character.name : character.nameKey;
}

function safeHttpUrl(value: string | undefined): string {
  if (value === undefined || value.length === 0) {
    return "";
  }

  try {
    const url = new URL(value);

    return url.protocol === "https:" || url.protocol === "http:" ? url.toString() : "";
  } catch {
    return "";
  }
}

export function firstAvailableUrl(
  urls: Readonly<Record<string, string>> | null | undefined,
  preferredKeys: ReadonlyArray<string>,
): string {
  if (urls === null || urls === undefined) {
    return "";
  }

  for (const key of preferredKeys) {
    const value = safeHttpUrl(urls[key]);
    if (value.length > 0) {
      return value;
    }
  }

  for (const value of Record.values(urls)) {
    const safeValue = safeHttpUrl(value);
    if (safeValue.length > 0) {
      return safeValue;
    }
  }

  return "";
}

export function cleanGameText(text: string): string {
  return pipe(
    text,
    EffectString.replace(/<br\s*\/?\s*>/gi, "\n"),
    EffectString.replace(/<\/?(?:size|b|i|color|ruby|rt|rp)[^>]*>/gi, ""),
    EffectString.replace(/<[^>]+>/g, ""),
    EffectString.replace(/\n{3,}/g, "\n\n"),
    EffectString.trim,
  );
}

export function isResourceKey(text: string): boolean {
  return /^\[[\w.:-]+\]$/.test(EffectString.trim(text));
}

export function archiveQuote(character: Character): string {
  const cleanedMessage = cleanGameText(character.profile.gachaMessage);
  if (cleanedMessage.length > 0 && !isResourceKey(cleanedMessage)) {
    return cleanedMessage;
  }

  if (character.id === 1) {
    return "I will do my best\nfor the ones I cherish.";
  }

  const cleanedTitle = cleanGameText(character.title);
  if (cleanedTitle.length > 0 && !isResourceKey(cleanedTitle)) {
    return cleanedTitle;
  }

  const sentences = EffectString.split(/(?<=[.!?])\s+/)(
    cleanGameText(character.profile.description),
  );
  const firstSentence = sentences[0];

  return firstSentence !== undefined && firstSentence.length > 0
    ? firstSentence
    : "Character archive record";
}

export interface AudioSource {
  readonly label: string;
  readonly src: string;
}

export function preferredAutoplayAudio(character: Character): string {
  return Option.getOrElse(
    Array.findFirst(
      [
        character.assets.songShortJp,
        character.assets.songJp,
        character.stories[0]?.audioJp[0] ?? "",
        character.voices[0]?.audioJp ?? "",
      ],
      (candidate) => candidate.length > 0,
    ),
    () => "",
  );
}

export function storyAudioSources(story: CharacterStory): ReadonlyArray<AudioSource> {
  return [
    ...Array.map(story.audioJp, (src, index) => ({ label: `JP segment ${index + 1}`, src })),
    ...Array.map(story.audioUs, (src, index) => ({ label: `US segment ${index + 1}`, src })),
  ];
}

export function voiceAudioSources(voice: CharacterVoice): ReadonlyArray<AudioSource> {
  return Array.filter(
    [
      voice.audioJp.length > 0 ? { label: "JP voice", src: voice.audioJp } : null,
      voice.audioUs.length > 0 ? { label: "US voice", src: voice.audioUs } : null,
    ],
    (source): source is AudioSource => source !== null,
  );
}
