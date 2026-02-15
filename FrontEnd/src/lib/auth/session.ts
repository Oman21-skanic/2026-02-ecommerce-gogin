import type { AstroCookies } from "astro";

export type Session = {
  token: string;
  role: string;
  email: string;
};

export const COOKIE_TOKEN = "mancafe_token";
export const COOKIE_ROLE = "mancafe_role";
export const COOKIE_EMAIL = "mancafe_email";

export function readSession(cookies: AstroCookies): Session | null {
  const token = cookies.get(COOKIE_TOKEN)?.value;
  if (!token) {
    return null;
  }

  return {
    token,
    role: cookies.get(COOKIE_ROLE)?.value || "user",
    email: cookies.get(COOKIE_EMAIL)?.value || "",
  };
}

export function cookieOptions() {
  return {
    path: "/",
    httpOnly: true,
    sameSite: "lax" as const,
    secure: import.meta.env.PROD,
    maxAge: 60 * 60 * 24,
  };
}
