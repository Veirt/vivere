import { describe, expect, it } from "@effect/vitest";
import { Array, Effect, Match, Option, Schema } from "effect";

import {
  CharacterDetailFromJson,
  CharacterListResponseFromJson,
  DEFAULT_ZOOM_SCALE,
  MAX_ZOOM_SCALE,
  MIN_ZOOM_SCALE,
  StageViewState,
  listMotionGroups,
  listSpriteEntries,
  selectPrimaryMusicTrack,
  stepZoomScale,
} from "../src/features/characters/domain";
import { sampleCharacterDetailJson, sampleCharacterListJson } from "./support/fixtures";

describe("characters domain", () => {
  it.effect(
    "decodes real CHR_000085 wire payloads including multi-entry motion arrays and group ids",
    () =>
      Effect.gen(function* () {
        const list = yield* Schema.decodeEffect(CharacterListResponseFromJson)(
          sampleCharacterListJson,
        );

        expect(list.characters).toHaveLength(1);
        expect(list.characters[0]?.id).toBe("CHR_000085");

        const detail =
          yield* Schema.decodeEffect(CharacterDetailFromJson)(sampleCharacterDetailJson);

        expect(detail.modelDetail.FileReferences.Motions["Skill.anim"]).toHaveLength(2);
        expect(detail.modelDetail.Groups[0]?.Ids).toEqual([
          "ParamEyeLOpen",
          "ParamEyeOpenUpaL",
          "ParamEyeROpen",
          "ParamEyeOpenUpaR",
        ]);

        expect(listMotionGroups({ detail })).toEqual([
          "Face_Smile.anim",
          "Idle.anim",
          "Skill.anim",
        ]);

        expect(selectPrimaryMusicTrack({ music: detail.music })).toEqual(
          Option.some("/api/assets/AudioClip/CHR_000085_SONG_SHORT_JP.wav"),
        );

        expect(listSpriteEntries({ sprites: detail.sprites })).toHaveLength(4);
      }),
  );

  it.effect("rejects malformed character detail JSON", () =>
    Effect.gen(function* () {
      const error = yield* Schema.decodeEffect(CharacterDetailFromJson)('{"id":"CHR_000085"}').pipe(
        Effect.flip,
      );

      expect(error).toBeInstanceOf(Schema.SchemaError);
    }),
  );

  it.effect("dispatches StageViewState exhaustively with Match", () =>
    Effect.sync(() => {
      const states = [
        StageViewState.Loading({ modelPath: "/model.json" }),
        StageViewState.Ready({
          modelPath: "/model.json",
          scale: DEFAULT_ZOOM_SCALE,
          activeMotion: Option.some("Idle.anim"),
        }),
        StageViewState.Failed({ modelPath: "/model.json", message: "WebGL unavailable" }),
      ];

      const labels = Array.map(states, (state) =>
        Match.value(state).pipe(
          Match.tagsExhaustive({
            Loading: () => "loading",
            Ready: (ready) => `ready:${ready.scale}`,
            Failed: (failed) => `failed:${failed.message}`,
          }),
        ),
      );

      expect(labels).toEqual(["loading", "ready:0.1", "failed:WebGL unavailable"]);
    }),
  );

  it.effect.prop(
    "keeps stepZoomScale bounded within [MIN_ZOOM_SCALE, MAX_ZOOM_SCALE]",
    [
      Schema.Finite.check(Schema.isBetween({ minimum: MIN_ZOOM_SCALE, maximum: MAX_ZOOM_SCALE })),
      Schema.Finite.check(Schema.isBetween({ minimum: -500, maximum: 500 })),
    ],
    ([current, deltaY]) =>
      Effect.sync(() => {
        const next = stepZoomScale({ current, deltaY });

        expect(next).toBeGreaterThanOrEqual(MIN_ZOOM_SCALE);
        expect(next).toBeLessThanOrEqual(MAX_ZOOM_SCALE);

        if (deltaY < 0) {
          expect(next).toBeGreaterThanOrEqual(current);
        } else if (deltaY > 0) {
          expect(next).toBeLessThanOrEqual(current);
        } else {
          expect(next).toBe(current);
        }
      }),
  );
});
