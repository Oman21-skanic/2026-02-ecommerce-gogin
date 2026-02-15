import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";

export const prerender = false;

export const GET: APIRoute = async ({ url, redirect }) => {
  const token = url.searchParams.get("token") || "";
  const result = await backendRequest<{ message: string }>(`/api/v1/auth/verify-email?token=${encodeURIComponent(token)}`);

  if (!result.data) {
    return redirect(`/auth/login?error=${encodeURIComponent(result.error || "Verifikasi email gagal")}`);
  }

  return redirect(`/auth/login?message=${encodeURIComponent("Email berhasil diverifikasi")}`);
};
