import { Router } from 'express';
import { z } from 'zod';
import { EmailService } from '../email/EmailService';
import { NotificationTemplateService } from '../services/notificationTemplates';

export const router = Router();

router.get('/health', (_req, res) => {
  res.json({ status: 'ok' });
});

const templateService = new NotificationTemplateService();

router.get('/notification-templates', async (_req, res) => {
  const templates = await templateService.listTemplates();
  res.json({ data: templates });
});

router.get('/notification-templates/:id', async (req, res) => {
  const template = await templateService.getTemplateById(req.params.id);
  if (!template) {
    return res.status(404).json({ error: 'Template not found' });
  }

  res.json({ data: template });
});

const EmailSchema = z.object({
  to: z.union([z.string(), z.array(z.string()).nonempty()]),
  subject: z.string().min(1),
  text: z.string().optional(),
  html: z.string().optional(),
  from: z.string().optional(),
});

router.post('/send-email', async (req, res) => {
  const parsed = EmailSchema.safeParse(req.body);
  if (!parsed.success) {
    return res.status(400).json({ error: 'Invalid payload', details: parsed.error.flatten() });
  }

  const service = new EmailService();
  try {
    const result = await service.send(parsed.data);
    return res.status(202).json(result ?? { status: 'queued' });
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : 'Unknown error';
    return res.status(502).json({ error: 'Failed to send email', message });
  }
});

