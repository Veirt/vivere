/// <reference types="vite/client" />
import * as PIXI from "PIXI";

declare global {
  interface Window {
    PIXI: PIXI;
  }
}
