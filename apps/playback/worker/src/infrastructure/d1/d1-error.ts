export class D1Error extends Error {
  constructor(message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "D1Error";
  }
}
