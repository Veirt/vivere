import { describe, expect, it } from "@effect/vitest";
import { Effect, Schema } from "effect";

import { VivereApi } from "../src/features/characters/api";
import { CharacterId } from "../src/features/characters/domain";
import { CharacterDecodeError, CharacterNotFoundError } from "../src/features/characters/errors";
import { sampleCharacterDetailJson, sampleCharacterListJson } from "./support/fixtures";
import { makeTestVivereApiLayer } from "./support/layers";

const HappyApiLayer = makeTestVivereApiLayer({
  "/api/characters": Effect.succeed(sampleCharacterListJson),
  "/api/characters/CHR_000085": Effect.succeed(sampleCharacterDetailJson),
  "/api/characters/MISSING": Effect.fail(CharacterNotFoundError.make({ characterId: "MISSING" })),
  "/api/characters/CORRUPT": Effect.succeed("{not valid json"),
});

describe("VivereApi", () => {
  it.layer(HappyApiLayer)((it) => {
    it.effect("lists decoded character summaries", () =>
      Effect.gen(function* () {
        const api = yield* VivereApi;
        const characters = yield* api.listCharacters;

        expect(characters).toHaveLength(1);
        expect(characters[0]?.id).toBe("CHR_000085");
      }),
    );

    it.effect("fetches and decodes character detail", () =>
      Effect.gen(function* () {
        const api = yield* VivereApi;
        const id = yield* Schema.decodeEffect(CharacterId)("CHR_000085");
        const detail = yield* api.getCharacterDetail({ id });

        expect(detail.id).toBe("CHR_000085");
        expect(detail.modelDetail.FileReferences.Motions["Idle.anim"]).toHaveLength(1);
      }),
    );

    it.effect("surfaces CharacterNotFoundError on 404 responses", () =>
      Effect.gen(function* () {
        const api = yield* VivereApi;
        const id = yield* Schema.decodeEffect(CharacterId)("MISSING");
        const error = yield* api.getCharacterDetail({ id }).pipe(Effect.flip);

        expect(Schema.is(CharacterNotFoundError)(error)).toBe(true);
      }),
    );

    it.effect("surfaces CharacterDecodeError on malformed JSON responses", () =>
      Effect.gen(function* () {
        const api = yield* VivereApi;
        const id = yield* Schema.decodeEffect(CharacterId)("CORRUPT");
        const error = yield* api.getCharacterDetail({ id }).pipe(Effect.flip);

        expect(Schema.is(CharacterDecodeError)(error)).toBe(true);
      }),
    );
  });
});
