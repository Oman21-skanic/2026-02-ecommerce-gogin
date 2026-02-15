import type { APIRoute } from "astro";
import { COOKIE_EMAIL, COOKIE_ROLE, COOKIE_TOKEN, cookieOptions } from "../../../lib/auth/session";

export const prerender = false;

export const GET: APIRoute = async ({ url, cookies, redirect }) => {
  const token = url.searchParams.get("token") || "";
  const email = url.searchParams.get("email") || "";
  const role = url.searchParams.get("role") || "user";
  const next = url.searchParams.get("next") || "/products";

  if (!token || !email) {
    return redirect(`/auth/login?error=${encodeURIComponent("Google login gagal")}`);
  }

  const options = cookieOptions();
  cookies.set(COOKIE_TOKEN, token, options);
  cookies.set(COOKIE_ROLE, role, options);
  cookies.set(COOKIE_EMAIL, email, options);

  return redirect(next || "/products");
};
