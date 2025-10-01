import sg from '@sendgrid/mail';
import { EmailProvider } from './EmailProvider';
import { EmailPayload } from './types';
import { config } from '../config';

export class SendgridProvider implements EmailProvider {
  constructor() {
    if (!config.SENDGRID_API_KEY) {
      throw new Error('SENDGRID_API_KEY is required for SendGrid provider');
    }
    sg.setApiKey(config.SENDGRID_API_KEY);
  }

  async sendEmail(payload: EmailPayload) {
    const from = payload.from || config.SENDGRID_DEFAULT_FROM;
    if (!from) {
      throw new Error('Missing From: provide payload.from or set SENDGRID_DEFAULT_FROM');
    }

    const res = await sg.send({
      to: payload.to as any,
      from,
      subject: payload.subject,
      text: payload.text || '',
      html: payload.html,
    });

    // SendGrid returns an array of responses; use the first message-id if available
    const msgId = res?.[0]?.headers?.['x-message-id'] as string | undefined;
    return { id: msgId };
  }
}
