# ohabits - دليل المشروع

## نظرة عامة
تطبيق لتتبع العادات اليومية والأدوية والتمارين والمزاج - بتصميم Retro Arabic Anime

## متطلبات التصميم

### اللغة والاتجاه
- **عربي بالكامل** - كل النصوص والواجهة بالعربي
- **RTL (Right-to-Left)** - اتجاه من اليمين لليسار في كل مكان
- `dir="rtl"` و `lang="ar"` في كل الصفحات

### التوافق مع الأجهزة (Mobile Responsive)
- **الجهاز المستهدف**: Samsung Galaxy Z Fold 7
- **الشاشة الخارجية**: ضيقة جداً (~3.4 inch) - عمود واحد
- **الشاشة الداخلية**: كبيرة (~7.6 inch) - عمودين أو أكثر
- **Desktop**: عمودين مع max-width

### نظام الألوان (Retro Orange Theme)
```
Primary (Orange):
  500: #F97316 (الرئيسي)
  600: #EA580C
  700: #C2410C
  800: #9A3412
  900: #7C2D12

Cream (الخلفية):
  50: #FFFBF5
  100: #FFF5E6
  200: #FFECD1

Retro:
  brown: #8B4513
  sepia: #704214
  dark: #3D2314

Accent:
  pink: #FF6B9D
  teal: #20B2AA
  gold: #FFD700
  coral: #FF7F7F
```

### الخطوط
- **الخط الرئيسي**: Tajawal (Google Fonts)
- **الأوزان**: 400, 500, 700, 800

### عناصر التصميم
- **الكروت**: `retro-card` - حدود برتقالية، ظل 6px
- **الأزرار**: `anime-btn` - تدرج برتقالي، ظل، تأثير عند الضغط
- **المدخلات**: `retro-input` - خلفية كريمية، حدود برتقالية فاتحة
- **Checkboxes**: `retro-checkbox` - مربعة مع أيقونة صح خضراء
- **Badges**: `retro-badge` - تدرج برتقالي فاتح

## هيكل الصفحات

### الصفحة الرئيسية (Dashboard)
1. **Header** - شعار + قائمة + صورة المستخدم
2. **التقويم** - اختيار التاريخ (اليوم افتراضياً)
3. **عادات اليوم** - قائمة العادات مع checkboxes
4. **الأدوية** - قائمة الأدوية مع زر "خذها"
5. **مهام اليوم** - todos مع إضافة سريعة
6. **المذكرة اليومية** - textarea للملاحظات
7. **المزاج** - 5 emojis للاختيار
8. **تمرين اليوم** - اختيار التمرين وتسجيل الوزن/الكارديو
9. **Bottom Navigation** - للموبايل فقط

## التوقيت
- **المنطقة الزمنية**: Asia/Kuwait (UTC+3)
- كل التواريخ والأوقات تُعرض بتوقيت الكويت

## Tech Stack
- **Backend**: Go + Echo v4
- **Templates**: Templ (type-safe)
- **Styling**: Tailwind CSS
- **Interactivity**: HTMX + Alpine.js
- **Database**: PostgreSQL
- **Auth**: JWT + bcrypt

## أوامر التطوير
```bash
make dev          # تشغيل مع hot reload
make build        # بناء للإنتاج
make templ        # توليد الـ templates
make css          # بناء Tailwind
make db-reset     # إعادة تعيين قاعدة البيانات
```

## ملف التصميم المرجعي
`design-reference.html` - يحتوي على كل عناصر التصميم والألوان والستايلات
