# Notification Services (Email)

A simple Node.js TypeScript microservice for sending emails via SendGrid or SMTP (Gmail).

## Features
- SendGrid and SMTP (Gmail) providers
- Express REST API with `/health` and `POST /send-email`
- Provider selection via env (`EMAIL_PROVIDER=sendgrid|smtp`)
- Type-safe config validation with Zod

## Requirements
- Node.js 18+

## Setup
1. Copy env template and fill in your credentials:
   ```bash
   cp .env.example .env
   ```
   - For Gmail SMTP, you will need an App Password (recommended) or OAuth2. Standard account passwords generally won't work with SMTP.

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start in development mode:
   ```bash
   npm run dev
   ```

4. Or build and run:
   ```bash
   npm run build
   npm start
   ```

## API
- `GET /health` -> `{ status: 'ok' }`
- `POST /send-email`
  - Body (JSON):
    ```json
    {
      "to": "recipient@example.com",
      "subject": "Hello",
      "text": "Plain text body",
      "html": "<strong>HTML body</strong>",
      "from": "Optional From <optional@example.com>"
    }
    ```
  - Response: `{ id: string }` or `{ status: 'queued' }` depending on provider

## Environment Variables
See `.env.example` for all variables. Key ones:
- `EMAIL_PROVIDER`: `sendgrid` or `smtp`
- For SendGrid: `SENDGRID_API_KEY`, `SENDGRID_DEFAULT_FROM`
- For SMTP: `SMTP_HOST`, `SMTP_PORT`, `SMTP_SECURE`, `SMTP_USER`, `SMTP_PASS`, `SMTP_DEFAULT_FROM`

## Notes
- For Gmail, enable 2FA and create an App Password, then use that for `SMTP_PASS`.
