import { createRouter } from "@tanstack/react-router";
import type { Effect } from "effect";

import { routeTree } from "./routeTree.gen";
import { runAppPromise, type AppServices } from "./shared/runtime";

export interface RouterContext {
  readonly runEffect: <A, E>(
    effect: Effect.Effect<A, E, AppServices>,
    signal?: AbortSignal,
  ) => Promise<A>;
}

export const router = createRouter({
  routeTree,
  context: {
    runEffect: runAppPromise,
  },
  defaultPreload: "intent",
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
