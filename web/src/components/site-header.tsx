import { Link } from "@tanstack/react-router";

import { Select, SelectContent, SelectItem, SelectTrigger } from "#/components/ui/select.tsx";
import { isLanguage, languages, type Language } from "#/lib/characters.ts";

type SiteHeaderProps = {
  language: Language;
  onLanguageChange: (language: Language) => void;
};

export function SiteHeader({ language, onLanguageChange }: SiteHeaderProps) {
  const selectedLanguage =
    languages.find((option) => option.value === language)?.label ?? "English";

  return (
    <header className="relative z-20 grid min-h-[4.5rem] items-center gap-4 border-b border-border/75 bg-card/72 px-5 py-3 text-foreground backdrop-blur-sm sm:px-9 lg:grid-cols-[auto_1fr_auto]">
      <Link
        className="font-display text-[2.55rem] font-normal leading-none tracking-[-0.05em] text-primary transition hover:text-foreground"
        to="/"
        search={{ lang: language, q: "", sort: "id" }}
        aria-label="Vivere home"
      >
        Vivere
      </Link>

      <nav
        className="hidden items-center gap-8 font-display text-[0.98rem] tracking-[0.03em] text-muted-foreground md:flex lg:pl-9"
        aria-label="Collection sections"
      >
        <Link
          className="archive-nav-active pb-2 text-primary"
          to="/"
          search={{ lang: language, q: "", sort: "id" }}
        >
          Characters
        </Link>
      </nav>

      <div className="flex items-center justify-between gap-4 lg:justify-end">
        <p className="hidden items-center gap-3 text-[0.7rem] font-semibold uppercase tracking-[0.28em] text-primary/70 xl:flex">
          <span aria-hidden="true" className="font-display text-lg leading-none text-primary/45">
            ✧
          </span>
          Illustrated archive
        </p>

        <span id="site-language-label" className="sr-only">
          Language
        </span>
        <Select
          value={language}
          onValueChange={(value) => {
            if (isLanguage(value)) {
              onLanguageChange(value);
            }
          }}
        >
          <SelectTrigger
            size="sm"
            className="w-[8.5rem] rounded-[0.65rem] border-border/75 bg-background/42 font-display text-foreground shadow-none"
            aria-labelledby="site-language-label"
          >
            <span>{selectedLanguage}</span>
          </SelectTrigger>
          <SelectContent align="end">
            {languages.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
    </header>
  );
}
