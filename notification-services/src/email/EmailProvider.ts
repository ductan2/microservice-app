import { EmailPayload } from './types';

export interface EmailProvider {
  sendEmail(payload: EmailPayload): Promise<{ id?: string } | void>;
}
