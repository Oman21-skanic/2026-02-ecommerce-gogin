import type { APIRoute } from "astro";
import { COOKIE_EMAIL, COOKIE_ROLE, COOKIE_TOKEN } from "../../../lib/auth/session";

export const prerender = false;

export const POST: APIRoute = async ({ cookies, redirect }) => {
  const clear = { path: "/", maxAge: 0 };
  cookies.set(COOKIE_TOKEN, "", clear);
  cookies.set(COOKIE_ROLE, "", clear);
  cookies.set(COOKIE_EMAIL, "", clear);
  return redirect("/");
};
