import { connect, Channel, ChannelModel, Connection } from 'amqplib';
import { config } from '../config';
import { logger } from '../logger';
import { EmailService } from '../email/EmailService';
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
      } catch (err: any) {
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
  await ch.bindQueue(
    config.RABBITMQ_USER_EVENTS_QUEUE,
    config.RABBITMQ_EXCHANGE,
    config.RABBITMQ_USER_EVENTS_ROUTING_KEY
  );

  await ch.prefetch(config.RABBITMQ_PREFETCH);

  const emailService = new EmailService();

  await ch.consume(
    config.RABBITMQ_USER_EVENTS_QUEUE,
    async (msg) => {
      if (!msg) return;
      try {
        const content = msg.content.toString('utf8');
        const payload = JSON.parse(content);

        logger.info({ payload }, 'Received user.created event');

        // Send welcome email
        const welcomeEmail: EmailPayload = {
          to: payload.email,
          subject: 'Welcome to English Learning App! 🎉',
          html: `
            <h1>Welcome ${payload.name || 'there'}!</h1>
            <p>Thank you for registering with us. We're excited to have you on board!</p>
            <p>Your account has been successfully created.</p>
            <p>Start your English learning journey today!</p>
            <br/>
            <p>Best regards,</p>
            <p>English Learning Team</p>
          `,
          text: `Welcome ${payload.name || 'there'}! Thank you for registering with us. Your account has been successfully created.`
        };

        await emailService.send(welcomeEmail);
        ch.ack(msg);

        logger.info({ email: payload.email }, 'Welcome email sent successfully');
      } catch (err: any) {
        logger.error({ err }, 'Failed to process user.created event');
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