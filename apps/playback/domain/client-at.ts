import type { UtcInstant } from "./utc-instant.ts";

/** 端末の時計で表した「その操作が起きた時刻」。merge の鍵で、信頼は低い（server の時計とは別の問い）。 */
export type ClientAt = UtcInstant;
