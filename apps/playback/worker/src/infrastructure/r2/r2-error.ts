export class R2Error extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "R2Error";
  }
}
