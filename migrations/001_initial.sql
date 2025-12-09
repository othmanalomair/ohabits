-- ohabits Database Schema
-- Version: 1.0 (migrated from old ohabits)
-- Arabic RTL Habit Tracking App

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- المستخدمين (Users)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    email character varying(255) NOT NULL UNIQUE,
    password text NOT NULL,
    display_name text NOT NULL,
    avatar_url text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- =====================================================
-- العادات (Habits)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.habits (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    name text NOT NULL,
    scheduled_days jsonb NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.habits_completions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    habit_id uuid REFERENCES public.habits(id) ON DELETE CASCADE,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    completed boolean NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- =====================================================
-- الأدوية (Medications)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.medications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    name text NOT NULL,
    dosage text NOT NULL,
    scheduled_days jsonb NOT NULL,
    times_per_day integer DEFAULT 1 NOT NULL,
    time_intervals text[],
    duration_type text NOT NULL,
    start_date date NOT NULL,
    end_date date,
    notes text,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT medications_duration_type_check CHECK ((duration_type = ANY (ARRAY['lifetime'::text, 'limited'::text])))
);

CREATE TABLE IF NOT EXISTS public.medication_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    medication_id uuid REFERENCES public.medications(id) ON DELETE CASCADE,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    taken boolean NOT NULL,
    scheduled_time text,
    actual_time timestamp without time zone,
    date date NOT NULL,
    notes text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- =====================================================
-- المهام اليومية (Todos)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.todos (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    text text NOT NULL,
    completed boolean NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- =====================================================
-- الملاحظات السريعة (Quick Notes)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.notes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    date date NOT NULL,
    text text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- =====================================================
-- تقييم المزاج (Mood Ratings) - 1 to 10
-- =====================================================
CREATE TABLE IF NOT EXISTS public.mood_ratings (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    rating integer NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT mood_ratings_rating_check CHECK (((rating >= 1) AND (rating <= 10)))
);

-- =====================================================
-- الملاحظات الطويلة (Markdown Notes)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.markdown_notes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    title text NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    is_rtl boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

COMMENT ON COLUMN public.markdown_notes.is_rtl IS 'Indicates text direction: true for RTL (Arabic), false for LTR (English)';

-- =====================================================
-- خطط التمارين (Workout Plans)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.workouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    name text NOT NULL,
    day text NOT NULL,
    exercises jsonb NOT NULL,
    display_order integer DEFAULT 0,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- =====================================================
-- سجلات التمارين (Workout Logs)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.workout_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    name text DEFAULT ''::text NOT NULL,
    completed_exercises jsonb NOT NULL,
    cardio jsonb NOT NULL,
    weight double precision NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- =====================================================
-- المسلسلات (Shows)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.shows (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    external_id integer NOT NULL,
    name text NOT NULL,
    summary text,
    image_url text,
    status text,
    premiered date,
    ended date,
    network text,
    genres jsonb,
    rating jsonb,
    show_type character varying(20) DEFAULT 'tv'::character varying NOT NULL,
    last_episode_sync timestamp without time zone,
    last_info_sync timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT shows_type_check CHECK (((show_type)::text = ANY ((ARRAY['tv'::character varying, 'anime'::character varying])::text[])))
);

-- =====================================================
-- الحلقات (Episodes)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.episodes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    show_id uuid REFERENCES public.shows(id) ON DELETE CASCADE,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    external_id integer NOT NULL,
    name text NOT NULL,
    season integer NOT NULL,
    number integer NOT NULL,
    summary text,
    airdate date,
    runtime integer,
    image_url text,
    show_type character varying(20) DEFAULT 'tv'::character varying NOT NULL,
    filler boolean DEFAULT false NOT NULL,
    recap boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT episodes_type_check CHECK (((show_type)::text = ANY ((ARRAY['tv'::character varying, 'anime'::character varying])::text[])))
);

-- =====================================================
-- تتبع الحلقات (Episode Tracking)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.episode_tracking (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    episode_id uuid REFERENCES public.episodes(id) ON DELETE CASCADE,
    user_id uuid REFERENCES public.users(id) ON DELETE CASCADE,
    watched boolean DEFAULT false NOT NULL,
    rating integer,
    notes text,
    watched_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT episode_tracking_rating_check CHECK (((rating >= 1) AND (rating <= 10))),
    CONSTRAINT episode_tracking_episode_user_unique UNIQUE (episode_id, user_id)
);

-- =====================================================
-- المشاريع (Projects)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.projects (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- =====================================================
-- المهام (Tasks)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.tasks (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    project_id uuid NOT NULL REFERENCES public.projects(id) ON DELETE CASCADE,
    parent_task_id uuid REFERENCES public.tasks(id) ON DELETE CASCADE,
    title text NOT NULL,
    description text,
    status text DEFAULT 'Not Started'::text NOT NULL,
    priority text DEFAULT 'None'::text NOT NULL,
    due_date timestamp without time zone,
    completed boolean DEFAULT false NOT NULL,
    collapsed boolean DEFAULT false NOT NULL,
    display_order integer DEFAULT 0,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT tasks_priority_check CHECK ((priority = ANY (ARRAY['None'::text, 'Low'::text, 'Medium'::text, 'High'::text]))),
    CONSTRAINT tasks_status_check CHECK ((status = ANY (ARRAY['Not Started'::text, 'In Progress'::text, 'Blocked'::text, 'In Review'::text, 'Completed'::text])))
);

-- =====================================================
-- تعليقات المهام (Task Comments)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.task_comments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    task_id uuid NOT NULL REFERENCES public.tasks(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    comment text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

-- =====================================================
-- مرفقات المهام (Task Attachments)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.task_attachments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    task_id uuid NOT NULL REFERENCES public.tasks(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    filename text NOT NULL,
    file_path text NOT NULL,
    file_size bigint NOT NULL,
    mime_type text NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);

-- =====================================================
-- اعتمادات المهام (Task Dependencies)
-- =====================================================
CREATE TABLE IF NOT EXISTS public.task_dependencies (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL PRIMARY KEY,
    task_id uuid NOT NULL REFERENCES public.tasks(id) ON DELETE CASCADE,
    depends_on_task_id uuid NOT NULL REFERENCES public.tasks(id) ON DELETE CASCADE,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT task_dependencies_no_self_reference CHECK ((task_id <> depends_on_task_id)),
    CONSTRAINT task_dependencies_unique_pair UNIQUE (task_id, depends_on_task_id)
);

-- =====================================================
-- Indexes
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_habits_user_id ON public.habits(user_id);
CREATE INDEX IF NOT EXISTS idx_habits_completions_habit_id ON public.habits_completions(habit_id);
CREATE INDEX IF NOT EXISTS idx_habits_completions_user_id ON public.habits_completions(user_id);
CREATE INDEX IF NOT EXISTS idx_habits_completions_date ON public.habits_completions(date);

CREATE INDEX IF NOT EXISTS idx_medications_user_id ON public.medications(user_id);
CREATE INDEX IF NOT EXISTS idx_medication_logs_medication_id ON public.medication_logs(medication_id);
CREATE INDEX IF NOT EXISTS idx_medication_logs_user_id ON public.medication_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_medication_logs_date ON public.medication_logs(date);

CREATE INDEX IF NOT EXISTS idx_todos_user_id ON public.todos(user_id);
CREATE INDEX IF NOT EXISTS idx_todos_date ON public.todos(date);

CREATE INDEX IF NOT EXISTS idx_notes_user_id ON public.notes(user_id);
CREATE INDEX IF NOT EXISTS idx_notes_date ON public.notes(date);

CREATE INDEX IF NOT EXISTS idx_mood_ratings_user_id ON public.mood_ratings(user_id);
CREATE INDEX IF NOT EXISTS idx_mood_ratings_date ON public.mood_ratings(date);

CREATE INDEX IF NOT EXISTS idx_markdown_notes_user_id ON public.markdown_notes(user_id);

CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON public.workouts(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_logs_user_id ON public.workout_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_workout_logs_date ON public.workout_logs(date);

CREATE INDEX IF NOT EXISTS idx_shows_user_id ON public.shows(user_id);
CREATE INDEX IF NOT EXISTS idx_episodes_show_id ON public.episodes(show_id);
CREATE INDEX IF NOT EXISTS idx_episodes_external_id ON public.episodes(external_id);
CREATE INDEX IF NOT EXISTS idx_episodes_filler ON public.episodes(filler);
CREATE INDEX IF NOT EXISTS idx_episodes_recap ON public.episodes(recap);
CREATE INDEX IF NOT EXISTS idx_episodes_filler_recap ON public.episodes(filler, recap);
CREATE INDEX IF NOT EXISTS idx_episode_tracking_episode_id ON public.episode_tracking(episode_id);
CREATE INDEX IF NOT EXISTS idx_episode_tracking_user_id ON public.episode_tracking(user_id);

CREATE INDEX IF NOT EXISTS idx_projects_user_id ON public.projects(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON public.tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON public.tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_parent_task_id ON public.tasks(parent_task_id);
