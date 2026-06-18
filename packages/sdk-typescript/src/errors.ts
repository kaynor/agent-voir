export class AgentVoirError extends Error {
  readonly statusCode?: number;

  constructor(message: string, statusCode?: number) {
    super(message);
    this.name = "AgentVoirError";
    this.statusCode = statusCode;
  }
}
