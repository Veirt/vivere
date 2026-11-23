import { Context, Effect, Layer, Schema, String as EffectString } from "effect";

import {
  type CharacterDetail,
  CharacterDetailFromJson,
  type CharacterId,
  CharacterListResponseFromJson,
  type CharacterSummary,
} from "./domain";
import {
  type CharacterApiError,
  CharacterDecodeError,
  CharacterNotFoundError,
  CharacterTransportError,
} from "./errors";

export class HttpTransport extends Context.Service<
  HttpTransport,
  {
    readonly getText: (input: {
      readonly endpoint: string;
    }) => Effect.Effect<string, CharacterNotFoundError | CharacterTransportError>;
  }
>()("vivere/features/characters/api/HttpTransport", {
  make: Effect.sync(() => ({
    getText: Effect.fn("http_transport.get_text")(function* (input: { readonly endpoint: string }) {
      return yield* Effect.tryPromise({
        try: (signal) =>
          globalThis.fetch(input.endpoint, { signal }).then((response) => {
            if (response.status === 404) {
              return Promise.reject(CharacterNotFoundError.make({ characterId: input.endpoint }));
            }

            if (!response.ok) {
              return Promise.reject(
                CharacterTransportError.make({
                  endpoint: input.endpoint,
                  reason: EffectString.trim(`HTTP ${response.status} ${response.statusText}`),
                  status: response.status,
                }),
              );
            }

            return response.text();
          }),
        catch: (error) => {
          if (Schema.is(CharacterNotFoundError)(error)) return error;

          if (Schema.is(CharacterTransportError)(error)) return error;

          return CharacterTransportError.make({
            endpoint: input.endpoint,
            reason: error instanceof Error ? error.message : String(error),
          });
        },
      }).pipe(Effect.tapError((error) => Effect.logError("HTTP transport request failed", error)));
    }),
  })),
}) {
  static readonly layer = Layer.effect(this, this.make);
}

export class VivereApi extends Context.Service<
  VivereApi,
  {
    readonly listCharacters: Effect.Effect<
      ReadonlyArray<CharacterSummary>,
      CharacterTransportError | CharacterDecodeError
    >;
    readonly getCharacterDetail: (input: {
      readonly id: CharacterId;
    }) => Effect.Effect<CharacterDetail, CharacterApiError>;
  }
>()("vivere/features/characters/api/VivereApi", {
  make: Effect.gen(function* () {
    const transport = yield* HttpTransport;

    const listCharacters = Effect.gen(function* () {
      const endpoint = "/api/characters";

      const raw = yield* transport.getText({ endpoint }).pipe(
        Effect.catchTag("CharacterNotFoundError", (error) =>
          Effect.fail(
            CharacterTransportError.make({
              endpoint,
              reason: `Characters endpoint not found: ${error.characterId}`,
              status: 404,
            }),
          ),
        ),
      );

      const parsed = yield* Schema.decodeEffect(CharacterListResponseFromJson)(raw).pipe(
        Effect.tapError((error) => Effect.logError("Character list decode failed", error)),
        Effect.mapError((error) =>
          CharacterDecodeError.make({
            endpoint,
            reason: error.message,
          }),
        ),
      );

      return parsed.characters;
    }).pipe(Effect.withSpan("vivere_api.list_characters"));

    const getCharacterDetail = Effect.fn("vivere_api.get_character_detail")(function* (input: {
      readonly id: CharacterId;
    }) {
      const endpoint = `/api/characters/${encodeURIComponent(input.id)}`;

      const raw = yield* transport
        .getText({ endpoint })
        .pipe(
          Effect.catchTag("CharacterNotFoundError", () =>
            Effect.fail(CharacterNotFoundError.make({ characterId: input.id })),
          ),
        );

      return yield* Schema.decodeEffect(CharacterDetailFromJson)(raw).pipe(
        Effect.tapError((error) => Effect.logError("Character detail decode failed", error)),
        Effect.mapError((error) =>
          CharacterDecodeError.make({
            endpoint,
            reason: error.message,
          }),
        ),
      );
    });

    return {
      listCharacters,
      getCharacterDetail,
    };
  }),
}) {
  static readonly layer = Layer.effect(this, this.make);
}
