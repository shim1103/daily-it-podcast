export class PodcastError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly cause?: unknown,
  ) {
    super(message);
    this.name = 'PodcastError';
  }
}

export class InfoFetchError extends PodcastError {
  constructor(message: string, cause?: unknown) {
    super(message, 'INFO_FETCH_ERROR', cause);
    this.name = 'InfoFetchError';
  }
}

export class ScriptGenerationError extends PodcastError {
  constructor(message: string, cause?: unknown) {
    super(message, 'SCRIPT_GENERATION_ERROR', cause);
    this.name = 'ScriptGenerationError';
  }
}

export class TtsError extends PodcastError {
  constructor(message: string, cause?: unknown) {
    super(message, 'TTS_ERROR', cause);
    this.name = 'TtsError';
  }
}

export class DriveError extends PodcastError {
  constructor(message: string, cause?: unknown) {
    super(message, 'DRIVE_ERROR', cause);
    this.name = 'DriveError';
  }
}
