// غيّر هذا الرقم مع كل تحديث مهم
// major.minor.patch - مثال: 1.0.0 → 1.0.1 (fix) → 1.1.0 (feature) → 2.0.0 (breaking)
const CACHE_VERSION = '1.2.1';
const CACHE_NAME = `ohabits-v${CACHE_VERSION}`;
const STATIC_CACHE = `ohabits-static-v${CACHE_VERSION}`;
const DYNAMIC_CACHE = `ohabits-dynamic-v${CACHE_VERSION}`;

// الملفات الأساسية للتخزين المؤقت
const STATIC_ASSETS = [
  '/',
  '/static/css/app.css',
  '/static/js/htmx.min.js',
  '/static/js/alpine.min.js',
  '/static/images/logo.png',
  '/static/manifest.json'
];

// تثبيت Service Worker
self.addEventListener('install', (event) => {
  console.log('[SW] Installing...');
  event.waitUntil(
    caches.open(STATIC_CACHE)
      .then((cache) => {
        console.log('[SW] Caching static assets');
        return cache.addAll(STATIC_ASSETS);
      })
      .then(() => self.skipWaiting())
  );
});

// تفعيل Service Worker وحذف الكاش القديم
self.addEventListener('activate', (event) => {
  console.log('[SW] Activating...');
  event.waitUntil(
    caches.keys().then((keys) => {
      return Promise.all(
        keys
          .filter((key) => key !== STATIC_CACHE && key !== DYNAMIC_CACHE)
          .map((key) => {
            console.log('[SW] Deleting old cache:', key);
            return caches.delete(key);
          })
      );
    }).then(() => self.clients.claim())
  );
});

// استراتيجية الشبكة أولاً مع fallback للكاش
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // تجاهل الطلبات غير GET
  if (request.method !== 'GET') {
    return;
  }

  // تجاهل طلبات API - نريدها دائماً من الشبكة
  if (url.pathname.startsWith('/api/')) {
    return;
  }

  // للملفات الثابتة: الكاش أولاً
  if (url.pathname.startsWith('/static/')) {
    event.respondWith(
      caches.match(request)
        .then((cached) => {
          if (cached) {
            return cached;
          }
          return fetch(request).then((response) => {
            if (!response || response.status !== 200) {
              return response;
            }
            const responseClone = response.clone();
            caches.open(STATIC_CACHE).then((cache) => {
              cache.put(request, responseClone);
            });
            return response;
          });
        })
    );
    return;
  }

  // للصفحات: الشبكة أولاً مع fallback للكاش
  event.respondWith(
    fetch(request)
      .then((response) => {
        if (!response || response.status !== 200 || response.type !== 'basic') {
          return response;
        }
        const responseClone = response.clone();
        caches.open(DYNAMIC_CACHE).then((cache) => {
          cache.put(request, responseClone);
        });
        return response;
      })
      .catch(() => {
        return caches.match(request).then((cached) => {
          if (cached) {
            return cached;
          }
          // صفحة offline افتراضية
          if (request.headers.get('accept').includes('text/html')) {
            return caches.match('/');
          }
        });
      })
  );
});

// معالجة رسائل من الصفحة
self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
});
