import { z } from "zod";

/**
 * Drive `files.list` 応答の1要素（`{id, name}`）を検証する schema。
 */
export const DriveFileEntrySchema = z.object({
  id: z.string(),
  name: z.string(),
});
