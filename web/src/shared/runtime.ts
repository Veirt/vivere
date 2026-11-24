import { Effect, Layer, ManagedRuntime, Schedule } from "effect";
import { FetchHttpClient, HttpClient } from "effect/http";

import { VivereApi } from "../features/characters/api.ts";

export type AppServices = VivereApi;

const ResilientHttpClientLive = Layer.effect(
  HttpClient.HttpClient,
  Effect.map(
    HttpClient.HttpClient,
    HttpClient.retryTransient({
      times: 2,
      schedule: Schedule.exponential("150 millis"),
    }),
  ),
).pipe(
  Layer.provide(FetchHttpClient.layer),
  Layer.provide(Layer.succeed(HttpClient.TracerPropagationEnabled, false)),
);

export const AppLive = VivereApi.layer.pipe(Layer.provide(ResilientHttpClientLive));

export const appRuntime = ManagedRuntime.make(AppLive);

export function runAppPromise<A, E>(
  effect: Effect.Effect<A, E, AppServices>,
  signal?: AbortSignal,
): Promise<A> {
  return appRuntime.runPromise(effect, signal === undefined ? undefined : { signal });
}
