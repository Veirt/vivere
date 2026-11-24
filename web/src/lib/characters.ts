import { queryOptions } from "@tanstack/react-query";
import { Effect } from "effect";

import { VivereApi } from "../features/characters/api.ts";
import { type Character, type Language } from "../features/characters/domain.ts";
import { runAppPromise } from "../shared/runtime.ts";

export * from "../features/characters/domain.ts";

export const QUERY_STALE_TIME_MS = 5 * 60_000;

export function fetchCharacters(
  language: Language,
  signal?: AbortSignal,
  origin?: string,
): Promise<ReadonlyArray<Character>> {
  return runAppPromise(
    Effect.gen(function* () {
      const api = yield* VivereApi;

      return yield* api.listCharacters({
        language,
        ...(origin !== undefined ? { origin } : {}),
      });
    }),
    signal,
  );
}

export function fetchCharacter(
  characterId: number,
  language: Language,
  signal?: AbortSignal,
  origin?: string,
): Promise<Character | null> {
  return runAppPromise(
    Effect.gen(function* () {
      const api = yield* VivereApi;

      return yield* api
        .getCharacter({
          id: characterId,
          language,
          ...(origin !== undefined ? { origin } : {}),
        })
        .pipe(Effect.catchTag("CharacterNotFoundError", () => Effect.succeed(null)));
    }),
    signal,
  );
}

export function charactersQueryOptions(language: Language) {
  return queryOptions({
    queryKey: ["characters", language] as const,
    queryFn: ({ signal }) => fetchCharacters(language, signal),
    staleTime: QUERY_STALE_TIME_MS,
  });
}

export function characterQueryOptions(characterId: number, language: Language) {
  return queryOptions({
    queryKey: ["character", characterId, language] as const,
    queryFn: ({ signal }) => fetchCharacter(characterId, language, signal),
    staleTime: QUERY_STALE_TIME_MS,
  });
}
