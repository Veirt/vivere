import { Schema } from "effect";

export class CharacterNotFoundError extends Schema.TaggedError<
  CharacterNotFoundError,
  { readonly brand: unique symbol }
>()("CharacterNotFoundError", {
  characterId: Schema.String,
}) {}

export class CharacterTransportError extends Schema.TaggedError<
  CharacterTransportError,
  { readonly brand: unique symbol }
>()("CharacterTransportError", {
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

export class Live2DStageError extends Schema.TaggedError<
  Live2DStageError,
  { readonly brand: unique symbol }
>()("Live2DStageError", {
  modelPath: Schema.String,
  reason: Schema.String,
}) {}

export type CharacterApiError =
  | CharacterNotFoundError
  | CharacterTransportError
  | CharacterDecodeError;
