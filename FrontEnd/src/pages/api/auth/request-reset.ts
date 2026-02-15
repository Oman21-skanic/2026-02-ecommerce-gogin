import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";

export const prerender = false;

export const POST: APIRoute = async ({ request, redirect }) => {
  const form = await request.formData();
  const email = String(form.get("email") || "").trim();

  const result = await backendRequest<{ message: string }>("/api/v1/auth/request-password-reset", {
    method: "POST",
    body: { email },
  });

  if (!result.data) {
    return redirect(`/auth/request-reset?error=${encodeURIComponent(result.error || "Gagal request reset")}`);
  }

  return redirect(`/auth/request-reset?message=${encodeURIComponent("Jika email terdaftar, link reset dikirim")}`);
};
