import { defineConfig } from "drizzle-kit";

export default defineConfig({
  dialect: "sqlite",
  schema: "./worker/src/infrastructure/d1/schema.ts",
  out: "./worker/migrations",
});
