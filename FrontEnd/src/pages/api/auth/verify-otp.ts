import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";

export const prerender = false;

export const POST: APIRoute = async ({ request, redirect }) => {
  const form = await request.formData();
  const email = String(form.get("email") || "").trim();
  const otp = String(form.get("otp") || "").trim();

  const result = await backendRequest<{ message: string }>("/api/v1/auth/verify-otp", {
    method: "POST",
    body: { email, otp },
  });

  if (!result.data) {
    return redirect(`/auth/register?error=${encodeURIComponent(result.error || "Verifikasi OTP gagal")}&email=${encodeURIComponent(email)}`);
  }

  return redirect(`/auth/login?message=${encodeURIComponent("Verifikasi berhasil. Silakan login.")}&email=${encodeURIComponent(email)}`);
};
