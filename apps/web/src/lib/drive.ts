import { MockDriveService, GoogleDriveService } from '@daily-it-podcast/drive';
import type { DriveService } from '@daily-it-podcast/core';

function createDriveService(): DriveService {
  const hasGoogleCreds =
    process.env['GOOGLE_CLIENT_ID'] &&
    process.env['GOOGLE_CLIENT_SECRET'] &&
    process.env['GOOGLE_REFRESH_TOKEN'] &&
    process.env['DRIVE_FOLDER_ID'];

  if (hasGoogleCreds) {
    return new GoogleDriveService();
  }
  return new MockDriveService();
}

export const driveService: DriveService = createDriveService();
