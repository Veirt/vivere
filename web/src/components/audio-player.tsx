import * as React from "react";
import { createStore } from "@tanstack/store";
import { Pause, Play, RotateCcw } from "lucide-react";

import { Button } from "#/components/ui/button.tsx";
import { Slider } from "#/components/ui/slider.tsx";
import { Tooltip, TooltipContent, TooltipTrigger } from "#/components/ui/tooltip.tsx";
import { cn } from "#/lib/utils.ts";

const activeAudioPlayerStore = createStore<string | null>(null);

type AudioPlayerSize = "default" | "compact";

type AudioPlayerProps = {
  label: string;
  src: string;
  autoPlay?: boolean;
  size?: AudioPlayerSize;
  className?: string;
};

type AudioState = {
  src: string;
  currentTime: number;
  duration: number;
  isPlaying: boolean;
  isLoading: boolean;
  error: string | null;
};

function makeInitialAudioState(src: string): AudioState {
  return {
    src,
    currentTime: 0,
    duration: 0,
    isPlaying: false,
    isLoading: false,
    error: null,
  };
}

function formatTime(seconds: number) {
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return "0:00";
  }

  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = Math.floor(seconds % 60)
    .toString()
    .padStart(2, "0");
  return `${minutes}:${remainingSeconds}`;
}

function useAudioController({ src, autoPlay = false }: Pick<AudioPlayerProps, "src" | "autoPlay">) {
  const audioRef = React.useRef<HTMLAudioElement | null>(null);
  const instanceId = React.useId();
  const [rawState, setRawState] = React.useState<AudioState>(() => makeInitialAudioState(src));
  const state = rawState.src === src ? rawState : makeInitialAudioState(src);

  React.useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;

    const setPartialState = (nextState: Partial<Omit<AudioState, "src">>) => {
      setRawState((previous) => ({
        ...(previous.src === src ? previous : makeInitialAudioState(src)),
        ...nextState,
      }));
    };

    const handleLoadedMetadata = () => {
      setPartialState({ duration: audio.duration || 0, isLoading: false, error: null });
    };
    const handleTimeUpdate = () => {
      setPartialState({ currentTime: audio.currentTime || 0, duration: audio.duration || 0 });
    };
    const handleWaiting = () => setPartialState({ isLoading: true });
    const handleCanPlay = () =>
      setPartialState({ isLoading: false, duration: audio.duration || 0 });
    const handlePlay = () => setPartialState({ isPlaying: true, isLoading: false, error: null });
    const handlePause = () => setPartialState({ isPlaying: false, isLoading: false });
    const handleEnded = () => setPartialState({ isPlaying: false, currentTime: 0 });
    const handleError = () =>
      setPartialState({ isPlaying: false, isLoading: false, error: "Audio unavailable" });
    audio.addEventListener("loadedmetadata", handleLoadedMetadata);
    audio.addEventListener("timeupdate", handleTimeUpdate);
    audio.addEventListener("waiting", handleWaiting);
    audio.addEventListener("canplay", handleCanPlay);
    audio.addEventListener("play", handlePlay);
    audio.addEventListener("pause", handlePause);
    audio.addEventListener("ended", handleEnded);
    audio.addEventListener("error", handleError);
    const subscription = activeAudioPlayerStore.subscribe((activeId) => {
      if (activeId !== instanceId && !audio.paused) {
        audio.pause();
      }
    });

    if (autoPlay) {
      activeAudioPlayerStore.setState(() => instanceId);
      try {
        const playAttempt = audio.play();
        if (playAttempt && typeof playAttempt.catch === "function") {
          void playAttempt.catch(() => {
            setPartialState({ isPlaying: false, isLoading: false });
          });
        }
      } catch {
        setPartialState({ isPlaying: false, isLoading: false });
      }
    }

    return () => {
      audio.pause();
      if (activeAudioPlayerStore.state === instanceId) {
        activeAudioPlayerStore.setState(() => null);
      }
      audio.removeEventListener("loadedmetadata", handleLoadedMetadata);
      audio.removeEventListener("timeupdate", handleTimeUpdate);
      audio.removeEventListener("waiting", handleWaiting);
      audio.removeEventListener("canplay", handleCanPlay);
      audio.removeEventListener("play", handlePlay);
      audio.removeEventListener("pause", handlePause);
      audio.removeEventListener("ended", handleEnded);
      audio.removeEventListener("error", handleError);
      subscription.unsubscribe();
    };
  }, [autoPlay, instanceId, src]);

  const play = React.useCallback(() => {
    const audio = audioRef.current;
    if (!audio) return;

    activeAudioPlayerStore.setState(() => instanceId);
    setRawState((previous) => ({
      ...(previous.src === src ? previous : makeInitialAudioState(src)),
      isLoading: true,
      error: null,
    }));

    try {
      const playAttempt = audio.play();
      if (playAttempt && typeof playAttempt.then === "function") {
        void playAttempt.then(undefined, () => {
          setRawState((previous) => ({
            ...(previous.src === src ? previous : makeInitialAudioState(src)),
            isLoading: false,
            isPlaying: false,
            error: "Playback blocked",
          }));
        });
      }
    } catch {
      setRawState((previous) => ({
        ...(previous.src === src ? previous : makeInitialAudioState(src)),
        isLoading: false,
        isPlaying: false,
        error: "Playback blocked",
      }));
    }
  }, [instanceId, src]);

  const pause = React.useCallback(() => {
    audioRef.current?.pause();
  }, []);

  const toggle = React.useCallback(() => {
    if (state.isPlaying) {
      pause();
      return;
    }

    play();
  }, [pause, play, state.isPlaying]);

  const seek = React.useCallback(
    (value: Array<number>) => {
      const audio = audioRef.current;
      if (!audio) return;

      const nextTime = value[0] ?? 0;
      audio.currentTime = nextTime;
      setRawState((previous) => ({
        ...(previous.src === src ? previous : makeInitialAudioState(src)),
        currentTime: nextTime,
      }));
    },
    [src],
  );

  const replay = React.useCallback(() => {
    const audio = audioRef.current;
    if (!audio) return;

    audio.currentTime = 0;
    setRawState((previous) => ({
      ...(previous.src === src ? previous : makeInitialAudioState(src)),
      currentTime: 0,
    }));
    play();
  }, [play, src]);

  return {
    audioRef,
    state,
    toggle,
    seek,
    replay,
    formattedCurrentTime: formatTime(state.currentTime),
    formattedDuration: formatTime(state.duration),
  };
}

export function AudioPlayer({
  label,
  src,
  autoPlay = false,
  size = "default",
  className,
}: AudioPlayerProps) {
  const { audioRef, state, toggle, seek, replay, formattedCurrentTime, formattedDuration } =
    useAudioController({ src, autoPlay });
  const isCompact = size === "compact";
  const canSeek = state.duration > 0 && !state.error;

  return (
    <figure
      className={cn(
        "archive-audio-player grid gap-3 rounded-[0.85rem] border border-border/70 bg-background/48 shadow-[inset_0_1px_0_rgba(255,255,255,0.58)]",
        isCompact ? "p-3" : "p-4",
        className,
      )}
    >
      <audio ref={audioRef} src={src} preload="metadata" aria-hidden="true" />

      <div className="grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              type="button"
              variant="outline"
              size={isCompact ? "icon-sm" : "icon"}
              className="archive-audio-button rounded-full border-primary/20 bg-foreground text-background shadow-none hover:bg-primary hover:text-primary-foreground"
              aria-label={label}
              disabled={Boolean(state.error)}
              onClick={toggle}
            >
              {state.isPlaying ? (
                <Pause className="size-4" aria-hidden="true" />
              ) : (
                <Play className="size-4 translate-x-px" aria-hidden="true" />
              )}
            </Button>
          </TooltipTrigger>
          <TooltipContent>{state.isPlaying ? "Pause" : "Play"}</TooltipContent>
        </Tooltip>

        <figcaption className="min-w-0">
          <span className="block truncate text-xs font-medium text-foreground">{label}</span>
          {state.error ? (
            <span className="mt-0.5 block text-[0.68rem] text-destructive">{state.error}</span>
          ) : state.isLoading ? (
            <span className="mt-0.5 block text-[0.68rem] text-muted-foreground">Loading</span>
          ) : null}
        </figcaption>

        <div
          className="text-right text-[0.68rem] tabular-nums text-muted-foreground"
          aria-label={`${formattedCurrentTime} of ${formattedDuration}`}
        >
          <span>{formattedCurrentTime}</span>
          <span className="mx-1 text-border">/</span>
          <span>{formattedDuration}</span>
        </div>
      </div>

      <div className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3">
        <Slider
          aria-label={`Seek ${label}`}
          value={[Math.min(state.currentTime, state.duration || 0)]}
          min={0}
          max={state.duration || 1}
          step={0.1}
          disabled={!canSeek}
          onValueChange={seek}
          className="archive-audio-slider"
        />
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              type="button"
              variant="ghost"
              size="icon-xs"
              className="rounded-full text-muted-foreground hover:bg-accent/40 hover:text-primary"
              aria-label={`Replay ${label}`}
              disabled={Boolean(state.error)}
              onClick={replay}
            >
              <RotateCcw className="size-3" aria-hidden="true" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>Replay</TooltipContent>
        </Tooltip>
      </div>
    </figure>
  );
}
