import { Effect, Exit, Layer, ManagedRuntime, Scope } from "effect";

import { HttpTransport, VivereApi } from "../features/characters/api";
import { Live2DStage, type Live2DStageSession } from "../features/characters/stage";
import { Live2DStageError } from "../features/characters/errors";

export type AppServices = VivereApi | Live2DStage;

const AppLayer = Layer.mergeAll(
  VivereApi.layer.pipe(Layer.provide(HttpTransport.layer)),
  Live2DStage.layer,
);

const appRuntime = ManagedRuntime.make(AppLayer);

export function runAppPromise<A, E>(
  effect: Effect.Effect<A, E, AppServices>,
  signal?: AbortSignal,
): Promise<A> {
  return appRuntime.runPromise(effect, signal === undefined ? undefined : { signal });
}

export function mountStageWithScope(input: {
  readonly canvas: HTMLCanvasElement;
  readonly modelPath: string;
  readonly initialScale: number;
  readonly onReady: (session: Live2DStageSession) => void;
  readonly onError: (error: Live2DStageError) => void;
}): () => void {
  const scope = Scope.makeUnsafe();
  const controller = new AbortController();

  const program = Effect.gen(function* () {
    const stage = yield* Live2DStage;

    return yield* stage
      .mount({
        canvas: input.canvas,
        modelPath: input.modelPath,
        initialScale: input.initialScale,
      })
      .pipe(Scope.provide(scope));
  });

  void appRuntime.runPromiseExit(program, { signal: controller.signal }).then((exit) => {
    if (controller.signal.aborted) return;

    if (Exit.isSuccess(exit)) {
      input.onReady(exit.value);

      return;
    }

    input.onError(
      Live2DStageError.make({
        modelPath: input.modelPath,
        reason: String(exit.cause),
      }),
    );
  });

  return () => {
    controller.abort();
    void appRuntime.runPromise(Scope.close(scope, Exit.void));
  };
}
