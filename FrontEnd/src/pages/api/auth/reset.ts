import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";

export const prerender = false;

export const POST: APIRoute = async ({ request, redirect }) => {
  const form = await request.formData();
  const token = String(form.get("token") || "").trim();
  const newPassword = String(form.get("new_password") || "").trim();

  const result = await backendRequest<{ message: string }>("/api/v1/auth/reset-password", {
    method: "POST",
    body: {
      token,
      new_password: newPassword,
    },
  });

  if (!result.data) {
    return redirect(`/auth/reset?token=${encodeURIComponent(token)}&error=${encodeURIComponent(result.error || "Reset password gagal")}`);
  }

  return redirect(`/auth/login?message=${encodeURIComponent("Password berhasil direset")}`);
};
