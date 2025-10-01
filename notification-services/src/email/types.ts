export type EmailAddress = string; // e.g. "User <user@example.com>" or plain address

export interface EmailPayload {
  to: EmailAddress | EmailAddress[];
  subject: string;
  text?: string;
  html?: string;
  from?: EmailAddress;
}
