import { Router } from 'express';
import { z } from 'zod';
import { EmailService } from '../email/EmailService';

const router = Router();

router.get('/health', (_req, res) => {
  res.json({ status: 'ok' });
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
  } catch (err: any) {
    return res.status(502).json({ error: 'Failed to send email', message: err?.message });
  }
});

export default router;
