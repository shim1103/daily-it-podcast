import { MockDriveService } from '@daily-it-podcast/drive';
import type { DriveService } from '@daily-it-podcast/core';

// MVPではモック。将来は実Drive APIに差し替える。
export const driveService: DriveService = new MockDriveService();
