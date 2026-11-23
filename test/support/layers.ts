import { Effect, Layer, Option, Record } from "effect";

import { HttpTransport, VivereApi } from "../../src/features/characters/api";
import type {
  CharacterNotFoundError,
  CharacterTransportError,
} from "../../src/features/characters/errors";

export type ScriptedTransportResponses = Readonly<
  Record<string, Effect.Effect<string, CharacterNotFoundError | CharacterTransportError>>
>;

export function makeTestHttpTransportLayer(
  responses: ScriptedTransportResponses,
): Layer.Layer<HttpTransport> {
  return Layer.succeed(
    HttpTransport,
    HttpTransport.of({
      getText: ({ endpoint }) =>
        Option.getOrElse(Record.get(responses, endpoint), () =>
          Effect.die(`Unexpected HTTP endpoint requested in test: ${endpoint}`),
        ),
    }),
  );
}

export function makeTestVivereApiLayer(
  responses: ScriptedTransportResponses,
): Layer.Layer<VivereApi> {
  return VivereApi.layer.pipe(Layer.provide(makeTestHttpTransportLayer(responses)));
}
