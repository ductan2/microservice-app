import { v4 as uuidv4 } from 'uuid';
import { NotificationRepository } from '../repositories/notificationRepository';
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

export class NotificationService {
    private repository: NotificationRepository;

    constructor() {
        this.repository = new NotificationRepository();
    }

    // Notification Template Services
    async createTemplate(templateData: CreateNotificationTemplate): Promise<NotificationTemplate> {
        try {
            const template = await this.repository.createTemplate(templateData);
            logger.info({ templateId: template.id }, 'Notification template created');
            return template;
        } catch (error) {
            logger.error({ error, templateData }, 'Failed to create notification template');
            throw error;
        }
    }

    async getTemplateById(id: string): Promise<NotificationTemplate | null> {
        try {
            return await this.repository.getTemplateById(id);
        } catch (error) {
            logger.error({ error, id }, 'Failed to get notification template');
            throw error;
        }
    }

    async getAllTemplates(): Promise<NotificationTemplateWithCount[]> {
        try {
            return await this.repository.getAllTemplates();
        } catch (error) {
            logger.error({ error }, 'Failed to get all notification templates');
            throw error;
        }
    }

    async updateTemplate(id: string, updates: UpdateNotificationTemplate): Promise<NotificationTemplate | null> {
        try {
            const template = await this.repository.updateTemplate(id, updates);
            if (template) {
                logger.info({ templateId: id }, 'Notification template updated');
            }
            return template;
        } catch (error) {
            logger.error({ error, id, updates }, 'Failed to update notification template');
            throw error;
        }
    }

    async deleteTemplate(id: string): Promise<boolean> {
        try {
            const deleted = await this.repository.deleteTemplate(id);
            if (deleted) {
                logger.info({ templateId: id }, 'Notification template deleted');
            }
            return deleted;
        } catch (error) {
            logger.error({ error, id }, 'Failed to delete notification template');
            throw error;
        }
    }

    // User Notification Services
    async createUserNotification(notificationData: CreateUserNotification): Promise<UserNotification> {
        try {
            // Verify template exists
            const template = await this.repository.getTemplateById(notificationData.notification_id);
            if (!template) {
                throw new Error('Notification template not found');
            }

            const notification = await this.repository.createUserNotification(notificationData);
            logger.info({
                notificationId: notification.id,
                userId: notification.user_id,
                templateId: notification.notification_id
            }, 'User notification created');

            return notification;
        } catch (error) {
            logger.error({ error, notificationData }, 'Failed to create user notification');
            throw error;
        }
    }

    async getUserNotifications(
        userId: string,
        limit: number = 50,
        offset: number = 0,
        isRead?: boolean
    ): Promise<UserNotificationWithTemplate[]> {
        try {
            return await this.repository.getUserNotifications(userId, limit, offset, isRead);
        } catch (error) {
            logger.error({ error, userId, limit, offset, isRead }, 'Failed to get user notifications');
            throw error;
        }
    }

    async markNotificationsAsRead(userId: string, notificationIds: string[]): Promise<number> {
        try {
            const updatedCount = await this.repository.markNotificationsAsRead(userId, notificationIds);
            logger.info({
                userId,
                notificationIds,
                updatedCount
            }, 'Notifications marked as read');
            return updatedCount;
        } catch (error) {
            logger.error({ error, userId, notificationIds }, 'Failed to mark notifications as read');
            throw error;
        }
    }

    async getUnreadCount(userId: string): Promise<number> {
        try {
            return await this.repository.getUnreadCount(userId);
        } catch (error) {
            logger.error({ error, userId }, 'Failed to get unread count');
            throw error;
        }
    }

    async deleteUserNotification(id: string, userId: string): Promise<boolean> {
        try {
            const deleted = await this.repository.deleteUserNotification(id, userId);
            if (deleted) {
                logger.info({ notificationId: id, userId }, 'User notification deleted');
            }
            return deleted;
        } catch (error) {
            logger.error({ error, id, userId }, 'Failed to delete user notification');
            throw error;
        }
    }

    // Bulk operations
    async sendNotificationToUsers(templateId: string, userIds: string[]): Promise<UserNotification[]> {
        try {
            // Verify template exists
            const template = await this.repository.getTemplateById(templateId);
            if (!template) {
                throw new Error('Notification template not found');
            }

            const notifications: UserNotification[] = [];

            for (const userId of userIds) {
                const notification = await this.repository.createUserNotification({
                    user_id: userId,
                    notification_id: templateId,
                });
                notifications.push(notification);
            }

            logger.info({
                templateId,
                userIds: userIds.length,
                notificationsCreated: notifications.length
            }, 'Bulk notifications sent');

            return notifications;
        } catch (error) {
            logger.error({ error, templateId, userIds }, 'Failed to send bulk notifications');
            throw error;
        }
    }
}
