import { AudioPlayer } from "#/components/audio-player.tsx";
import type { AudioSource } from "#/lib/character-media.ts";

export function AudioPreview({
  label,
  src,
  autoPlay = false,
}: AudioSource & { autoPlay?: boolean }) {
  return <AudioPlayer label={label} src={src} autoPlay={autoPlay} />;
}

export function AudioList({
  title,
  sources,
}: {
  title: string;
  sources: ReadonlyArray<AudioSource>;
}) {
  if (sources.length === 0) {
    return <p className="mt-3 text-xs text-muted-foreground">No local audio file found.</p>;
  }

  return (
    <div className="mt-4 grid gap-2 sm:grid-cols-2 xl:grid-cols-3" aria-label={title}>
      {sources.map((source) => (
        <AudioPlayer
          key={`${source.label}-${source.src}`}
          label={source.label}
          src={source.src}
          size="compact"
        />
      ))}
    </div>
  );
}
