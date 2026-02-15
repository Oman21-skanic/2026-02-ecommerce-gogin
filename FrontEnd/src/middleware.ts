import { defineMiddleware } from "astro:middleware";
import { COOKIE_ROLE, COOKIE_TOKEN } from "./lib/auth/session";

const userPathPrefixes = ["/cart", "/checkout", "/orders", "/api/cart", "/api/checkout"];
const adminPathPrefixes = ["/admin", "/api/admin"];

export const onRequest = defineMiddleware(async ({ url, cookies, redirect }, next) => {
  const pathname = url.pathname;
  const token = cookies.get(COOKIE_TOKEN)?.value;
  const role = cookies.get(COOKIE_ROLE)?.value;

  const needsUser = userPathPrefixes.some((prefix) => pathname.startsWith(prefix));
  const needsAdmin = adminPathPrefixes.some((prefix) => pathname.startsWith(prefix));

  if ((needsUser || needsAdmin) && !token) {
    return redirect(`/auth/login?next=${encodeURIComponent(pathname)}`);
  }

  if (needsAdmin && role !== "admin") {
    return redirect("/auth/login?error=admin");
  }

  return next();
});
