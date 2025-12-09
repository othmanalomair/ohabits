-- Migration: 002_cleanup_unused_tables
-- Date: 2025-12-09
-- Description: Remove unused tables (news, market, finance, sync, etc.)

-- =====================================================
-- حذف الجداول الغير مستخدمة
-- =====================================================

-- News tables (الأخبار)
DROP TABLE IF EXISTS public.news_articles CASCADE;
DROP TABLE IF EXISTS public.news_sources CASCADE;
DROP TABLE IF EXISTS public.user_news_preferences CASCADE;
DROP TABLE IF EXISTS public.user_interests CASCADE;

-- Market tables (السوق)
DROP TABLE IF EXISTS public.market_data CASCADE;
DROP TABLE IF EXISTS public.market_watchlist CASCADE;

-- Finance tables (المالية)
DROP TABLE IF EXISTS public.finance_analytics CASCADE;
DROP TABLE IF EXISTS public.payment_logs CASCADE;
DROP TABLE IF EXISTS public.recurring_payments CASCADE;
DROP TABLE IF EXISTS public.savings_contributions CASCADE;
DROP TABLE IF EXISTS public.savings_goals CASCADE;
DROP TABLE IF EXISTS public.user_finances CASCADE;

-- Sync tables (المزامنة)
DROP TABLE IF EXISTS public.sync_logs CASCADE;
DROP TABLE IF EXISTS public.sync_settings CASCADE;

-- Notifications (الإشعارات)
DROP TABLE IF EXISTS public.notifications CASCADE;

-- Unused functions
DROP FUNCTION IF EXISTS public.update_user_news_preferences_updated_at() CASCADE;
DROP FUNCTION IF EXISTS public.update_finance_updated_at_column() CASCADE;

-- =====================================================
-- تنظيف الـ sequences المتبقية
-- =====================================================
DROP SEQUENCE IF EXISTS public.sync_settings_id_seq CASCADE;
