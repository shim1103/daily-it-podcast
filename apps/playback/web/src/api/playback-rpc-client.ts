import { hc } from "hono/client";
import type { AppType } from "../../../worker/src/routes/app.ts";

export const rpcClient = hc<AppType>("");
