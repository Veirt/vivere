import { it as effectIt } from "@effect/vitest";
import { Effect, Layer, Schema } from "effect";
import { HttpClient, HttpClientError, HttpClientResponse } from "effect/http";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vite-plus/test";

import { VivereApi } from "#/features/characters/api.ts";
import {
  CharacterApiError,
  CharacterDecodeError,
  CharacterFromJson,
  CharacterListFromJson,
  CharacterNotFoundError,
} from "#/features/characters/domain.ts";

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
  validateCharacterDetailSearch,
  validateCharacterIndexSearch,
  type Character,
  type CharacterAssets,
  type CharacterProfile,
  type CharacterSkill,
  type CharacterSkills,
  type CharacterStats,
  type UniqueWeapon,
} from "./characters.ts";

function spyOnFetch() {
  return vi.spyOn(globalThis, "fetch");
}

describe("character helpers", () => {
  // FetchHttpClient's Context.Reference caches `globalThis.fetch` on first access,
  // so keep a single spy installed across the suite and reset its mock queue per test.
  let fetchSpy: ReturnType<typeof spyOnFetch> | undefined;

  beforeAll(() => {
    fetchSpy = spyOnFetch();
  });

  afterEach(() => {
    fetchSpy?.mockReset();
  });

  afterAll(() => {
    vi.restoreAllMocks();
  });
  const characters: Array<Character> = [
    character({
      id: 2,
      name: "Beta",
      rarityFlags: 3,
      elementType: 1,
      profile: { description: "moon singer" },
    }),
    character({ id: 1, name: "Alpha", rarityFlags: 5, elementType: 2, memo: "healer" }),
    character({ id: 3, name: "Gamma", title: "Dawn Blade", rarityFlags: 5, elementType: 1 }),
  ];

  it("filters across display fields, profile data, skills, weapons, stories, and voices", () => {
    const withRichData = character({
      id: 4,
      name: "Delta",
      skills: { active: [{ ...skill(), id: 42, name: "Petal Storm" }] },
      uniqueWeapon: { name: "Silver Harp" },
      stories: [
        {
          id: 1,
          episodeId: 1,
          level: 0,
          rarityFlags: 1,
          title: "Hidden memory",
          text: "silver rain",
          titleKey: "story.title",
          textKey: "story.text",
          rewardItems: [],
          audioJp: [],
          audioUs: [],
        },
      ],
      voices: [
        {
          id: 1,
          category: 1,
          unlockCondition: 0,
          unlockQuestId: 0,
          sortOrder: 1,
          subtitle: "A quiet line",
          subtitleKey: "voice.subtitle",
          buttonText: "Greeting",
          buttonTextKey: "voice.button",
          audioJp: "",
          audioUs: "",
        },
      ],
    });
    const allCharacters = [...characters, withRichData];

    expect(filterCharacters(allCharacters, "moon")).toEqual([characters[0]]);
    expect(filterCharacters(allCharacters, "healer")).toEqual([characters[1]]);
    expect(filterCharacters(allCharacters, "Dawn")).toEqual([characters[2]]);
    expect(filterCharacters(allCharacters, "Petal Storm")).toEqual([withRichData]);
    expect(filterCharacters(allCharacters, "Silver Harp")).toEqual([withRichData]);
    expect(filterCharacters(allCharacters, "silver rain")).toEqual([withRichData]);
    expect(filterCharacters(allCharacters, "quiet line")).toEqual([withRichData]);
    expect(filterCharacters(allCharacters, "Petl Storm")).toEqual([withRichData]);
    expect(filterCharacters(allCharacters, "crimson healer")).toEqual([characters[1]]);
    expect(filterCharacters(allCharacters, "descriptionKey")).toEqual([]);
    expect(filterCharacters(allCharacters, "audioJp")).toEqual([]);

    const cjkCharacter = character({ id: 5, name: "アイリス", title: "蒼の魔女" });
    expect(filterCharacters([...allCharacters, cjkCharacter], "リス")).toEqual([cjkCharacter]);
  });

  it("validates router search params via Effect Standard Schema with safe fallbacks", () => {
    const validIndex = validateCharacterIndexSearch["~standard"].validate({
      lang: "TextResourceJaJpMB",
      q: "iris",
      sort: "rarity",
    });
    expect(validIndex).toEqual({
      value: {
        lang: "TextResourceJaJpMB",
        q: "iris",
        sort: "rarity",
      },
    });

    const missingIndex = validateCharacterIndexSearch["~standard"].validate({});
    expect(missingIndex).toEqual({
      value: {
        lang: "TextResourceEnUsMB",
        q: "",
        sort: "id",
      },
    });

    const invalidIndex = validateCharacterIndexSearch["~standard"].validate({
      lang: "InvalidLang",
      q: 123,
      sort: "unknown",
    });
    expect(invalidIndex).toEqual({
      value: {
        lang: "TextResourceEnUsMB",
        q: "",
        sort: "id",
      },
    });

    const invalidDetail = validateCharacterDetailSearch["~standard"].validate({ lang: 99 });
    expect(invalidDetail).toEqual({
      value: {
        lang: "TextResourceEnUsMB",
      },
    });
  });

  it("sorts by prototype-compatible keys", () => {
    expect(sortCharacters(characters, "id").map((character) => character.id)).toEqual([1, 2, 3]);
    expect(sortCharacters(characters, "name").map((character) => character.name)).toEqual([
      "Alpha",
      "Beta",
      "Gamma",
    ]);
    expect(sortCharacters(characters, "rarity").map((character) => character.id)).toEqual([
      1, 3, 2,
    ]);
    expect(sortCharacters(characters, "element").map((character) => character.id)).toEqual([
      2, 3, 1,
    ]);
    expect(
      sortCharacters(
        [character({ id: 4, nameKey: "Zeta" }), character({ id: 5, name: "Alpha" })],
        "name",
      ).map((character) => character.id),
    ).toEqual([5, 4]);
  });

  it("formats compact profile values", () => {
    expect(formatBirthday(101)).toBe("01/01");
    expect(formatBirthday(1225)).toBe("12/25");
    expect(formatBirthday(0)).toBe("");
    expect(formatHeight(162)).toBe("162 cm");
    expect(formatWeight(45)).toBe("45 kg");
    expect(formatBloodType(4)).toBe("AB");
    expect(formatBloodType(0)).toBe("");
  });

  it("formats enum-like character metadata with safe fallbacks", () => {
    expect(formatElement(1)).toBe("Azure");
    expect(formatElement(99)).toBe("Element 99");
    expect(formatJob(2)).toBe("Sniper");
    expect(formatJob(99)).toBe("Class 99");
  });

  it("selects the first safe available regional URL by preference", () => {
    expect(
      firstAvailableUrl({ jaJP: "https://example.test/jp", enUS: "https://example.test/us" }, [
        "enUS",
        "jaJP",
      ]),
    ).toBe("https://example.test/us");
    expect(firstAvailableUrl({ jaJP: "https://example.test/jp" }, ["enUS", "jaJP"])).toBe(
      "https://example.test/jp",
    );
    expect(
      firstAvailableUrl({ enUS: "javascript:alert(1)", jaJP: "notaurl" }, ["enUS", "jaJP"]),
    ).toBe("");
    expect(firstAvailableUrl(null, ["enUS"])).toBe("");
  });

  effectIt.effect(
    "fetches one character without downloading the full catalog and rewrites asset paths",
    () =>
      Effect.gen(function* () {
        const payload = yield* Schema.encodeEffect(CharacterFromJson)(
          character({
            id: 17,
            iconPath: "/assets/icon.png",
            assets: { iconLarge: "/assets/large.png" },
          }),
        );
        fetchSpy?.mockResolvedValueOnce(
          new Response(payload, {
            status: 200,
            headers: { "Content-Type": "application/json" },
          }),
        );

        const result = yield* Effect.promise(() =>
          fetchCharacter(17, "TextResourceEnUsMB", undefined, "https://api.example.test"),
        );

        expect(fetchSpy).toHaveBeenCalledWith(
          new URL("/api/characters/17?lang=TextResourceEnUsMB", "https://api.example.test"),
          expect.objectContaining({ method: "GET", signal: expect.any(AbortSignal) }),
        );
        expect(result?.iconPath).toBe("https://api.example.test/assets/icon.png");
        expect(result?.assets.iconLarge).toBe("https://api.example.test/assets/large.png");
      }),
  );

  effectIt.effect("rewrites nested skill, weapon, story, and voiceline asset paths", () =>
    Effect.gen(function* () {
      const payload = yield* Schema.encodeEffect(CharacterFromJson)(
        character({
          id: 18,
          skills: { active: [{ ...skill(), id: 42, icon: "/assets/skill.png" }] },
          uniqueWeapon: { icon: "/assets/weapon.png" },
          stories: [
            {
              id: 1,
              episodeId: 1,
              level: 0,
              rarityFlags: 1,
              title: "Memory",
              text: "",
              titleKey: "",
              textKey: "",
              rewardItems: [],
              audioJp: ["/assets/memory.wav"],
              audioUs: [],
            },
          ],
          voices: [
            {
              id: 1,
              category: 1,
              unlockCondition: 0,
              unlockQuestId: 0,
              sortOrder: 2,
              subtitle: "Hi",
              subtitleKey: "",
              buttonText: "Greeting",
              buttonTextKey: "",
              audioJp: "/assets/voice.wav",
              audioUs: "",
            },
          ],
        }),
      );
      fetchSpy?.mockResolvedValueOnce(
        new Response(payload, { status: 200, headers: { "Content-Type": "application/json" } }),
      );

      const result = yield* Effect.promise(() =>
        fetchCharacter(18, "TextResourceEnUsMB", undefined, "https://api.example.test"),
      );

      expect(result?.skills.active[0]?.icon).toBe("https://api.example.test/assets/skill.png");
      expect(result?.uniqueWeapon.icon).toBe("https://api.example.test/assets/weapon.png");
      expect(result?.stories[0]?.audioJp[0]).toBe("https://api.example.test/assets/memory.wav");
      expect(result?.voices[0]?.audioJp).toBe("https://api.example.test/assets/voice.wav");
    }),
  );

  effectIt.effect("returns null for missing or invalid character ids", () =>
    Effect.gen(function* () {
      fetchSpy?.mockResolvedValueOnce(new Response('{"error":"missing"}', { status: 404 }));

      const invalidResult = yield* Effect.promise(() =>
        fetchCharacter(Number.NaN, "TextResourceEnUsMB", undefined, "https://api.example.test"),
      );
      const missingResult = yield* Effect.promise(() =>
        fetchCharacter(999, "TextResourceEnUsMB", undefined, "https://api.example.test"),
      );

      expect(invalidResult).toBeNull();
      expect(missingResult).toBeNull();
      expect(fetchSpy).toHaveBeenCalledTimes(1);
    }),
  );
});

function makeTestApiLayer(
  responses: Readonly<Record<string, Effect.Effect<Response, HttpClientError.HttpClientError>>>,
) {
  const testHttpClient = Layer.succeed(
    HttpClient.HttpClient,
    HttpClient.make((request, url) =>
      Effect.map(
        responses[url.toString()] ??
          Effect.succeed(new Response('{"error":"not found"}', { status: 404 })),
        (webResponse) => HttpClientResponse.fromWeb(request, webResponse),
      ),
    ),
  );

  return VivereApi.layer.pipe(Layer.provide(testHttpClient));
}

const sampleApiLayer = makeTestApiLayer({
  "https://api.example.test/api/characters?lang=TextResourceEnUsMB": Schema.encodeEffect(
    CharacterListFromJson,
  )([character({ id: 17, name: "Iris", iconPath: "/assets/icon.png" })]).pipe(
    Effect.map(
      (payload) =>
        new Response(payload, { status: 200, headers: { "Content-Type": "application/json" } }),
    ),
    Effect.orDie,
  ),
  "https://api.example.test/api/characters/17?lang=TextResourceEnUsMB": Schema.encodeEffect(
    CharacterFromJson,
  )(character({ id: 17, name: "Iris", iconPath: "/assets/icon.png" })).pipe(
    Effect.map(
      (payload) =>
        new Response(payload, { status: 200, headers: { "Content-Type": "application/json" } }),
    ),
    Effect.orDie,
  ),
  "https://api.example.test/api/characters/404?lang=TextResourceEnUsMB": Effect.succeed(
    new Response('{"error":"missing"}', { status: 404 }),
  ),
  "https://api.example.test/api/characters/500?lang=TextResourceEnUsMB": Effect.succeed(
    new Response("server failure", { status: 500 }),
  ),
  "https://api.example.test/api/characters/999?lang=TextResourceEnUsMB": Effect.succeed(
    new Response("{bad json", { status: 200, headers: { "Content-Type": "application/json" } }),
  ),
});

describe("VivereApi", () => {
  effectIt.layer(sampleApiLayer)((it) => {
    it.effect("lists and validates character catalogs with rewritten asset origins", () =>
      Effect.gen(function* () {
        const api = yield* VivereApi;
        const list = yield* api.listCharacters({
          language: "TextResourceEnUsMB",
          origin: "https://api.example.test",
        });

        expect(list).toHaveLength(1);
        expect(list[0]?.name).toBe("Iris");
        expect(list[0]?.iconPath).toBe("https://api.example.test/assets/icon.png");
      }),
    );

    it.effect("fetches and validates single character detail", () =>
      Effect.gen(function* () {
        const api = yield* VivereApi;
        const detail = yield* api.getCharacter({
          id: 17,
          language: "TextResourceEnUsMB",
          origin: "https://api.example.test",
        });

        expect(detail.id).toBe(17);
        expect(detail.iconPath).toBe("https://api.example.test/assets/icon.png");
      }),
    );

    it.effect("surfaces CharacterNotFoundError on 404 responses", () =>
      Effect.gen(function* () {
        const api = yield* VivereApi;
        const error = yield* api
          .getCharacter({
            id: 404,
            language: "TextResourceEnUsMB",
            origin: "https://api.example.test",
          })
          .pipe(Effect.flip);

        expect(Schema.is(CharacterNotFoundError)(error)).toBe(true);
      }),
    );

    it.effect("surfaces CharacterApiError on 500 responses", () =>
      Effect.gen(function* () {
        const api = yield* VivereApi;
        const error = yield* api
          .getCharacter({
            id: 500,
            language: "TextResourceEnUsMB",
            origin: "https://api.example.test",
          })
          .pipe(Effect.flip);

        expect(Schema.is(CharacterApiError)(error)).toBe(true);
      }),
    );

    it.effect("surfaces CharacterDecodeError on malformed JSON payloads", () =>
      Effect.gen(function* () {
        const api = yield* VivereApi;
        const error = yield* api
          .getCharacter({
            id: 999,
            language: "TextResourceEnUsMB",
            origin: "https://api.example.test",
          })
          .pipe(Effect.flip);

        expect(Schema.is(CharacterDecodeError)(error)).toBe(true);
      }),
    );
  });
});

type CharacterFixtureOverrides = Partial<
  Omit<Character, "profile" | "stats" | "assets" | "skills" | "uniqueWeapon">
> & {
  profile?: Partial<CharacterProfile>;
  stats?: Partial<CharacterStats>;
  assets?: Partial<CharacterAssets>;
  skills?: Partial<CharacterSkills>;
  uniqueWeapon?: Partial<UniqueWeapon>;
};

function character(overrides: CharacterFixtureOverrides): Character {
  return {
    id: 0,
    memo: "",
    name: "",
    title: "",
    nameKey: "",
    name2Key: "",
    rarityFlags: 0,
    elementType: 0,
    jobFlags: 0,
    characterType: 0,
    startTimeFixJST: "",
    endTimeFixJST: "",
    iconPath: "",
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
      name: "",
      nameKey: "",
      iconId: 0,
      icon: "",
      jobFlags: 0,
      rarityFlags: 0,
      level: 0,
      slotType: 0,
      compositeId: 0,
      exclusiveEffectId: 0,
      skillDescriptionId: 0,
      battleParameter: {},
      bonusParameters: [],
      descriptions: { ssr: "", ur: "", lr: "" },
      ...overrides.uniqueWeapon,
    },
    profile: {
      birthday: 0,
      height: 0,
      weight: 0,
      bloodType: 0,
      description: "",
      voiceJp: "",
      voiceUs: "",
      vocalJp: "",
      vocalUs: "",
      lamentJp: "",
      lamentUs: "",
      lyricsJp: "",
      lyricsUs: "",
      gachaMessage: "",
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
      iconSmall: "",
      iconMedium: "",
      iconLarge: "",
      iconWide: "",
      iconEmotion: "",
      elementIcon: "",
      jobIcon: "",
      modelTexture: "",
      effectTexture: "",
      songJp: "",
      songUs: "",
      songShortJp: "",
      songShortUs: "",
      ...overrides.assets,
    },
  };
}

function skill(): CharacterSkill {
  return {
    id: 0,
    type: "active",
    name: "",
    nameKey: "",
    icon: "",
    conditions: "",
    initCooltime: 0,
    maxCooltime: 0,
    levels: [],
  };
}
