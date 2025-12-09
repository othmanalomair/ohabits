--
-- PostgreSQL database dump
--

\restrict WpYiFOWd1wDINmnHbZGvShh53Eg0nW7IpBOkEEqkbgXZkoZhR3xE8QxasdjMyMs

-- Dumped from database version 16.10 (Debian 16.10-1.pgdg13+1)
-- Dumped by pg_dump version 16.10 (Debian 16.10-1.pgdg13+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: daily_images; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.daily_images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    date date NOT NULL,
    original_path text NOT NULL,
    thumbnail_path text NOT NULL,
    filename text NOT NULL,
    mime_type text NOT NULL,
    size_bytes integer NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.daily_images OWNER TO postgres;

--
-- Name: episode_tracking; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.episode_tracking (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    episode_id uuid,
    user_id uuid,
    watched boolean DEFAULT false NOT NULL,
    rating integer,
    notes text,
    watched_at timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT episode_tracking_rating_check CHECK (((rating >= 1) AND (rating <= 10)))
);


ALTER TABLE public.episode_tracking OWNER TO most3mr;

--
-- Name: episodes; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.episodes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    show_id uuid,
    user_id uuid,
    external_id integer NOT NULL,
    name text NOT NULL,
    season integer NOT NULL,
    number integer NOT NULL,
    summary text,
    airdate date,
    runtime integer,
    image_url text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    show_type character varying(20) DEFAULT 'tv'::character varying NOT NULL,
    filler boolean DEFAULT false NOT NULL,
    recap boolean DEFAULT false NOT NULL,
    CONSTRAINT episodes_type_check CHECK (((show_type)::text = ANY (ARRAY[('tv'::character varying)::text, ('anime'::character varying)::text])))
);


ALTER TABLE public.episodes OWNER TO most3mr;

--
-- Name: habits; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.habits (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    name text NOT NULL,
    scheduled_days jsonb NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.habits OWNER TO most3mr;

--
-- Name: habits_completions; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.habits_completions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    habit_id uuid,
    user_id uuid,
    completed boolean NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.habits_completions OWNER TO most3mr;

--
-- Name: markdown_notes; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.markdown_notes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    title text NOT NULL,
    content text DEFAULT ''::text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    is_rtl boolean DEFAULT false NOT NULL
);


ALTER TABLE public.markdown_notes OWNER TO most3mr;

--
-- Name: COLUMN markdown_notes.is_rtl; Type: COMMENT; Schema: public; Owner: most3mr
--

COMMENT ON COLUMN public.markdown_notes.is_rtl IS 'Indicates text direction: true for RTL (Arabic), false for LTR (English)';


--
-- Name: medication_logs; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.medication_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    medication_id uuid,
    user_id uuid,
    taken boolean NOT NULL,
    scheduled_time text,
    actual_time timestamp without time zone,
    date date NOT NULL,
    notes text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.medication_logs OWNER TO most3mr;

--
-- Name: medications; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.medications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
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


ALTER TABLE public.medications OWNER TO most3mr;

--
-- Name: mood_ratings; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.mood_ratings (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    rating integer NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    CONSTRAINT mood_ratings_rating_check CHECK (((rating >= 1) AND (rating <= 10)))
);


ALTER TABLE public.mood_ratings OWNER TO most3mr;

--
-- Name: notes; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.notes (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    date date NOT NULL,
    text text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.notes OWNER TO most3mr;

--
-- Name: projects; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.projects (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.projects OWNER TO most3mr;

--
-- Name: shows; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.shows (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
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
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    last_episode_sync timestamp without time zone,
    last_info_sync timestamp without time zone,
    show_type character varying(20) DEFAULT 'tv'::character varying NOT NULL,
    CONSTRAINT shows_type_check CHECK (((show_type)::text = ANY (ARRAY[('tv'::character varying)::text, ('anime'::character varying)::text])))
);


ALTER TABLE public.shows OWNER TO most3mr;

--
-- Name: task_attachments; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.task_attachments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    filename text NOT NULL,
    file_path text NOT NULL,
    file_size bigint NOT NULL,
    mime_type text NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.task_attachments OWNER TO most3mr;

--
-- Name: task_comments; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.task_comments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    task_id uuid NOT NULL,
    user_id uuid NOT NULL,
    comment text NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.task_comments OWNER TO most3mr;

--
-- Name: task_dependencies; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.task_dependencies (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    task_id uuid NOT NULL,
    depends_on_task_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT task_dependencies_no_self_reference CHECK ((task_id <> depends_on_task_id))
);


ALTER TABLE public.task_dependencies OWNER TO most3mr;

--
-- Name: tasks; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.tasks (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    project_id uuid NOT NULL,
    parent_task_id uuid,
    title text NOT NULL,
    description text,
    status text DEFAULT 'Not Started'::text NOT NULL,
    priority text DEFAULT 'None'::text NOT NULL,
    due_date timestamp without time zone,
    completed boolean DEFAULT false NOT NULL,
    display_order integer DEFAULT 0,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    attachment_filename text,
    attachment_file_path text,
    attachment_content_type text,
    attachment_file_size bigint,
    collapsed boolean DEFAULT false NOT NULL,
    CONSTRAINT tasks_priority_check CHECK ((priority = ANY (ARRAY['None'::text, 'Low'::text, 'Medium'::text, 'High'::text]))),
    CONSTRAINT tasks_status_check CHECK ((status = ANY (ARRAY['Not Started'::text, 'In Progress'::text, 'Blocked'::text, 'In Review'::text, 'Completed'::text])))
);


ALTER TABLE public.tasks OWNER TO most3mr;

--
-- Name: todos; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.todos (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    text text NOT NULL,
    completed boolean NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.todos OWNER TO most3mr;

--
-- Name: users; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    password text NOT NULL,
    display_name text NOT NULL,
    avatar_url text,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    email character varying(255) NOT NULL
);


ALTER TABLE public.users OWNER TO most3mr;

--
-- Name: workout_logs; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.workout_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    completed_exercises jsonb NOT NULL,
    cardio jsonb NOT NULL,
    weight double precision NOT NULL,
    date date NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    name text DEFAULT ''::text NOT NULL
);


ALTER TABLE public.workout_logs OWNER TO most3mr;

--
-- Name: workouts; Type: TABLE; Schema: public; Owner: most3mr
--

CREATE TABLE public.workouts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid,
    name text NOT NULL,
    day text NOT NULL,
    exercises jsonb NOT NULL,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    display_order integer NOT NULL
);


ALTER TABLE public.workouts OWNER TO most3mr;

--
-- Name: daily_images daily_images_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.daily_images
    ADD CONSTRAINT daily_images_pkey PRIMARY KEY (id);


--
-- Name: episode_tracking episode_tracking_episode_user_unique; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_episode_user_unique UNIQUE (episode_id, user_id);


--
-- Name: episode_tracking episode_tracking_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_pkey PRIMARY KEY (id);


--
-- Name: episodes episodes_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_pkey PRIMARY KEY (id);


--
-- Name: habits_completions habits_completions_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.habits_completions
    ADD CONSTRAINT habits_completions_pkey PRIMARY KEY (id);


--
-- Name: habits habits_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.habits
    ADD CONSTRAINT habits_pkey PRIMARY KEY (id);


--
-- Name: markdown_notes markdown_notes_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.markdown_notes
    ADD CONSTRAINT markdown_notes_pkey PRIMARY KEY (id);


--
-- Name: medication_logs medication_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.medication_logs
    ADD CONSTRAINT medication_logs_pkey PRIMARY KEY (id);


--
-- Name: medications medications_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.medications
    ADD CONSTRAINT medications_pkey PRIMARY KEY (id);


--
-- Name: mood_ratings mood_ratings_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.mood_ratings
    ADD CONSTRAINT mood_ratings_pkey PRIMARY KEY (id);


--
-- Name: mood_ratings mood_ratings_user_date_unique; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.mood_ratings
    ADD CONSTRAINT mood_ratings_user_date_unique UNIQUE (user_id, date);


--
-- Name: notes notes_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.notes
    ADD CONSTRAINT notes_pkey PRIMARY KEY (id);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: shows shows_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.shows
    ADD CONSTRAINT shows_pkey PRIMARY KEY (id);


--
-- Name: task_attachments task_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_pkey PRIMARY KEY (id);


--
-- Name: task_comments task_comments_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_pkey PRIMARY KEY (id);


--
-- Name: task_dependencies task_dependencies_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_dependencies
    ADD CONSTRAINT task_dependencies_pkey PRIMARY KEY (id);


--
-- Name: task_dependencies task_dependencies_unique_pair; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_dependencies
    ADD CONSTRAINT task_dependencies_unique_pair UNIQUE (task_id, depends_on_task_id);


--
-- Name: tasks tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_pkey PRIMARY KEY (id);


--
-- Name: todos todos_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.todos
    ADD CONSTRAINT todos_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: workout_logs workout_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.workout_logs
    ADD CONSTRAINT workout_logs_pkey PRIMARY KEY (id);


--
-- Name: workouts workouts_pkey; Type: CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.workouts
    ADD CONSTRAINT workouts_pkey PRIMARY KEY (id);


--
-- Name: idx_daily_images_user_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_daily_images_user_date ON public.daily_images USING btree (user_id, date);


--
-- Name: idx_episode_tracking_episode_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episode_tracking_episode_id ON public.episode_tracking USING btree (episode_id);


--
-- Name: idx_episode_tracking_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episode_tracking_user_id ON public.episode_tracking USING btree (user_id);


--
-- Name: idx_episodes_external_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_external_id ON public.episodes USING btree (external_id);


--
-- Name: idx_episodes_filler; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_filler ON public.episodes USING btree (filler);


--
-- Name: idx_episodes_filler_recap; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_filler_recap ON public.episodes USING btree (filler, recap);


--
-- Name: idx_episodes_recap; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_recap ON public.episodes USING btree (recap);


--
-- Name: idx_episodes_show_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_episodes_show_id ON public.episodes USING btree (show_id);


--
-- Name: idx_markdown_notes_created_at; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_markdown_notes_created_at ON public.markdown_notes USING btree (created_at DESC);


--
-- Name: idx_markdown_notes_search; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_markdown_notes_search ON public.markdown_notes USING gin (to_tsvector('english'::regconfig, ((title || ' '::text) || content)));


--
-- Name: idx_markdown_notes_title; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_markdown_notes_title ON public.markdown_notes USING btree (title);


--
-- Name: idx_markdown_notes_updated_at; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_markdown_notes_updated_at ON public.markdown_notes USING btree (updated_at DESC);


--
-- Name: idx_markdown_notes_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_markdown_notes_user_id ON public.markdown_notes USING btree (user_id);


--
-- Name: idx_medication_logs_date; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_medication_logs_date ON public.medication_logs USING btree (date);


--
-- Name: idx_medication_logs_medication_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_medication_logs_medication_id ON public.medication_logs USING btree (medication_id);


--
-- Name: idx_medication_logs_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_medication_logs_user_id ON public.medication_logs USING btree (user_id);


--
-- Name: idx_medications_is_active; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_medications_is_active ON public.medications USING btree (is_active);


--
-- Name: idx_medications_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_medications_user_id ON public.medications USING btree (user_id);


--
-- Name: idx_projects_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_projects_user_id ON public.projects USING btree (user_id);


--
-- Name: idx_shows_external_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_external_id ON public.shows USING btree (external_id);


--
-- Name: idx_shows_external_id_type; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_external_id_type ON public.shows USING btree (external_id, show_type);


--
-- Name: idx_shows_last_episode_sync; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_last_episode_sync ON public.shows USING btree (last_episode_sync);


--
-- Name: idx_shows_last_info_sync; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_last_info_sync ON public.shows USING btree (last_info_sync);


--
-- Name: idx_shows_status; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_status ON public.shows USING btree (status);


--
-- Name: idx_shows_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_shows_user_id ON public.shows USING btree (user_id);


--
-- Name: idx_task_attachments_task_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_task_attachments_task_id ON public.task_attachments USING btree (task_id);


--
-- Name: idx_task_comments_task_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_task_comments_task_id ON public.task_comments USING btree (task_id);


--
-- Name: idx_task_dependencies_depends_on_task_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_task_dependencies_depends_on_task_id ON public.task_dependencies USING btree (depends_on_task_id);


--
-- Name: idx_task_dependencies_task_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_task_dependencies_task_id ON public.task_dependencies USING btree (task_id);


--
-- Name: idx_tasks_completed; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_completed ON public.tasks USING btree (completed);


--
-- Name: idx_tasks_due_date; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_due_date ON public.tasks USING btree (due_date);


--
-- Name: idx_tasks_parent_display_order; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_parent_display_order ON public.tasks USING btree (parent_task_id, display_order) WHERE (parent_task_id IS NOT NULL);


--
-- Name: idx_tasks_parent_task_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_parent_task_id ON public.tasks USING btree (parent_task_id);


--
-- Name: idx_tasks_priority; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_priority ON public.tasks USING btree (priority);


--
-- Name: idx_tasks_project_display_order; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_project_display_order ON public.tasks USING btree (project_id, display_order);


--
-- Name: idx_tasks_project_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_project_id ON public.tasks USING btree (project_id);


--
-- Name: idx_tasks_status; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_status ON public.tasks USING btree (status);


--
-- Name: idx_tasks_user_id; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_user_id ON public.tasks USING btree (user_id);


--
-- Name: idx_tasks_with_attachments; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_tasks_with_attachments ON public.tasks USING btree (id) WHERE (attachment_filename IS NOT NULL);


--
-- Name: idx_workouts_user_display_order; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE INDEX idx_workouts_user_display_order ON public.workouts USING btree (user_id, display_order);


--
-- Name: notes_user_date_idx; Type: INDEX; Schema: public; Owner: most3mr
--

CREATE UNIQUE INDEX notes_user_date_idx ON public.notes USING btree (user_id, date);


--
-- Name: daily_images daily_images_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.daily_images
    ADD CONSTRAINT daily_images_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: episode_tracking episode_tracking_episode_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_episode_id_fkey FOREIGN KEY (episode_id) REFERENCES public.episodes(id) ON DELETE CASCADE;


--
-- Name: episode_tracking episode_tracking_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episode_tracking
    ADD CONSTRAINT episode_tracking_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: episodes episodes_show_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_show_id_fkey FOREIGN KEY (show_id) REFERENCES public.shows(id) ON DELETE CASCADE;


--
-- Name: episodes episodes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.episodes
    ADD CONSTRAINT episodes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: habits_completions habits_completions_habit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.habits_completions
    ADD CONSTRAINT habits_completions_habit_id_fkey FOREIGN KEY (habit_id) REFERENCES public.habits(id) ON DELETE CASCADE;


--
-- Name: habits_completions habits_completions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.habits_completions
    ADD CONSTRAINT habits_completions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: habits habits_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.habits
    ADD CONSTRAINT habits_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: markdown_notes markdown_notes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.markdown_notes
    ADD CONSTRAINT markdown_notes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: medication_logs medication_logs_medication_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.medication_logs
    ADD CONSTRAINT medication_logs_medication_id_fkey FOREIGN KEY (medication_id) REFERENCES public.medications(id) ON DELETE CASCADE;


--
-- Name: medication_logs medication_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.medication_logs
    ADD CONSTRAINT medication_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: medications medications_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.medications
    ADD CONSTRAINT medications_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: mood_ratings mood_ratings_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.mood_ratings
    ADD CONSTRAINT mood_ratings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: notes notes_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.notes
    ADD CONSTRAINT notes_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: projects projects_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: shows shows_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.shows
    ADD CONSTRAINT shows_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: task_attachments task_attachments_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_attachments task_attachments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_attachments
    ADD CONSTRAINT task_attachments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: task_comments task_comments_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_comments task_comments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_comments
    ADD CONSTRAINT task_comments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: task_dependencies task_dependencies_depends_on_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_dependencies
    ADD CONSTRAINT task_dependencies_depends_on_task_id_fkey FOREIGN KEY (depends_on_task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: task_dependencies task_dependencies_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.task_dependencies
    ADD CONSTRAINT task_dependencies_task_id_fkey FOREIGN KEY (task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: tasks tasks_parent_task_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_parent_task_id_fkey FOREIGN KEY (parent_task_id) REFERENCES public.tasks(id) ON DELETE CASCADE;


--
-- Name: tasks tasks_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: tasks tasks_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.tasks
    ADD CONSTRAINT tasks_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: todos todos_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.todos
    ADD CONSTRAINT todos_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: workout_logs workout_logs_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.workout_logs
    ADD CONSTRAINT workout_logs_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: workouts workouts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: most3mr
--

ALTER TABLE ONLY public.workouts
    ADD CONSTRAINT workouts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict WpYiFOWd1wDINmnHbZGvShh53Eg0nW7IpBOkEEqkbgXZkoZhR3xE8QxasdjMyMs

