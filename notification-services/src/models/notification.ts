import { z } from 'zod';

// Notification Template Schemas
export const NotificationTemplateSchema = z.object({
    id: z.string().uuid(),
    type: z.string(),
    title: z.string(),
    body: z.string(),
    data: z.record(z.any()).optional(),
    created_at: z.string().datetime(),
});

export const CreateNotificationTemplateSchema = z.object({
    type: z.string().min(1),
    title: z.string().min(1),
    body: z.string().min(1),
    data: z.record(z.any()).optional(),
});

export const UpdateNotificationTemplateSchema = z.object({
    type: z.string().min(1).optional(),
    title: z.string().min(1).optional(),
    body: z.string().min(1).optional(),
    data: z.record(z.any()).optional(),
});

// User Notification Schemas
export const UserNotificationSchema = z.object({
    id: z.string().uuid(),
    user_id: z.string().uuid(),
    notification_id: z.string().uuid(),
    is_read: z.boolean(),
    created_at: z.string().datetime(),
    read_at: z.string().datetime().optional(),
});

export const CreateUserNotificationSchema = z.object({
    user_id: z.string().uuid(),
    notification_id: z.string().uuid(),
});

export const MarkAsReadSchema = z.object({
    notification_ids: z.array(z.string().uuid()).min(1),
});

// Response Schemas
export const NotificationTemplateWithCountSchema = NotificationTemplateSchema.extend({
    user_count: z.number().int().min(0),
});

export const UserNotificationWithTemplateSchema = UserNotificationSchema.extend({
    template: NotificationTemplateSchema,
});

// Types
export type NotificationTemplate = z.infer<typeof NotificationTemplateSchema>;
export type CreateNotificationTemplate = z.infer<typeof CreateNotificationTemplateSchema>;
export type UpdateNotificationTemplate = z.infer<typeof UpdateNotificationTemplateSchema>;
export type UserNotification = z.infer<typeof UserNotificationSchema>;
export type CreateUserNotification = z.infer<typeof CreateUserNotificationSchema>;
export type MarkAsRead = z.infer<typeof MarkAsReadSchema>;
export type NotificationTemplateWithCount = z.infer<typeof NotificationTemplateWithCountSchema>;
export type UserNotificationWithTemplate = z.infer<typeof UserNotificationWithTemplateSchema>;
