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

  const priceCents = Number(form.get("price_cents") || "0");

  if (!id) {
    return redirect(`/admin/products?error=${encodeURIComponent("ID produk wajib")}`);
  }

  if (!Number.isFinite(priceCents) || priceCents <= 0 || priceCents > Number.MAX_SAFE_INTEGER) {
    return redirect(`/admin/products?error=${encodeURIComponent("Harga tidak valid")}`);
  }

  const result = await backendRequest(`/api/v1/admin/products/${id}`, {
    method: "PUT",
    token: session.token,
    body: {
      name: String(form.get("name") || "").trim(),
      description: String(form.get("description") || "").trim(),
      category: String(form.get("category") || "").trim(),
      price_cents: priceCents,
      sku: String(form.get("sku") || "").trim(),
      stock: Number(form.get("stock") || "0"),
      thumbnail: String(form.get("thumbnail") || "").trim(),
    },
  });

  if (result.error) {
    return redirect(`/admin/products?error=${encodeURIComponent(result.error)}`);
  }

  return redirect(`/admin/products?message=${encodeURIComponent("Produk berhasil diupdate")}`);
};
