import express from 'express';
import { config } from './config';
import { logger } from './logger';
import { router } from './routes/emailRoutes';
import { notificationRouter } from './routes/notificationRoutes';
import { initRabbitEmailConsumer, closeRabbit } from './messaging/rabbitmq';
import { initDatabase } from './database/connection';

async function main() {
  // Initialize database
  await initDatabase();

  // Initialize RabbitMQ
  await initRabbitEmailConsumer();

  const app = express();
  app.use(express.json());

  // Routes
  app.use('/email', router);
  app.use('/api/notifications', notificationRouter);

  const server = app.listen(config.PORT, () => {
    logger.info({ port: config.PORT }, 'Notification service listening');
  });

  process.on('SIGTERM', async () => {
    logger.info('SIGTERM received, shutting down');
    try {
      await closeRabbit();
    } catch (e) {
      logger.warn({ err: e }, 'Error closing RabbitMQ');
    }
    server.close(() => process.exit(0));
  });
}

main().catch((err) => {
  logger.error({ err }, 'Failed to start service');
  process.exit(1);
});
