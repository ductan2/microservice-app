import { Router } from 'express';
import { z } from 'zod';
import { NotificationService } from '../services/notificationService';
import { logger } from '../logger';
import {
    CreateNotificationTemplateSchema,
    UpdateNotificationTemplateSchema,
    CreateUserNotificationSchema,
    MarkAsReadSchema,
} from '../models/notification';

const router = Router();
const notificationService = new NotificationService();

// Validation middleware
const validateBody = (schema: z.ZodSchema) => (req: any, res: any, next: any) => {
    try {
        req.body = schema.parse(req.body);
        next();
    } catch (error) {
        if (error instanceof z.ZodError) {
            return res.status(400).json({
                error: 'Validation error',
                details: error.errors,
            });
        }
        next(error);
    }
};

const validateParams = (schema: z.ZodSchema) => (req: any, res: any, next: any) => {
    try {
        req.params = schema.parse(req.params);
        next();
    } catch (error) {
        if (error instanceof z.ZodError) {
            return res.status(400).json({
                error: 'Invalid parameters',
                details: error.errors,
            });
        }
        next(error);
    }
};

const validateQuery = (schema: z.ZodSchema) => (req: any, res: any, next: any) => {
    try {
        req.query = schema.parse(req.query);
        next();
    } catch (error) {
        if (error instanceof z.ZodError) {
            return res.status(400).json({
                error: 'Invalid query parameters',
                details: error.errors,
            });
        }
        next(error);
    }
};

// Notification Template Routes
router.post('/templates', validateBody(CreateNotificationTemplateSchema), async (req, res) => {
    try {
        const template = await notificationService.createTemplate(req.body);
        res.status(201).json({
            success: true,
            data: template,
        });
    } catch (error) {
        logger.error({ error }, 'Failed to create notification template');
        res.status(500).json({
            success: false,
            error: 'Failed to create notification template',
        });
    }
});

router.get('/templates', async (req, res) => {
    try {
        const templates = await notificationService.getAllTemplates();
        res.json({
            success: true,
            data: templates,
        });
    } catch (error) {
        logger.error({ error }, 'Failed to get notification templates');
        res.status(500).json({
            success: false,
            error: 'Failed to get notification templates',
        });
    }
});

router.get('/templates/:id', validateParams(z.object({ id: z.string().uuid() })), async (req, res) => {
    try {
        const template = await notificationService.getTemplateById(req.params.id);
        if (!template) {
            return res.status(404).json({
                success: false,
                error: 'Notification template not found',
            });
        }
        res.json({
            success: true,
            data: template,
        });
    } catch (error) {
        logger.error({ error }, 'Failed to get notification template');
        res.status(500).json({
            success: false,
            error: 'Failed to get notification template',
        });
    }
});

router.put('/templates/:id',
    validateParams(z.object({ id: z.string().uuid() })),
    validateBody(UpdateNotificationTemplateSchema),
    async (req, res) => {
        try {
            const template = await notificationService.updateTemplate(req.params.id, req.body);
            if (!template) {
                return res.status(404).json({
                    success: false,
                    error: 'Notification template not found',
                });
            }
            res.json({
                success: true,
                data: template,
            });
        } catch (error) {
            logger.error({ error }, 'Failed to update notification template');
            res.status(500).json({
                success: false,
                error: 'Failed to update notification template',
            });
        }
    }
);

router.delete('/templates/:id', validateParams(z.object({ id: z.string().uuid() })), async (req, res) => {
    try {
        const deleted = await notificationService.deleteTemplate(req.params.id);
        if (!deleted) {
            return res.status(404).json({
                success: false,
                error: 'Notification template not found',
            });
        }
        res.json({
            success: true,
            message: 'Notification template deleted successfully',
        });
    } catch (error) {
        logger.error({ error }, 'Failed to delete notification template');
        res.status(500).json({
            success: false,
            error: 'Failed to delete notification template',
        });
    }
});

// User Notification Routes
router.post('/users/:userId/notifications',
    validateParams(z.object({ userId: z.string().uuid() })),
    validateBody(CreateUserNotificationSchema.omit({ user_id: true })),
    async (req, res) => {
        try {
            const notification = await notificationService.createUserNotification({
                user_id: req.params.userId,
                notification_id: req.body.notification_id,
            });
            res.status(201).json({
                success: true,
                data: notification,
            });
        } catch (error) {
            logger.error({ error }, 'Failed to create user notification');
            res.status(500).json({
                success: false,
                error: 'Failed to create user notification',
            });
        }
    }
);

router.get('/users/:userId/notifications',
    validateParams(z.object({ userId: z.string().uuid() })),
    validateQuery(z.object({
        limit: z.coerce.number().int().min(1).max(100).default(50),
        offset: z.coerce.number().int().min(0).default(0),
        is_read: z.coerce.boolean().optional(),
    })),
    async (req, res) => {
        try {
            const notifications = await notificationService.getUserNotifications(
                req.params.userId,
                req.query.limit as unknown as number,
                req.query.offset as unknown as number,
                req.query.is_read as unknown as boolean | undefined
            );
            res.json({
                success: true,
                data: notifications,
            });
        } catch (error) {
            logger.error({ error }, 'Failed to get user notifications');
            res.status(500).json({
                success: false,
                error: 'Failed to get user notifications',
            });
        }
    }
);

router.put('/users/:userId/notifications/read',
    validateParams(z.object({ userId: z.string().uuid() })),
    validateBody(MarkAsReadSchema),
    async (req, res) => {
        try {
            const updatedCount = await notificationService.markNotificationsAsRead(
                req.params.userId,
                req.body.notification_ids
            );
            res.json({
                success: true,
                data: {
                    updated_count: updatedCount,
                },
            });
        } catch (error) {
            logger.error({ error }, 'Failed to mark notifications as read');
            res.status(500).json({
                success: false,
                error: 'Failed to mark notifications as read',
            });
        }
    }
);

router.get('/users/:userId/notifications/unread-count',
    validateParams(z.object({ userId: z.string().uuid() })),
    async (req, res) => {
        try {
            const count = await notificationService.getUnreadCount(req.params.userId);
            res.json({
                success: true,
                data: {
                    unread_count: count,
                },
            });
        } catch (error) {
            logger.error({ error }, 'Failed to get unread count');
            res.status(500).json({
                success: false,
                error: 'Failed to get unread count',
            });
        }
    }
);

router.delete('/users/:userId/notifications/:notificationId',
    validateParams(z.object({
        userId: z.string().uuid(),
        notificationId: z.string().uuid(),
    })),
    async (req, res) => {
        try {
            const deleted = await notificationService.deleteUserNotification(
                req.params.notificationId,
                req.params.userId
            );
            if (!deleted) {
                return res.status(404).json({
                    success: false,
                    error: 'User notification not found',
                });
            }
            res.json({
                success: true,
                message: 'User notification deleted successfully',
            });
        } catch (error) {
            logger.error({ error }, 'Failed to delete user notification');
            res.status(500).json({
                success: false,
                error: 'Failed to delete user notification',
            });
        }
    }
);

// Bulk operations
router.post('/templates/:templateId/send',
    validateParams(z.object({ templateId: z.string().uuid() })),
    validateBody(z.object({
        user_ids: z.array(z.string().uuid()).min(1),
    })),
    async (req, res) => {
        try {
            const notifications = await notificationService.sendNotificationToUsers(
                req.params.templateId,
                req.body.user_ids
            );
            res.status(201).json({
                success: true,
                data: {
                    notifications_created: notifications.length,
                    notifications,
                },
            });
        } catch (error) {
            logger.error({ error }, 'Failed to send bulk notifications');
            res.status(500).json({
                success: false,
                error: 'Failed to send bulk notifications',
            });
        }
    }
);

export { router as notificationRouter };
