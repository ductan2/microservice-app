Prompt for Building Admin DashboardI need you to build a complete Admin Dashboard for a Learning Management System with the following requirements:🎯 Main Requirements1. Technology Stack

Frontend: React with TypeScript
UI Framework: Tailwind CSS + shadcn/ui components
State Management: TanStack Query (React Query) for data fetching & caching
Table Management: TanStack Table for advanced data tables
Form Management: React Hook Form with Zod validation
Icons: lucide-react
Routing: Simulate routing with state (no react-router available)
2. Core FeaturesA. Media Assets Management (HIGH PRIORITY)Folder Management:

Tree view to display folder hierarchy
Create/rename/delete folders
Move folders (drag & drop or context menu)
Breadcrumb navigation
Statistics: file count, total size per folder
Nested folder support (unlimited depth)
Folder permissions (optional: created_by tracking)
Image/Media Management:

Grid view with thumbnails
List view option (toggle)
Upload files (image, video, audio) with drag & drop
Direct media preview (lightbox for images, inline player for video/audio)
Metadata display: name, size, mime type, upload date, uploader
Bulk actions: select multiple files to delete/move
Filter by media type (image/video/audio/document)
Search by filename
Copy URL/storage_key with one click
Sort by: name, size, date, type
File details panel/drawe
B. Lesson Management

Full CRUD operations for lessons
Rich text editor for description (TipTap or similar if available, otherwise textarea)
Add/edit/delete lesson sections with types:

Text content
Video embed
Image
Quiz reference


Drag & drop to reorder sections
Publish/unpublish lessons
Preview mode (student view)
Filter by topic, level, published status
Version control display
Associate media assets with lessons
C. Quiz Management

Create quiz attached to lesson
Add/edit/delete questions with types:

Multiple choice
True/False
Fill in the blank
Short answer
Audio/Speaking questions (with prompt_media_id)


Manage options for each question (correct answer marking)
Set points per question
Configure time limits
Preview quiz with answer key
Question bank/reusable questions (optional)
D. User Management

User list with TanStack Table (sortable, filterable, paginated)
Filters: status, email_verified, date range
View detailed user profile
Enrollment history with progress
Quiz attempts with scores
Streaks and points tracking
Ban/unban users
Reset password (admin action)
User activity timeline
E. Dashboard Analytics
Overview Cards:

Total users (with growth %)
Active users today/week
Total lessons, published lessons
Total enrollments, completion rate
Total media storage used (with quota if applicable)
Average quiz score
Charts (use Recharts):

User registration over time (line chart)
Lesson completion rates (bar chart)
Top performers leaderboard (current period)
Quiz score distribution (histogram)
Popular lessons (most enrollments)
Media storage by type (pie chart)
Quick Actions:

Create new lesson
Upload media
View recent user activity
Pending quiz reviews
F. Leaderboard Management

View leaderboard snapshots by period (daily/weekly/monthly)
Manual trigger to generate new snapshot
Export leaderboard data
Filter by date range
User detail drill-down
3. UI/UX RequirementsLayout:
┌─────────────────────────────────────┐
│         Top Navigation Bar          │
│  Logo | Breadcrumb | User Menu      │
├──────┬──────────────────────────────┤
│      │                              │
│ Side │   Main Content Area          │
│ bar  │                              │
│      │   (Dynamic based on route)   │
│      │                              │
└──────┴──────────────────────────────┘Sidebar Navigation:

📊 Dashboard
🖼️ Media Library

Folders
All Assets


📚 Lessons

All Lessons
Topics
Levels


❓ Quizzes
👥 Users
🏆 Leaderboard
⚙️ Settings
Design Principles:

Modern, clean, professional aesthetic
Responsive design (desktop-first, but mobile-friendly)
Loading skeletons for async content
Error boundaries with helpful messages
Toast notifications for actions (success/error)
Confirmation dialogs for destructive actions
Empty states with CTAs
Consistent spacing and typography
Dark mode support (optional but nice to have)
Key UI Patterns:

Modal dialogs for forms (create/edit)
Dropdown menus for action buttons
Tabs for organized content sections
Pagination with page size options
Debounced search inputs
Sortable table headers
Collapsible sidebar
Context menus (right-click actions)
Keyboard shortcuts (optional)
5. Media Library Detailed Specs
Upload Flow:

Click "Upload" button or drag files to drop zone
Select destination folder from dropdown
Choose multiple files
Show upload progress for each file (progress bar)
Auto-generate thumbnails for images (canvas API)
Validate file types and sizes
Show success/error for each upload
Auto-refresh the current folder view
File Operations:

View file details in side panel
Download file (blob URL)
Delete single or multiple files
Move files to another folder (modal with folder tree)
Copy storage_key or full URL to clipboard
Replace file while keeping same ID
Rename file

Folder Operations:

Create subfolder (modal input)
Rename folder (inline edit or modal)
Delete folder (with confirmation, show file count)
Move folder to different parent (drag & drop or modal)
Calculate and display folder size recursively
Show file count badge

6. Mock Data & API Layer
API Service Structure:
typescript// src/services/api.ts
export const api = {
  users: {
    getAll: (filters?: UserFilters) => Promise<User[]>,
    getById: (id: string) => Promise<User>,
    create: (data: CreateUserDto) => Promise<User>,
    update: (id: string, data: UpdateUserDto) => Promise<User>,
    delete: (id: string) => Promise<void>,
  },
  media: {
    getFolders: () => Promise<Folder[]>,
    getAssets: (folderId?: string) => Promise<MediaAsset[]>,
    createFolder: (data: CreateFolderDto) => Promise<Folder>,
    uploadAssets: (formData: FormData) => Promise<MediaAsset[]>,
    deleteAsset: (id: string) => Promise<void>,
    moveAssets: (ids: string[], folderId: string) => Promise<void>,
  },
  lessons: {
    getAll: (filters?: LessonFilters) => Promise<Lesson[]>,
    getById: (id: string) => Promise<Lesson>,
    create: (data: CreateLessonDto) => Promise<Lesson>,
    update: (id: string, data: UpdateLessonDto) => Promise<Lesson>,
    publish: (id: string) => Promise<Lesson>,
  },
  // ... more endpoints
};

// Mock with setTimeout for async simulation
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));
Generate Realistic Mock Data:

Use faker or manual realistic data
Maintain referential integrity (foreign keys)
Include edge cases (empty folders, very long names, etc.)
Provide at least 50+ users, 20+ lessons, 100+ media assets

7. Form Handling
React Hook Form + Zod Example:
typescriptconst lessonSchema = z.object({
  title: z.string().min(3, 'Title must be at least 3 characters'),
  description: z.string().optional(),
  topic_id: z.string().uuid(),
  level_id: z.string().uuid(),
  is_published: z.boolean().default(false),
});

type LessonFormData = z.infer<typeof lessonSchema>;

function LessonForm() {
  const form = useForm<LessonFormData>({
    resolver: zodResolver(lessonSchema),
    defaultValues: {
      title: '',
      description: '',
      is_published: false,
    },
  });

  const { mutate, isLoading } = useCreateLesson();

  const onSubmit = (data: LessonFormData) => {
    mutate(data);
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        {/* Form fields */}
      </form>
    </Form>
  );
}
8. Additional Features
Search & Filters:

Global search across entities
Advanced filters with multiple conditions
Save filter presets
Clear all filters button

Bulk Actions:

Select all/none checkboxes
Bulk delete with confirmation
Bulk move
Bulk export

Export Capabilities:

Export user list to CSV
Export lesson data
Export quiz results
Export leaderboard

Notifications:

Toast for success/error messages
Badge count for pending actions
Real-time updates (simulate with polling)


📝 Technical Notes

No localStorage - all state managed in memory or with TanStack Query cache
Use TypeScript strict mode
Proper error boundaries around major sections
Accessible components (ARIA labels, keyboard navigation)
Code splitting considerations (keep in single artifact but organize well)
Reusable components in separate sections
Proper loading states and optimistic updates