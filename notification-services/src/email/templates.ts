import { EmailPayload } from './types';

export interface UserRegistrationEmailData {
  name?: string | null;
  appName?: string;
  dashboardUrl?: string;
  supportEmail?: string;
}

export function buildUserRegistrationEmailTemplate(
  data: UserRegistrationEmailData
): Pick<EmailPayload, 'subject' | 'html' | 'text'> {
  const {
    appName = 'English Learning App',
    dashboardUrl,
    supportEmail,
  } = data;
  const name = data.name?.trim() || 'there';

  const subject = `Welcome to ${appName}! 🎉`;

  const htmlParts = [
    `<h1>Welcome ${name}!</h1>`,
    `<p>Thank you for registering with <strong>${appName}</strong>. We're excited to have you on board!</p>`,
    '<p>Your account has been successfully created.</p>',
  ];

  if (dashboardUrl) {
    htmlParts.push(
      `<p style="margin:24px 0"><a href="${dashboardUrl}" style="background-color:#2563eb;color:#ffffff;padding:12px 20px;border-radius:6px;text-decoration:none;font-weight:600;">Start learning now</a></p>`
    );
  }

  htmlParts.push(
    '<p>We hope you enjoy your learning journey and reach your goals soon.</p>',
    '<br/>',
    '<p>Best regards,</p>',
    `<p>${appName} Team${supportEmail ? `<br/><a href="mailto:${supportEmail}">${supportEmail}</a>` : ''}</p>`
  );

  const textLines = [
    `Welcome ${name}!`,
    `Thank you for registering with ${appName}.`,
    'Your account has been successfully created.',
  ];

  if (dashboardUrl) {
    textLines.push(`Start learning now: ${dashboardUrl}`);
  }

  textLines.push(
    'We hope you enjoy your learning journey and reach your goals soon.',
    'Best regards,',
    `${appName} Team${supportEmail ? ` (${supportEmail})` : ''}`
  );

  return {
    subject,
    html: htmlParts.join('\n'),
    text: textLines.join('\n'),
  };
}

export interface PasswordResetEmailData {
  name?: string | null;
  resetLink: string;
  expiresInMinutes?: number | null;
  appName?: string;
  supportEmail?: string;
}

export function buildPasswordResetEmailTemplate(
  data: PasswordResetEmailData
): Pick<EmailPayload, 'subject' | 'html' | 'text'> {
  if (!data.resetLink) {
    throw new Error('Password reset email template requires a reset link');
  }

  const {
    appName = 'English Learning App',
    supportEmail,
    expiresInMinutes,
  } = data;
  const name = data.name?.trim() || 'there';

  const subject = `${appName} password reset instructions`;

  const expiryMessage =
    typeof expiresInMinutes === 'number' && expiresInMinutes > 0
      ? `This link will expire in ${expiresInMinutes} minute${expiresInMinutes === 1 ? '' : 's'}.`
      : 'This link will expire soon for your security.';

  const html = `
    <h1>Hi ${name},</h1>
    <p>We received a request to reset your password for your <strong>${appName}</strong> account.</p>
    <p>Please click the button below to choose a new password:</p>
    <p style="margin:24px 0"><a href="${data.resetLink}" style="background-color:#2563eb;color:#ffffff;padding:12px 20px;border-radius:6px;text-decoration:none;font-weight:600;">Reset your password</a></p>
    <p>If the button doesn't work, copy and paste this link into your browser:</p>
    <p><a href="${data.resetLink}">${data.resetLink}</a></p>
    <p>${expiryMessage}</p>
    <p>If you did not request a password reset, you can safely ignore this email.</p>
    <br/>
    <p>Best regards,</p>
    <p>${appName} Team${supportEmail ? `<br/><a href="mailto:${supportEmail}">${supportEmail}</a>` : ''}</p>
  `;

  const text = [
    `Hi ${name},`,
    `We received a request to reset your password for your ${appName} account.`,
    'Use the link below to choose a new password:',
    data.resetLink,
    expiryMessage,
    'If you did not request a password reset, you can ignore this email.',
    'Best regards,',
    `${appName} Team${supportEmail ? ` (${supportEmail})` : ''}`,
  ].join('\n');

  return {
    subject,
    html,
    text,
  };
}
