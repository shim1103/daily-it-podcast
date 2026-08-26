import { describe, expect, it } from "vitest";
import { rpcClient } from "./playback-rpc-client.ts";

describe("rpcClient", () => {
  it("Hono RPC client を export する", () => {
    expect(rpcClient).toBeDefined();
  });
});
