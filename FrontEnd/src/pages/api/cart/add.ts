import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";
import { readSession } from "../../../lib/auth/session";

export const prerender = false;

export const GET: APIRoute = async ({ redirect }) => {
  return redirect("/cart?error=Gunakan tombol tambah keranjang");
};

export const POST: APIRoute = async ({ request, cookies, redirect }) => {
  const session = readSession(cookies);
  if (!session) {
    return redirect("/auth/login?next=/cart");
  }

  const form = await request.formData();
  const productID = String(form.get("product_id") || "").trim();
  const quantity = Number(form.get("quantity") || "1");
  const redirectTo = String(form.get("redirect_to") || "/cart");

  const result = await backendRequest<null>("/api/v1/me/cart/add", {
    method: "POST",
    token: session.token,
    body: {
      product_id: productID,
      quantity,
    },
  });

  if (result.error) {
    return redirect(`${redirectTo}${redirectTo.includes("?") ? "&" : "?"}error=${encodeURIComponent(result.error)}`);
  }

  return redirect(`${redirectTo}${redirectTo.includes("?") ? "&" : "?"}message=${encodeURIComponent("Produk ditambahkan ke keranjang")}`);
};
