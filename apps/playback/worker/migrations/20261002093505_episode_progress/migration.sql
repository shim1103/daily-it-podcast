CREATE TABLE `episode_progress` (
	`episode_id` text PRIMARY KEY NOT NULL,
	`position_sec` real NOT NULL,
	`first_played_at` text NOT NULL,
	`first_completed_at` text,
	`last_played_at` text NOT NULL,
	`seq` integer DEFAULT 0 NOT NULL
);
--> statement-breakpoint
CREATE INDEX `episode_progress_seq_idx` ON `episode_progress` (`seq`);