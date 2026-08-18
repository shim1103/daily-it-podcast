export class EpisodeNotFoundError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "EpisodeNotFoundError";
  }
}
