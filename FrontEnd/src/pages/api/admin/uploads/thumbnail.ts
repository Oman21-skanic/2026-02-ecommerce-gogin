import type { APIRoute } from "astro";
import { API_BASE_URL } from "../../../../lib/env";
import { readSession } from "../../../../lib/auth/session";

export const prerender = false;

export const POST: APIRoute = async ({ request, cookies }) => {
  const session = readSession(cookies);
  if (!session || session.role !== "admin") {
    return new Response(JSON.stringify({ error: "Unauthorized" }), { status: 401 });
  }

  const form = await request.formData();
  const file = form.get("file");
  if (!(file instanceof File)) {
    return new Response(JSON.stringify({ error: "file diperlukan" }), { status: 400 });
  }

  const upstream = new FormData();
  upstream.append("file", file);

  const response = await fetch(`${API_BASE_URL}/api/v1/admin/uploads/thumbnail`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${session.token}`,
    },
    body: upstream,
  });

  const payload = await response.json().catch(() => ({}));
  return new Response(JSON.stringify(payload), { status: response.status });
};
