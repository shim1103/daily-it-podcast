import { describe, expect, it } from "vitest";
import { app } from "./app.ts";

describe("app", () => {
  it("Hono instance を export する", () => {
    expect(app).toBeDefined();
    expect(typeof app.fetch).toBe("function");
  });
});
