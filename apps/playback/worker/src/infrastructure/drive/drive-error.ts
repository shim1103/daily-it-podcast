export class DriveError extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "DriveError";
  }
}
