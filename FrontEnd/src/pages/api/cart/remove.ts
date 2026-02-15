import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";
import { readSession } from "../../../lib/auth/session";

export const prerender = false;

export const GET: APIRoute = async ({ redirect }) => {
  return redirect("/cart?error=Gunakan tombol hapus keranjang");
};

export const POST: APIRoute = async ({ request, cookies, redirect }) => {
  const session = readSession(cookies);
  if (!session) {
    return redirect("/auth/login?next=/cart");
  }

  const form = await request.formData();
  const productID = String(form.get("product_id") || "").trim();
  const quantity = Number(form.get("quantity") || "1");

  const result = await backendRequest<null>("/api/v1/me/cart/remove", {
    method: "POST",
    token: session.token,
    body: {
      product_id: productID,
      quantity,
    },
  });

  if (result.error) {
    return redirect(`/cart?error=${encodeURIComponent(result.error)}`);
  }

  return redirect(`/cart?message=${encodeURIComponent("Item dihapus dari keranjang")}`);
};
