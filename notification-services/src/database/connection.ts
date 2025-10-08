import { Pool } from 'pg';
import { config } from '../config';
import { logger } from '../logger';

const pool = new Pool({
    user: config.POSTGRES_USER,
    password: config.POSTGRES_PASSWORD,
    host: config.POSTGRES_HOST,
    port: config.POSTGRES_PORT,
    database: config.POSTGRES_DB,
    ssl: config.POSTGRES_SSLMODE === 'require' ? { rejectUnauthorized: false } : false,
});

pool.on('error', (err: Error) => {
    logger.error({ err }, 'Unexpected error on idle client');
});

export const db = pool;

export async function initDatabase() {
    try {
        // Create tables if they don't exist
        await db.query(`
      CREATE TABLE IF NOT EXISTS notification_templates (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        type TEXT NOT NULL,
        title TEXT NOT NULL,
        body TEXT NOT NULL,
        data JSONB,
        created_at TIMESTAMPTZ DEFAULT NOW()
      );
    `);

        await db.query(`
      CREATE TABLE IF NOT EXISTS user_notifications (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id UUID NOT NULL,
        notification_id UUID NOT NULL REFERENCES notification_templates(id) ON DELETE CASCADE,
        is_read BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMPTZ DEFAULT NOW(),
        read_at TIMESTAMPTZ
      );
    `);

        // Create indexes for better performance
        await db.query(`
      CREATE INDEX IF NOT EXISTS idx_user_notifications_user_id ON user_notifications(user_id);
    `);

        await db.query(`
      CREATE INDEX IF NOT EXISTS idx_user_notifications_is_read ON user_notifications(is_read);
    `);

        await db.query(`
      CREATE INDEX IF NOT EXISTS idx_user_notifications_created_at ON user_notifications(created_at);
    `);

        logger.info('Database initialized successfully');
    } catch (error) {
        logger.error({ error }, 'Failed to initialize database');
        throw error;
    }
}
