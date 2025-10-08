export interface NotificationTemplate {
  id: string;
  name: string;
  description: string;
  subject: string;
  body: string;
  placeholders: string[];
  updated_at: string;
}

const mockNotificationTemplates: NotificationTemplate[] = [
  {
    id: 'welcome',
    name: 'Welcome Email',
    description: 'Sent to new users when they join the platform.',
    subject: 'Welcome to LearnHub!',
    body:
      "Hi {{user_name}},\n\nWe're excited to have you at LearnHub. Start exploring courses that match your goals today!\n\nHappy learning,\nThe LearnHub Team",
    placeholders: ['user_name'],
    updated_at: '2024-10-01T12:00:00Z',
  },
  {
    id: 'course-update',
    name: 'Course Update',
    description: 'Notify learners when new lessons are added to a course they follow.',
    subject: "New lesson available: {{lesson_title}}",
    body:
      "Hello {{user_name}},\n\nA new lesson '{{lesson_title}}' has just been published in {{course_title}}. Jump back in and continue your progress!\n\nView the lesson: {{lesson_url}}\n\nBest regards,\nLearnHub Team",
    placeholders: ['user_name', 'lesson_title', 'course_title', 'lesson_url'],
    updated_at: '2024-10-05T09:30:00Z',
  },
  {
    id: 'deadline-reminder',
    name: 'Deadline Reminder',
    description: 'Friendly reminder to complete upcoming assignments or quizzes.',
    subject: 'Reminder: {{item_title}} is due soon',
    body:
      "Hi {{user_name}},\n\nThis is a quick reminder that {{item_title}} is due on {{due_date}}. Stay on track and submit before the deadline!\n\nNeed help? Reply to this email and our team will assist you.\n\nThanks,\nLearnHub Support",
    placeholders: ['user_name', 'item_title', 'due_date'],
    updated_at: '2024-10-07T16:45:00Z',
  },
];

export class NotificationTemplateService {
  async listTemplates(): Promise<NotificationTemplate[]> {
    return mockNotificationTemplates;
  }

  async getTemplateById(id: string): Promise<NotificationTemplate | undefined> {
    return mockNotificationTemplates.find((template) => template.id === id);
  }
}
