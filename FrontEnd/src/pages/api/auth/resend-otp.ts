import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";

export const prerender = false;

export const POST: APIRoute = async ({ request, redirect }) => {
  const form = await request.formData();
  const email = String(form.get("email") || "").trim();

  const result = await backendRequest<{ message: string }>("/api/v1/auth/resend-otp", {
    method: "POST",
    body: { email },
  });

  if (!result.data) {
    return redirect(`/auth/register?error=${encodeURIComponent(result.error || "Gagal kirim ulang OTP")}&email=${encodeURIComponent(email)}`);
  }

  return redirect(`/auth/register?message=${encodeURIComponent("OTP baru sudah dikirim ke email Anda")}&email=${encodeURIComponent(email)}`);
};
