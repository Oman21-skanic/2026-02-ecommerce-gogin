import type { APIRoute } from "astro";
import { backendRequest } from "../../../../lib/api/backend";
import { readSession } from "../../../../lib/auth/session";

export const prerender = false;

export const POST: APIRoute = async ({ request, cookies, redirect }) => {
  const session = readSession(cookies);
  if (!session || session.role !== "admin") {
    return redirect("/auth/login?error=admin");
  }

  const form = await request.formData();
  const id = String(form.get("id") || "").trim();
  const status = String(form.get("status") || "pending").trim();

  if (!id) {
    return redirect(`/admin/orders?error=${encodeURIComponent("ID pesanan wajib")}`);
  }

  const result = await backendRequest(`/api/v1/admin/orders/${id}/status`, {
    method: "PUT",
    token: session.token,
    body: { status },
  });

  if (result.error) {
    return redirect(`/admin/orders?error=${encodeURIComponent(result.error)}`);
  }

  return redirect(`/admin/orders?message=${encodeURIComponent("Status pesanan diperbarui")}`);
};
