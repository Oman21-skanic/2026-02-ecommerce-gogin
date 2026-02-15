import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";

export const prerender = false;

export const POST: APIRoute = async ({ request, redirect }) => {
  const form = await request.formData();
  const fullName = String(form.get("full_name") || "").trim();
  const phone = String(form.get("phone") || "").trim();
  const email = String(form.get("email") || "").trim();
  const password = String(form.get("password") || "").trim();
  const confirmPassword = String(form.get("confirm_password") || "").trim();

  if (password !== confirmPassword) {
    return redirect(`/auth/register?error=${encodeURIComponent("Konfirmasi password tidak cocok")}&email=${encodeURIComponent(email)}`);
  }

  const result = await backendRequest<{ message: string }>("/api/v1/auth/signup", {
    method: "POST",
    body: { full_name: fullName, phone, email, password, confirm_password: confirmPassword },
  });

  if (!result.data) {
    return redirect(`/auth/register?error=${encodeURIComponent(result.error || "Register gagal")}&email=${encodeURIComponent(email)}`);
  }

  return redirect(`/auth/register?message=${encodeURIComponent("Akun berhasil dibuat. Masukkan OTP dari email untuk verifikasi.")}&email=${encodeURIComponent(email)}`);
};
