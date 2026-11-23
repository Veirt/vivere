import { Array, Data, Option, Order, Record, Schema } from "effect";

export const CharacterId = Schema.NonEmptyString.pipe(Schema.brand("CharacterId"));

export type CharacterId = typeof CharacterId.Type;

export const CharacterSummary = Schema.Struct({
  id: CharacterId,
  spriteSPath: Schema.NonEmptyString,
});

export type CharacterSummary = typeof CharacterSummary.Type;

export const CharacterListResponse = Schema.Struct({
  characters: Schema.Array(CharacterSummary),
});

export type CharacterListResponse = typeof CharacterListResponse.Type;

export const CharacterSprites = Schema.Struct({
  spriteSPath: Schema.NonEmptyString,
  spriteMPath: Schema.NonEmptyString,
  spriteLPath: Schema.NonEmptyString,
  spriteWPath: Schema.NonEmptyString,
});

export type CharacterSprites = typeof CharacterSprites.Type;

export const CharacterMusic = Schema.Struct({
  fullJPMusicPath: Schema.String,
  fullENMusicPath: Schema.String,
  shortJPMusicPath: Schema.String,
  shortENMusicPath: Schema.String,
});

export type CharacterMusic = typeof CharacterMusic.Type;

export const ModelMotionFile = Schema.Struct({
  File: Schema.NonEmptyString,
});

export type ModelMotionFile = typeof ModelMotionFile.Type;

export const ModelGroup = Schema.Struct({
  Target: Schema.String,
  Name: Schema.String,
  Ids: Schema.Array(Schema.String),
});

export type ModelGroup = typeof ModelGroup.Type;

export const ModelFileReferences = Schema.Struct({
  Moc: Schema.NonEmptyString,
  Textures: Schema.Array(Schema.NonEmptyString),
  Physics: Schema.String,
  Motions: Schema.Record(Schema.String, Schema.Array(ModelMotionFile)),
});

export type ModelFileReferences = typeof ModelFileReferences.Type;

export const ModelDetail = Schema.Struct({
  Version: Schema.Int,
  Name: Schema.String,
  FileReferences: ModelFileReferences,
  Groups: Schema.Array(ModelGroup),
});

export type ModelDetail = typeof ModelDetail.Type;

export const CharacterDetail = Schema.Struct({
  id: CharacterId,
  sprites: CharacterSprites,
  music: CharacterMusic,
  modelPath: Schema.NonEmptyString,
  modelDetail: ModelDetail,
});

export type CharacterDetail = typeof CharacterDetail.Type;

export const CharacterListResponseFromJson = Schema.fromJsonString(CharacterListResponse);

export const CharacterDetailFromJson = Schema.fromJsonString(CharacterDetail);

export interface SpriteEntry {
  readonly key: "small" | "medium" | "large" | "wide";
  readonly label: string;
  readonly src: string;
}

export function selectPrimaryMusicTrack(input: {
  readonly music: CharacterMusic;
}): Option.Option<string> {
  const candidates = [
    input.music.shortJPMusicPath,
    input.music.fullJPMusicPath,
    input.music.shortENMusicPath,
    input.music.fullENMusicPath,
  ];

  return Array.findFirst(candidates, (track) => track.length > 0);
}

export function listMotionGroups(input: {
  readonly detail: CharacterDetail;
}): ReadonlyArray<string> {
  return Array.sort(Record.keys(input.detail.modelDetail.FileReferences.Motions), Order.String);
}

export function listSpriteEntries(input: {
  readonly sprites: CharacterSprites;
}): ReadonlyArray<SpriteEntry> {
  return [
    { key: "small", label: "Small", src: input.sprites.spriteSPath },
    { key: "medium", label: "Medium", src: input.sprites.spriteMPath },
    { key: "large", label: "Large", src: input.sprites.spriteLPath },
    { key: "wide", label: "Wide", src: input.sprites.spriteWPath },
  ];
}

export const DEFAULT_ZOOM_SCALE = 0.1;

export const MIN_ZOOM_SCALE = 0.06;

export const MAX_ZOOM_SCALE = 0.5;

export const ZOOM_STEP = 0.01;

export function stepZoomScale(input: {
  readonly current: number;
  readonly deltaY: number;
}): number {
  if (input.deltaY === 0) return input.current;

  const delta = input.deltaY < 0 ? ZOOM_STEP : -ZOOM_STEP;
  const clamped = Math.min(MAX_ZOOM_SCALE, Math.max(MIN_ZOOM_SCALE, input.current + delta));

  return Math.round(clamped * 100) / 100;
}

export type StageViewState = Data.TaggedEnum<{
  Loading: { readonly modelPath: string };
  Ready: {
    readonly modelPath: string;
    readonly scale: number;
    readonly activeMotion: Option.Option<string>;
  };
  Failed: {
    readonly modelPath: string;
    readonly message: string;
  };
}>;

export const StageViewState = Data.taggedEnum<StageViewState>();
