export interface UserRegistrationEmailParams {
  name?: string;
  appName?: string;
  dashboardUrl?: string;
  supportEmail?: string;
}

export interface PasswordResetEmailParams {
  name?: string;
  resetLink: string;
  expiresInMinutes?: number;
  appName?: string;
  supportEmail?: string;
}

export interface EmailVerificationParams {
  name?: string;
  verificationLink: string;
  appName?: string;
  supportEmail?: string;
}

export function buildUserRegistrationEmailTemplate(params: UserRegistrationEmailParams) {
  const {
    name,
    appName = 'English Learning App',
    dashboardUrl = 'https://app.example.com',
    supportEmail = 'support@example.com',
  } = params;

  const displayName = name || 'there';

  return {
    subject: `Welcome to ${appName}! 🎉`,
    html: `
      <!DOCTYPE html>
      <html>
      <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <style>
          body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; }
          .container { max-width: 600px; margin: 0 auto; padding: 20px; }
          .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
          .content { background: #ffffff; padding: 30px; border: 1px solid #e0e0e0; border-top: none; }
          .button { display: inline-block; padding: 12px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
          .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        </style>
      </head>
      <body>
        <div class="container">
          <div class="header">
            <h1>Welcome to ${appName}! 🎉</h1>
          </div>
          <div class="content">
            <h2>Hi ${displayName},</h2>
            <p>Thank you for registering with us! We're thrilled to have you on board.</p>
            <p>Your account has been successfully created and you can now start your English learning journey.</p>
            <p>Here's what you can do next:</p>
            <ul>
              <li>Complete your profile</li>
              <li>Browse available lessons</li>
              <li>Start your first learning session</li>
              <li>Track your progress</li>
            </ul>
            <p style="text-align: center;">
              <a href="${dashboardUrl}" class="button">Go to Dashboard</a>
            </p>
            <p>If you have any questions or need assistance, feel free to reach out to our support team.</p>
            <p>Happy learning!</p>
            <p><strong>Best regards,</strong><br>${appName} Team</p>
          </div>
          <div class="footer">
            <p>Need help? Contact us at <a href="mailto:${supportEmail}">${supportEmail}</a></p>
            <p>&copy; ${new Date().getFullYear()} ${appName}. All rights reserved.</p>
          </div>
        </div>
      </body>
      </html>
    `,
    text: `Welcome to ${appName}!

Hi ${displayName},

Thank you for registering with us! We're thrilled to have you on board.

Your account has been successfully created and you can now start your English learning journey.

Visit your dashboard: ${dashboardUrl}

If you have any questions, contact us at ${supportEmail}

Best regards,
${appName} Team`,
  };
}

export function buildPasswordResetEmailTemplate(params: PasswordResetEmailParams) {
  const {
    name,
    resetLink,
    expiresInMinutes = 60,
    appName = 'English Learning App',
    supportEmail = 'support@example.com',
  } = params;

  const displayName = name || 'there';

  return {
    subject: `Reset Your ${appName} Password`,
    html: `
      <!DOCTYPE html>
      <html>
      <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <style>
          body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; }
          .container { max-width: 600px; margin: 0 auto; padding: 20px; }
          .header { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
          .content { background: #ffffff; padding: 30px; border: 1px solid #e0e0e0; border-top: none; }
          .button { display: inline-block; padding: 12px 30px; background: #f5576c; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
          .warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
          .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        </style>
      </head>
      <body>
        <div class="container">
          <div class="header">
            <h1>🔐 Password Reset Request</h1>
          </div>
          <div class="content">
            <h2>Hi ${displayName},</h2>
            <p>We received a request to reset your password for your ${appName} account.</p>
            <p>Click the button below to reset your password:</p>
            <p style="text-align: center;">
              <a href="${resetLink}" class="button">Reset Password</a>
            </p>
            <div class="warning">
              <p><strong>⚠️ Important:</strong></p>
              <ul style="margin: 5px 0;">
                <li>This link will expire in ${expiresInMinutes} minutes</li>
                <li>If you didn't request this, you can safely ignore this email</li>
                <li>Your password will remain unchanged</li>
              </ul>
            </div>
            <p>For security reasons, we recommend that you:</p>
            <ul>
              <li>Choose a strong, unique password</li>
              <li>Don't share your password with anyone</li>
              <li>Enable two-factor authentication if available</li>
            </ul>
            <p>If you have any concerns about your account security, please contact us immediately.</p>
            <p><strong>Best regards,</strong><br>${appName} Team</p>
          </div>
          <div class="footer">
            <p>If the button doesn't work, copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #ffffff;">${resetLink}</p>
            <br>
            <p>Need help? Contact us at <a href="mailto:${supportEmail}">${supportEmail}</a></p>
            <p>&copy; ${new Date().getFullYear()} ${appName}. All rights reserved.</p>
          </div>
        </div>
      </body>
      </html>
    `,
    text: `Password Reset Request

Hi ${displayName},

We received a request to reset your password for your ${appName} account.

Reset your password by clicking this link:
${resetLink}

This link will expire in ${expiresInMinutes} minutes.

If you didn't request this, you can safely ignore this email and your password will remain unchanged.

For help, contact us at ${supportEmail}

Best regards,
${appName} Team`,
  };
}

export function buildEmailVerificationTemplate(params: EmailVerificationParams) {
  const {
    name,
    verificationLink,
    appName = 'English Learning App',
    supportEmail = 'support@example.com',
  } = params;

  const displayName = name || 'there';

  return {
    subject: `Verify Your ${appName} Email Address`,
    html: `
      <!DOCTYPE html>
      <html>
      <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <style>
          body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; }
          .container { max-width: 600px; margin: 0 auto; padding: 20px; }
          .header { background: linear-gradient(135deg, #4CAF50 0%, #2E7D32 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
          .content { background: #ffffff; padding: 30px; border: 1px solid #e0e0e0; border-top: none; }
          .button { display: inline-block; padding: 12px 30px; background: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
          .footer { text-align: center; margin-top: 30px; color: #666; font-size: 14px; }
        </style>
      </head>
      <body>
        <div class="container">
          <div class="header">
            <h1>✉️ Verify Your Email</h1>
          </div>
          <div class="content">
            <h2>Hi ${displayName},</h2>
            <p>Thank you for signing up for ${appName}!</p>
            <p>To complete your registration and start learning English, please verify your email address by clicking the button below:</p>
            <p style="text-align: center; color: #ffffff;">
              <a href="${verificationLink}" class="button">Verify Email Address</a>
            </p>
            <p><strong>This link will expire in 24 hours.</strong></p>
            <p>If you didn't create an account with ${appName}, you can safely ignore this email.</p>
            <p><strong>Best regards,</strong><br>${appName} Team</p>
          </div>
          <div class="footer">
            <p>If the button doesn't work, copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #4CAF50;">${verificationLink}</p>
            <br>
            <p>Need help? Contact us at <a href="mailto:${supportEmail}">${supportEmail}</a></p>
            <p>&copy; ${new Date().getFullYear()} ${appName}. All rights reserved.</p>
          </div>
        </div>
      </body>
      </html>
    `,
    text: `Verify Your Email Address

Hi ${displayName},

Thank you for signing up for ${appName}!

To complete your registration and start learning English, please verify your email address by clicking this link:
${verificationLink}

This link will expire in 24 hours.

If you didn't create an account with ${appName}, you can safely ignore this email.

Best regards,
${appName} Team`,
  };
}
