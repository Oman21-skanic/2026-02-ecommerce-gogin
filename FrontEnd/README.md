# Mancafe Frontend

Frontend e-commerce Mancafe menggunakan Astro + Tailwind dengan arsitektur SEO-friendly dan BFF auth via HTTP-only cookie.

## Stack
- Astro (SSR mode with Node adapter)
- Tailwind CSS
- React Islands (untuk komponen interaktif jika dibutuhkan)

## Setup
1. Copy `.env.example` menjadi `.env`
2. Sesuaikan `API_BASE_URL` ke backend Go
3. Install dependency: `npm install`
4. Jalankan dev server: `npm run dev`

## Scripts
- `npm run dev` → local development
- `npm run typecheck` → cek Astro/TS
- `npm run build` → production build
- `npm run preview` → preview hasil build

## Struktur Utama
- `src/pages` → halaman public, user, admin
- `src/pages/api` → BFF endpoint (auth/cart/checkout/admin)
- `src/lib` → helper env, session, backend request
- `src/components` → layout, seo, ui components
- `src/middleware.ts` → proteksi route user/admin

## Alur Auth
- Login via `POST /api/auth/login`
- Token backend disimpan di HTTP-only cookie (`mancafe_token`)
- Route sensitif (`/cart`, `/checkout`, `/orders`, `/admin`) diproteksi middleware
