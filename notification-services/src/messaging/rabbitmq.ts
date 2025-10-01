import { connect, Channel, ChannelModel } from 'amqplib';
import { config } from '../config';
import { logger } from '../logger';
import { EmailService } from '../email/EmailService';
import { buildPasswordResetEmailTemplate, buildUserRegistrationEmailTemplate } from '../email/templates';
import { EmailPayload } from '../email/types';

let connection: ChannelModel | null = null;
let channel: Channel | null = null;

export async function initRabbitConsumers() {
  if (connection && channel) return; // already initialized

  logger.info({ url: config.RABBITMQ_URL }, 'Connecting to RabbitMQ');
  connection = await connect(config.RABBITMQ_URL) as unknown as ChannelModel;
  channel = await connection?.createChannel() as unknown as Channel;

  await channel?.assertExchange(config.RABBITMQ_EXCHANGE, 'topic', { durable: true });

  // Initialize Email Consumer
  await initEmailConsumer(channel);

  // Initialize User Events Consumer
  await initUserEventsConsumer(channel);

  // Handle connection close/errors
  connection.on('error', (err) => logger.error({ err }, 'RabbitMQ connection error'));
  connection.on('close', () => logger.warn('RabbitMQ connection closed'));

  logger.info('RabbitMQ consumers initialized');
}

async function initEmailConsumer(ch: Channel) {
  await ch.assertQueue(config.RABBITMQ_EMAIL_QUEUE, { durable: true });
  await ch.bindQueue(
    config.RABBITMQ_EMAIL_QUEUE,
    config.RABBITMQ_EXCHANGE,
    config.RABBITMQ_EMAIL_ROUTING_KEY
  );

  await ch.prefetch(config.RABBITMQ_PREFETCH);

  const service = new EmailService();

  await ch.consume(
    config.RABBITMQ_EMAIL_QUEUE,
    async (msg) => {
      if (!msg) return;
      try {
        const content = msg.content.toString('utf8');
        const payload: EmailPayload = JSON.parse(content);
        await service.send(payload);
        ch.ack(msg);
      } catch (err: unknown) {
        logger.error({ err }, 'Failed to process email message');
        // discard the message to avoid infinite redelivery loop
        ch.nack(msg, false, false);
      }
    },
    { noAck: false }
  );

  logger.info('Email consumer initialized');
}

async function initUserEventsConsumer(ch: Channel) {
  await ch.assertQueue(config.RABBITMQ_USER_EVENTS_QUEUE, { durable: true });
  const routingKeys = config.RABBITMQ_USER_EVENTS_ROUTING_KEY.split(',')
    .map((key) => key.trim())
    .filter(Boolean);

  if (routingKeys.length === 0) {
    routingKeys.push('user.created');
  }

  await Promise.all(
    routingKeys.map((key) =>
      ch.bindQueue(config.RABBITMQ_USER_EVENTS_QUEUE, config.RABBITMQ_EXCHANGE, key)
    )
  );

  await ch.prefetch(config.RABBITMQ_PREFETCH);

  const emailService = new EmailService();

  await ch.consume(
    config.RABBITMQ_USER_EVENTS_QUEUE,
    async (msg) => {
      if (!msg) return;
      try {
        const content = msg.content.toString('utf8');
        const payload = JSON.parse(content) as Record<string, unknown>;

        const payloadType = payload['type'];
        const eventType =
          msg.properties?.type ||
          (msg.properties?.headers?.event_type as string | undefined) ||
          (typeof payloadType === 'string' ? payloadType : undefined) ||
          msg.fields.routingKey;

        logger.info({ payload, eventType }, 'Received user event');

        const email = buildEmailFromUserEvent(eventType, payload);

        if (!email) {
          logger.warn({ eventType, payload }, 'No email generated for user event');
          ch.ack(msg);
          return;
        }

        await emailService.send(email);
        ch.ack(msg);

        const recipients = Array.isArray(email.to) ? email.to : [email.to];
        logger.info({ email: recipients, eventType }, 'User event email sent successfully');
      } catch (err: unknown) {
        logger.error({ err }, 'Failed to process user event');
        // discard the message to avoid infinite redelivery loop
        ch.nack(msg, false, false);
      }
    },
    { noAck: false }
  );

  logger.info('User events consumer initialized');
}

// Keep the old function name for backward compatibility
export const initRabbitEmailConsumer = initRabbitConsumers;

function buildEmailFromUserEvent(
  eventType: string | undefined,
  payload: Record<string, unknown>
): EmailPayload | null {
  const normalizedType = eventType?.toLowerCase();

  const email = getString(payload, 'email');
  if (!email) {
    throw new Error('User event payload is missing the email address');
  }

  switch (normalizedType) {
    case 'usercreated':
    case 'user.created':
      return {
        to: email,
        ...buildUserRegistrationEmailTemplate({
          name: getString(payload, 'name'),
          appName: getString(payload, 'appName'),
          dashboardUrl: getString(payload, 'dashboard_url', 'dashboardUrl'),
          supportEmail: getString(payload, 'support_email', 'supportEmail'),
        }),
      };
    case 'passwordresetrequested':
    case 'user.password_reset':
    case 'user.passwordreset': {
      const resetLink = getString(payload, 'reset_link', 'resetLink', 'reset_url', 'resetUrl');
      if (!resetLink) {
        throw new Error('Password reset event payload is missing reset link');
      }
      return {
        to: email,
        ...buildPasswordResetEmailTemplate({
          name: getString(payload, 'name'),
          resetLink,
          expiresInMinutes: getNumber(
            payload,
            'expires_in_minutes',
            'expiresInMinutes',
            'expires_in',
            'expiresIn'
          ),
          appName: getString(payload, 'appName'),
          supportEmail: getString(payload, 'support_email', 'supportEmail'),
        }),
      };
    }
    default:
      return null;
  }
}

function getString(payload: Record<string, unknown>, ...keys: string[]): string | undefined {
  for (const key of keys) {
    const value = payload[key];
    if (typeof value === 'string') {
      const trimmed = value.trim();
      if (trimmed.length > 0) {
        return trimmed;
      }
    }
  }
  return undefined;
}

function getNumber(payload: Record<string, unknown>, ...keys: string[]): number | undefined {
  for (const key of keys) {
    const value = payload[key];
    if (typeof value === 'number' && Number.isFinite(value)) {
      return value;
    }
    if (typeof value === 'string' && value.trim().length > 0) {
      const parsed = Number(value);
      if (Number.isFinite(parsed)) {
        return parsed;
      }
    }
  }
  return undefined;
}

export async function publishEmailMessage(payload: EmailPayload) {
  if (!channel) throw new Error('RabbitMQ channel not initialized');
  const buf = Buffer.from(JSON.stringify(payload));
  const ok = channel.publish(
    config.RABBITMQ_EXCHANGE,
    config.RABBITMQ_EMAIL_ROUTING_KEY,
    buf,
    { contentType: 'application/json', persistent: true }
  );
  return ok;
}

export async function closeRabbit() {
  try {
    if (channel) {
      await channel?.close();
      channel = null;
    }
  } finally {
    if (connection) {
      await connection?.close();
      connection = null;
    }
  }
}