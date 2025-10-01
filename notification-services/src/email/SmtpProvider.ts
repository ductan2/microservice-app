import nodemailer from 'nodemailer';
import { EmailProvider } from './EmailProvider';
import { EmailPayload } from './types';
import { config } from '../config';

export class SmtpProvider implements EmailProvider {
  private transporter = nodemailer.createTransport({
    host: config.SMTP_HOST,
    port: config.SMTP_PORT,
    secure: config.SMTP_SECURE,
    auth:
      config.SMTP_USER && config.SMTP_PASS
        ? { user: config.SMTP_USER, pass: config.SMTP_PASS }
        : undefined,
  });

  async sendEmail(payload: EmailPayload) {
    const from = payload.from || config.SMTP_DEFAULT_FROM || config.SMTP_USER;
    if (!from) {
      throw new Error('Missing From: provide payload.from or set SMTP_DEFAULT_FROM/SMTP_USER');
    }

    const info = await this.transporter.sendMail({
      from,
      to: payload.to,
      subject: payload.subject,
      text: payload.text,
      html: payload.html,
    });

    return { id: info.messageId };
  }
}
