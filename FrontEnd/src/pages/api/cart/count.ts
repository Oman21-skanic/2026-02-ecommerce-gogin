import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";
import { readSession } from "../../../lib/auth/session";
import type { Cart } from "../../../lib/types";

export const prerender = false;

export const GET: APIRoute = async ({ cookies }) => {
  const session = readSession(cookies);
  if (!session) {
    return new Response(JSON.stringify({ count: 0 }), {
      status: 200,
      headers: {
        "Content-Type": "application/json",
      },
    });
  }

  const result = await backendRequest<Cart>("/api/v1/me/cart", { token: session.token });
  const count = result.data?.items?.reduce((acc, item) => acc + item.quantity, 0) ?? 0;

  return new Response(JSON.stringify({ count }), {
    status: 200,
    headers: {
      "Content-Type": "application/json",
    },
  });
};
