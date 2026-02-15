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

  const result = await backendRequest(`/api/v1/admin/products/${id}`, {
    method: "DELETE",
    token: session.token,
  });

  if (result.error) {
    return redirect(`/admin/products?error=${encodeURIComponent(result.error)}`);
  }

  return redirect(`/admin/products?message=${encodeURIComponent("Produk berhasil dihapus")}`);
};
