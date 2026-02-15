import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";
import { COOKIE_EMAIL, COOKIE_ROLE, COOKIE_TOKEN, cookieOptions } from "../../../lib/auth/session";

export const prerender = false;

export const POST: APIRoute = async ({ request, cookies, redirect }) => {
  const form = await request.formData();
  const email = String(form.get("email") || "").trim();
  const password = String(form.get("password") || "").trim();
  const next = String(form.get("next") || "/products");

  const result = await backendRequest<{ token: string; role: string; email: string }>("/api/v1/auth/login", {
    method: "POST",
    body: { email, password },
  });

  if (!result.data) {
    return redirect(`/auth/login?error=${encodeURIComponent(result.error || "Login gagal")}`);
  }

  const options = cookieOptions();
  cookies.set(COOKIE_TOKEN, result.data.token, options);
  cookies.set(COOKIE_ROLE, result.data.role, options);
  cookies.set(COOKIE_EMAIL, result.data.email, options);

  return redirect(next || "/products");
};
