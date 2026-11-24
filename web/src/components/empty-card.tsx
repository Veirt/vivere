import { Card, CardContent, CardHeader, CardTitle } from "#/components/ui/card.tsx";

export function EmptyCard({ title, description }: { title: string; description: string }) {
  return (
    <Card className="archive-card border-border/80 bg-card/82 shadow-none">
      <CardHeader className="px-5 pb-0 sm:px-6">
        <CardTitle
          role="heading"
          aria-level={2}
          className="archive-section-title text-lg font-semibold tracking-[-0.01em] text-foreground"
        >
          <span>{title}</span>
        </CardTitle>
      </CardHeader>
      <CardContent className="px-5 sm:px-6">
        <p className="text-sm leading-7 text-muted-foreground">{description}</p>
      </CardContent>
    </Card>
  );
}
