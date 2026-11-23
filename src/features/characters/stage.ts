import { Context, Effect, Layer, Scope } from "effect";
import { Application, extensions } from "pixi.js";
import { Live2DModel, Live2DPlugin, MotionPriority } from "untitled-pixi-live2d-engine/cubism";

import { Live2DStageError } from "./errors";

export interface Live2DStageSession {
  readonly playMotion: (input: {
    readonly motionGroup: string;
  }) => Effect.Effect<boolean, Live2DStageError>;
  readonly setScale: (input: { readonly scale: number }) => Effect.Effect<void>;
}

let pluginRegistered = false;

function ensureLive2DPlugin(): void {
  if (pluginRegistered) return;

  extensions.add(Live2DPlugin);
  pluginRegistered = true;
}

export class Live2DStage extends Context.Service<
  Live2DStage,
  {
    readonly mount: (input: {
      readonly canvas: HTMLCanvasElement;
      readonly modelPath: string;
      readonly initialScale: number;
    }) => Effect.Effect<Live2DStageSession, Live2DStageError, Scope.Scope>;
  }
>()("vivere/features/characters/stage/Live2DStage", {
  make: Effect.sync(() => ({
    mount: Effect.fn("live2d_stage.mount")(function* (input: {
      readonly canvas: HTMLCanvasElement;
      readonly modelPath: string;
      readonly initialScale: number;
    }) {
      const resources = yield* Effect.acquireRelease(
        Effect.tryPromise({
          try: () => {
            ensureLive2DPlugin();

            const app = new Application();

            return app
              .init({
                canvas: input.canvas,
                backgroundAlpha: 0,
                preference: "webgl",
              })
              .then(() =>
                Live2DModel.from(input.modelPath, {
                  idleMotionGroup: "Idle.anim",
                  autoInteract: false,
                  breathDepth: 0,
                  ticker: app.ticker,
                }),
              )
              .then((model) => {
                model.scale.set(input.initialScale, input.initialScale);
                model.anchor.set(0.5, 0.5);
                model.position.set(input.canvas.width / 2, input.canvas.height / 2);
                app.stage.addChild(model);

                return { app, model };
              });
          },
          catch: (error) =>
            Live2DStageError.make({
              modelPath: input.modelPath,
              reason: error instanceof Error ? error.message : String(error),
            }),
        }).pipe(
          Effect.tapError((error) => Effect.logError("Failed to initialize Live2D stage", error)),
        ),
        ({ app, model }) =>
          Effect.sync(() => {
            app.stage.removeChild(model);
            model.destroy({ children: true, texture: true });
            app.destroy(false, {
              children: true,
              texture: true,
              textureSource: true,
              context: true,
            });
          }).pipe(
            Effect.catchDefect((defect) =>
              Effect.logWarning("Error during Live2D stage teardown", defect),
            ),
          ),
      );

      const playMotion = Effect.fn("live2d_stage.play_motion")(function* (motionInput: {
        readonly motionGroup: string;
      }) {
        return yield* Effect.tryPromise({
          try: () => resources.model.motion(motionInput.motionGroup, 0, MotionPriority.FORCE),
          catch: (error) =>
            Live2DStageError.make({
              modelPath: input.modelPath,
              reason: error instanceof Error ? error.message : String(error),
            }),
        }).pipe(
          Effect.tapError((error) => Effect.logError("Failed to trigger Live2D motion", error)),
        );
      });

      const setScale = (scaleInput: { readonly scale: number }): Effect.Effect<void> =>
        Effect.sync(() => {
          resources.model.scale.set(scaleInput.scale, scaleInput.scale);
        }).pipe(Effect.withSpan("live2d_stage.set_scale"));

      return {
        playMotion,
        setScale,
      };
    }),
  })),
}) {
  static readonly layer = Layer.effect(this, this.make);
}
