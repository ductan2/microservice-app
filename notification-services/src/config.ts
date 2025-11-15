import 'dotenv/config';
import { z } from 'zod';

const EnvSchema = z.object({
  PORT: z.coerce.number().int().positive().default(3000),
  EMAIL_PROVIDER: z.enum(['sendgrid', 'smtp']).default('smtp'),

  SENDGRID_API_KEY: z.string().optional(),
  SENDGRID_DEFAULT_FROM: z.string().optional(),

  SMTP_HOST: z.string().default('smtp.gmail.com'),
  SMTP_PORT: z.coerce.number().int().default(587),
  SMTP_SECURE: z
    .union([z.string(), z.boolean()])
    .transform((v) => (typeof v === 'string' ? v === 'true' : !!v))
    .default(false),
  SMTP_USER: z.string().optional(),
  SMTP_PASS: z.string().optional(),
  SMTP_DEFAULT_FROM: z.string().optional(),

  // RabbitMQ
  RABBITMQ_URL: z.string().default('amqp://localhost:5672'),
  RABBITMQ_EXCHANGE: z.string().default('notifications'),
  RABBITMQ_EMAIL_QUEUE: z.string().default('notifications.email'),
  RABBITMQ_EMAIL_ROUTING_KEY: z.string().default('email.send'),
  RABBITMQ_USER_EVENTS_QUEUE: z.string().default('notifications.user_events'),
  RABBITMQ_USER_EVENTS_ROUTING_KEY: z.string().default('user.created,user.password_reset,user.email_verification'),
  RABBITMQ_PREFETCH: z.coerce.number().int().positive().default(10),

  // PostgreSQL
  POSTGRES_USER: z.string().default('user'),
  POSTGRES_PASSWORD: z.string().default('password'),
  POSTGRES_DB_USER_SERVICES: z.string().default('lms_user_services'),
  POSTGRES_HOST: z.string().default('localhost'),
  POSTGRES_PORT: z.coerce.number().int().positive().default(5432),
  POSTGRES_SSLMODE: z.string().default('disable'),
});

const parsed = EnvSchema.safeParse(process.env);
if (!parsed.success) {
  // eslint-disable-next-line no-console
  console.error('Invalid environment variables:', parsed.error.flatten().fieldErrors);
  process.exit(1);
}

export const config = parsed.data;
