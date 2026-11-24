import { Array, Context, Effect, Layer, Schema, String as EffectString } from "effect";
import { HttpClient, type HttpClientError } from "effect/http";

import {
  type Character,
  CharacterApiError,
  CharacterDecodeError,
  type CharacterFetchError,
  CharacterFromJson,
  CharacterListFromJson,
  CharacterNotFoundError,
  type Language,
  getApiOrigin,
  withAssetOrigin,
} from "./domain.ts";

function toApiError(endpoint: string, error: HttpClientError.HttpClientError): CharacterApiError {
  return CharacterApiError.make({
    endpoint,
    reason: error.message,
    ...(error.response !== undefined ? { status: error.response.status } : {}),
  });
}

export class VivereApi extends Context.Service<
  VivereApi,
  {
    readonly listCharacters: (input: {
      readonly language: Language;
      readonly origin?: string;
    }) => Effect.Effect<ReadonlyArray<Character>, CharacterApiError | CharacterDecodeError>;
    readonly getCharacter: (input: {
      readonly id: number;
      readonly language: Language;
      readonly origin?: string;
    }) => Effect.Effect<Character, CharacterFetchError>;
  }
>()("@vivere/web/features/characters/api/VivereApi", {
  make: Effect.gen(function* () {
    const httpClient = yield* HttpClient.HttpClient;

    const getText = Effect.fn("vivere_api.get_text")(function* (endpoint: string) {
      const response = yield* httpClient
        .get(endpoint)
        .pipe(Effect.mapError((error) => toApiError(endpoint, error)));

      if (response.status === 404) {
        return yield* CharacterNotFoundError.make({ characterId: endpoint });
      }

      if (response.status < 200 || response.status >= 300) {
        const body = yield* response.text.pipe(Effect.orElseSucceed(() => ""));
        const trimmedBody = EffectString.trim(body);
        const reason =
          trimmedBody.length > 0 ? trimmedBody : EffectString.trim(`HTTP ${response.status}`);

        return yield* CharacterApiError.make({
          endpoint,
          reason,
          status: response.status,
        });
      }

      return yield* response.text.pipe(Effect.mapError((error) => toApiError(endpoint, error)));
    });

    const listCharacters = Effect.fn("vivere_api.list_characters")(function* (input: {
      readonly language: Language;
      readonly origin?: string;
    }) {
      const apiOrigin = getApiOrigin(input.origin);
      const endpoint = new URL(
        `/api/characters?lang=${encodeURIComponent(input.language)}`,
        apiOrigin,
      ).toString();

      const raw = yield* getText(endpoint).pipe(
        Effect.catchTag("CharacterNotFoundError", (error) =>
          Effect.fail(
            CharacterApiError.make({
              endpoint,
              reason: `Characters endpoint not found: ${error.characterId}`,
              status: 404,
            }),
          ),
        ),
        Effect.tapError((error) => Effect.logError("HTTP transport request failed", error)),
      );

      const characters = yield* Schema.decodeEffect(CharacterListFromJson)(raw).pipe(
        Effect.tapError((error) => Effect.logError("Character list decode failed", error)),
        Effect.mapError((error) =>
          CharacterDecodeError.make({
            endpoint,
            reason: error.message,
          }),
        ),
      );

      return Array.map(characters, (character) => withAssetOrigin(character, apiOrigin));
    });

    const getCharacter = Effect.fn("vivere_api.get_character")(function* (input: {
      readonly id: number;
      readonly language: Language;
      readonly origin?: string;
    }) {
      if (!Number.isInteger(input.id)) {
        return yield* CharacterNotFoundError.make({ characterId: String(input.id) });
      }

      const apiOrigin = getApiOrigin(input.origin);
      const endpoint = new URL(
        `/api/characters/${input.id}?lang=${encodeURIComponent(input.language)}`,
        apiOrigin,
      ).toString();

      const raw = yield* getText(endpoint).pipe(
        Effect.catchTag("CharacterNotFoundError", () =>
          Effect.fail(CharacterNotFoundError.make({ characterId: String(input.id) })),
        ),
        Effect.tapError((error) => Effect.logError("HTTP transport request failed", error)),
      );

      const character = yield* Schema.decodeEffect(CharacterFromJson)(raw).pipe(
        Effect.tapError((error) => Effect.logError("Character detail decode failed", error)),
        Effect.mapError((error) =>
          CharacterDecodeError.make({
            endpoint,
            reason: error.message,
          }),
        ),
      );

      return withAssetOrigin(character, apiOrigin);
    });

    return {
      listCharacters,
      getCharacter,
    };
  }),
}) {
  static readonly layer = Layer.effect(this, this.make);
}
