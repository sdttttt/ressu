/** The message of an error raised by a call into Go. */
export function messageOf(error: unknown): string {
  if (error instanceof Error) return error.message;
  return String(error);
}
