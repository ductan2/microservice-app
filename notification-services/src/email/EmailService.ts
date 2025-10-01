import { EmailProvider } from './EmailProvider';
import { SendgridProvider } from './SendgridProvider';
import { SmtpProvider } from './SmtpProvider';
import { config } from '../config';
import { EmailPayload } from './types';

export class EmailService {
  private provider: EmailProvider;

  constructor(customProvider?: EmailProvider) {
    this.provider =
      customProvider ?? (config.EMAIL_PROVIDER === 'sendgrid' ? new SendgridProvider() : new SmtpProvider());
  }

  async send(payload: EmailPayload) {
    return this.provider.sendEmail(payload);
  }
}
