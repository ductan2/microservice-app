import { db } from '../database/connection';
import { logger } from '../logger';
import {
    NotificationTemplate,
    CreateNotificationTemplate,
    UpdateNotificationTemplate,
    UserNotification,
    CreateUserNotification,
    NotificationTemplateWithCount,
    UserNotificationWithTemplate,
} from '../models/notification';

export class NotificationRepository {
    // Notification Templates
    async createTemplate(template: CreateNotificationTemplate): Promise<NotificationTemplate> {
        const query = `
      INSERT INTO notification_templates (type, title, body, data)
      VALUES ($1, $2, $3, $4)
      RETURNING *
    `;

        const values = [template.type, template.title, template.body, JSON.stringify(template.data || {})];

        try {
            const result = await db.query(query, values);
            const row = result.rows[0];
            return {
                id: row.id,
                type: row.type,
                title: row.title,
                body: row.body,
                data: row.data,
                created_at: row.created_at.toISOString(),
            };
        } catch (error) {
            logger.error({ error, template }, 'Failed to create notification template');
            throw error;
        }
    }

    async getTemplateById(id: string): Promise<NotificationTemplate | null> {
        const query = 'SELECT * FROM notification_templates WHERE id = $1';

        try {
            const result = await db.query(query, [id]);
            if (result.rows.length === 0) return null;

            const row = result.rows[0];
            return {
                id: row.id,
                type: row.type,
                title: row.title,
                body: row.body,
                data: row.data,
                created_at: row.created_at.toISOString(),
            };
        } catch (error) {
            logger.error({ error, id }, 'Failed to get notification template by id');
            throw error;
        }
    }

    async getAllTemplates(): Promise<NotificationTemplateWithCount[]> {
        const query = `
      SELECT 
        nt.*,
        COUNT(un.id) as user_count
      FROM notification_templates nt
      LEFT JOIN user_notifications un ON nt.id = un.notification_id
      GROUP BY nt.id, nt.type, nt.title, nt.body, nt.data, nt.created_at
      ORDER BY nt.created_at DESC
    `;

        try {
            const result = await db.query(query);
            return result.rows.map((row: any) => ({
                id: row.id,
                type: row.type,
                title: row.title,
                body: row.body,
                data: row.data,
                created_at: row.created_at.toISOString(),
                user_count: parseInt(row.user_count),
            }));
        } catch (error) {
            logger.error({ error }, 'Failed to get all notification templates');
            throw error;
        }
    }

    async updateTemplate(id: string, updates: UpdateNotificationTemplate): Promise<NotificationTemplate | null> {
        const fields = [];
        const values = [];
        let paramCount = 1;

        if (updates.type !== undefined) {
            fields.push(`type = $${paramCount++}`);
            values.push(updates.type);
        }
        if (updates.title !== undefined) {
            fields.push(`title = $${paramCount++}`);
            values.push(updates.title);
        }
        if (updates.body !== undefined) {
            fields.push(`body = $${paramCount++}`);
            values.push(updates.body);
        }
        if (updates.data !== undefined) {
            fields.push(`data = $${paramCount++}`);
            values.push(JSON.stringify(updates.data));
        }

        if (fields.length === 0) {
            return this.getTemplateById(id);
        }

        const query = `
      UPDATE notification_templates 
      SET ${fields.join(', ')}
      WHERE id = $${paramCount}
      RETURNING *
    `;
        values.push(id);

        try {
            const result = await db.query(query, values);
            if (result.rows.length === 0) return null;

            const row = result.rows[0];
            return {
                id: row.id,
                type: row.type,
                title: row.title,
                body: row.body,
                data: row.data,
                created_at: row.created_at.toISOString(),
            };
        } catch (error) {
            logger.error({ error, id, updates }, 'Failed to update notification template');
            throw error;
        }
    }

    async deleteTemplate(id: string): Promise<boolean> {
        const query = 'DELETE FROM notification_templates WHERE id = $1';

        try {
            const result = await db.query(query, [id]);
            return (result.rowCount ?? 0) > 0;
        } catch (error) {
            logger.error({ error, id }, 'Failed to delete notification template');
            throw error;
        }
    }

    // User Notifications
    async createUserNotification(notification: CreateUserNotification): Promise<UserNotification> {
        const query = `
      INSERT INTO user_notifications (user_id, notification_id)
      VALUES ($1, $2)
      RETURNING *
    `;

        const values = [notification.user_id, notification.notification_id];

        try {
            const result = await db.query(query, values);
            const row = result.rows[0];
            return {
                id: row.id,
                user_id: row.user_id,
                notification_id: row.notification_id,
                is_read: row.is_read,
                created_at: row.created_at.toISOString(),
                read_at: row.read_at ? row.read_at.toISOString() : undefined,
            };
        } catch (error) {
            logger.error({ error, notification }, 'Failed to create user notification');
            throw error;
        }
    }

    async getUserNotifications(
        userId: string,
        limit: number = 50,
        offset: number = 0,
        isRead?: boolean
    ): Promise<UserNotificationWithTemplate[]> {
        let query = `
      SELECT 
        un.*,
        nt.type,
        nt.title,
        nt.body,
        nt.data,
        nt.created_at as template_created_at
      FROM user_notifications un
      JOIN notification_templates nt ON un.notification_id = nt.id
      WHERE un.user_id = $1
    `;

        const values: any[] = [userId];
        let paramCount = 1;

        if (isRead !== undefined) {
            query += ` AND un.is_read = $${++paramCount}`;
            values.push(isRead);
        }

        query += ` ORDER BY un.created_at DESC LIMIT $${++paramCount} OFFSET $${++paramCount}`;
        values.push(limit, offset);

        try {
            const result = await db.query(query, values);
            return result.rows.map((row: any) => ({
                id: row.id,
                user_id: row.user_id,
                notification_id: row.notification_id,
                is_read: row.is_read,
                created_at: row.created_at.toISOString(),
                read_at: row.read_at ? row.read_at.toISOString() : undefined,
                template: {
                    id: row.notification_id,
                    type: row.type,
                    title: row.title,
                    body: row.body,
                    data: row.data,
                    created_at: row.template_created_at.toISOString(),
                },
            }));
        } catch (error) {
            logger.error({ error, userId, limit, offset, isRead }, 'Failed to get user notifications');
            throw error;
        }
    }

    async markNotificationsAsRead(userId: string, notificationIds: string[]): Promise<number> {
        const query = `
      UPDATE user_notifications 
      SET is_read = true, read_at = NOW()
      WHERE user_id = $1 AND id = ANY($2) AND is_read = false
    `;

        try {
            const result = await db.query(query, [userId, notificationIds]);
            return result.rowCount ?? 0;
        } catch (error) {
            logger.error({ error, userId, notificationIds }, 'Failed to mark notifications as read');
            throw error;
        }
    }

    async getUnreadCount(userId: string): Promise<number> {
        const query = `
      SELECT COUNT(*) as count
      FROM user_notifications
      WHERE user_id = $1 AND is_read = false
    `;

        try {
            const result = await db.query(query, [userId]);
            return parseInt(result.rows[0].count);
        } catch (error) {
            logger.error({ error, userId }, 'Failed to get unread count');
            throw error;
        }
    }

    async deleteUserNotification(id: string, userId: string): Promise<boolean> {
        const query = 'DELETE FROM user_notifications WHERE id = $1 AND user_id = $2';

        try {
            const result = await db.query(query, [id, userId]);
            return (result.rowCount ?? 0) > 0;
        } catch (error) {
            logger.error({ error, id, userId }, 'Failed to delete user notification');
            throw error;
        }
    }
}
