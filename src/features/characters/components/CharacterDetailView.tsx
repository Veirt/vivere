import { Array, Match, Option } from "effect";
import { useEffect, useRef, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { mountStageWithScope, runAppPromise } from "../../../shared/runtime";
import {
  DEFAULT_ZOOM_SCALE,
  StageViewState,
  listMotionGroups,
  listSpriteEntries,
  selectPrimaryMusicTrack,
  stepZoomScale,
  type CharacterDetail,
} from "../domain";
import type { Live2DStageSession } from "../stage";

export interface CharacterDetailViewProps {
  readonly detail: CharacterDetail;
}

export function CharacterDetailView({ detail }: CharacterDetailViewProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const audioRef = useRef<HTMLAudioElement>(null);
  const sessionRef = useRef<Live2DStageSession | null>(null);
  const scaleRef = useRef<number>(DEFAULT_ZOOM_SCALE);

  const [stageState, setStageState] = useState<StageViewState>(() =>
    StageViewState.Loading({ modelPath: detail.modelPath }),
  );

  const [motionError, setMotionError] = useState<string | null>(null);

  const primaryTrack = selectPrimaryMusicTrack({ music: detail.music });
  const primaryTrackUrl = Option.getOrUndefined(primaryTrack);
  const motionGroups = listMotionGroups({ detail });
  const spriteEntries = listSpriteEntries({ sprites: detail.sprites });

  useEffect(() => {
    const audio = audioRef.current;

    if (audio === null || primaryTrackUrl === undefined) return;

    audio.loop = true;
    void audio.play().catch(() => undefined);

    return () => {
      audio.pause();
    };
  }, [primaryTrackUrl]);

  useEffect(() => {
    const canvas = canvasRef.current;

    if (canvas === null) return;

    const cleanupStage = mountStageWithScope({
      canvas,
      modelPath: detail.modelPath,
      initialScale: DEFAULT_ZOOM_SCALE,
      onReady: (session) => {
        sessionRef.current = session;
        scaleRef.current = DEFAULT_ZOOM_SCALE;
        setStageState(
          StageViewState.Ready({
            modelPath: detail.modelPath,
            scale: DEFAULT_ZOOM_SCALE,
            activeMotion: Option.none(),
          }),
        );
      },
      onError: (error) => {
        sessionRef.current = null;
        setStageState(
          StageViewState.Failed({
            modelPath: detail.modelPath,
            message: error.reason,
          }),
        );
      },
    });

    const handleWheel = (event: WheelEvent) => {
      event.preventDefault();
      const session = sessionRef.current;

      if (session === null) return;

      const nextScale = stepZoomScale({
        current: scaleRef.current,
        deltaY: event.deltaY,
      });

      if (nextScale === scaleRef.current) return;

      scaleRef.current = nextScale;
      void runAppPromise(session.setScale({ scale: nextScale }));
      setStageState((previous) =>
        Match.value(previous).pipe(
          Match.tagsExhaustive({
            Loading: () => previous,
            Failed: () => previous,
            Ready: (ready) =>
              StageViewState.Ready({
                modelPath: ready.modelPath,
                scale: nextScale,
                activeMotion: ready.activeMotion,
              }),
          }),
        ),
      );
    };

    canvas.addEventListener("wheel", handleWheel, { passive: false });

    return () => {
      canvas.removeEventListener("wheel", handleWheel);
      sessionRef.current = null;
      cleanupStage();
    };
  }, [detail.modelPath]);

  function handleMotionClick(motionGroup: string) {
    const session = sessionRef.current;

    if (session === null) return;

    setMotionError(null);
    void runAppPromise(session.playMotion({ motionGroup })).then(
      (started) => {
        if (!started) return;

        setStageState((previous) =>
          Match.value(previous).pipe(
            Match.tagsExhaustive({
              Loading: () => previous,
              Failed: () => previous,
              Ready: (ready) =>
                StageViewState.Ready({
                  modelPath: ready.modelPath,
                  scale: ready.scale,
                  activeMotion: Option.some(motionGroup),
                }),
            }),
          ),
        );
      },
      (error) => {
        setMotionError(
          error instanceof Error ? error.message : "Could not trigger motion playback.",
        );
      },
    );
  }

  const stageBanner = Match.value(stageState).pipe(
    Match.tagsExhaustive({
      Loading: () => (
        <Badge variant="secondary" role="status">
          Loading Live2D model...
        </Badge>
      ),
      Ready: (ready) => (
        <Badge variant="outline" role="status">
          {Option.match(ready.activeMotion, {
            onNone: () => `Scale ${ready.scale.toFixed(2)}x · Scroll canvas to zoom`,
            onSome: (motion) =>
              `Scale ${ready.scale.toFixed(2)}x · Playing ${motion} · Scroll canvas to zoom`,
          })}
        </Badge>
      ),
      Failed: (failed) => (
        <Badge variant="destructive" role="alert">
          Could not load Live2D stage: {failed.message}
        </Badge>
      ),
    }),
  );

  const isCanvasVisible = Match.value(stageState).pipe(
    Match.tagsExhaustive({
      Loading: () => false,
      Ready: () => true,
      Failed: () => false,
    }),
  );

  const activeMotion = Match.value(stageState).pipe(
    Match.tagsExhaustive({
      Loading: () => undefined,
      Failed: () => undefined,
      Ready: (ready) => Option.getOrUndefined(ready.activeMotion),
    }),
  );

  return (
    <main className="mx-auto min-h-[95vh] max-w-6xl px-5 pb-8">
      <div className="my-5 flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-bold text-foreground">
          Live2D Model · {detail.modelDetail.Name.length > 0 ? detail.modelDetail.Name : detail.id}
        </h1>
        {stageBanner}
      </div>

      <div className="flex flex-col items-center gap-6 sm:items-start md:flex-row">
        <div className="w-[90%] md:w-1/2">
          <canvas
            aria-label={`${detail.id} Live2D stage`}
            className={`w-full rounded-xl border border-border bg-card ${
              isCanvasVisible ? "" : "invisible"
            }`}
            ref={canvasRef}
          />
        </div>

        <section className="flex w-[90%] flex-col gap-6 md:w-1/2">
          <Card>
            <CardHeader>
              <CardTitle className="text-xl">{detail.id}</CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-3">
              {motionError !== null ? (
                <p role="alert" className="text-sm text-destructive">
                  {motionError}
                </p>
              ) : null}
              {motionGroups.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No motion groups defined for this model.
                </p>
              ) : (
                <div className="flex flex-row flex-wrap gap-2">
                  {Array.map(motionGroups, (motionGroup) => (
                    <Button
                      key={motionGroup}
                      type="button"
                      size="sm"
                      variant={activeMotion === motionGroup ? "default" : "secondary"}
                      disabled={!isCanvasVisible}
                      onClick={() => handleMotionClick(motionGroup)}
                    >
                      {motionGroup}
                    </Button>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-xl">Sprites</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex flex-wrap gap-4">
                {Array.map(spriteEntries, (sprite) => (
                  <figure
                    key={sprite.key}
                    className="w-36 rounded-lg border border-border bg-background p-2"
                  >
                    <img
                      className="h-32 w-full object-contain"
                      src={sprite.src}
                      alt={`${detail.id} ${sprite.label} sprite`}
                    />
                    <figcaption className="mt-1 text-center text-xs text-muted-foreground">
                      {sprite.label}
                    </figcaption>
                  </figure>
                ))}
              </div>
            </CardContent>
          </Card>

          {primaryTrackUrl !== undefined ? (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Theme Audio</CardTitle>
              </CardHeader>
              <CardContent>
                <audio ref={audioRef} controls src={primaryTrackUrl} className="w-full max-w-md" />
              </CardContent>
            </Card>
          ) : null}
        </section>
      </div>
    </main>
  );
}
