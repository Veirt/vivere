import { createRootRouteWithContext, Link, Outlet } from "@tanstack/react-router";
import { IconArrowLeft } from "@tabler/icons-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { RouterContext } from "../router";

function RootLayout() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border bg-primary-950 text-primary-50">
        <nav
          aria-label="Primary navigation"
          className="mx-auto flex max-w-6xl items-center gap-4 px-5 py-3"
        >
          <Link
            to="/"
            className="rounded-md px-2.5 py-1.5 text-sm font-medium text-primary-200 transition-colors hover:text-primary-50 [&.active]:font-semibold [&.active]:text-primary-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            Home
          </Link>
        </nav>
      </header>
      <Outlet />
    </div>
  );
}

function RootNotFound() {
  return (
    <main className="mx-auto max-w-lg px-5 py-12">
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Page not found</CardTitle>
          <CardDescription>The requested route does not exist.</CardDescription>
        </CardHeader>
        <CardContent className="flex justify-center">
          <Button asChild>
            <Link to="/">
              <IconArrowLeft className="size-4" />
              Back to Home
            </Link>
          </Button>
        </CardContent>
      </Card>
    </main>
  );
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootLayout,
  notFoundComponent: RootNotFound,
});
