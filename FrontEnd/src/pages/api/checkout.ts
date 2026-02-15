import type { APIRoute } from "astro";
import { backendRequest } from "../../lib/api/backend";
import { readSession } from "../../lib/auth/session";

export const prerender = false;

export const POST: APIRoute = async ({ request, cookies, redirect }) => {
  const session = readSession(cookies);
  if (!session) {
    return redirect("/auth/login?next=/checkout");
  }

  const form = await request.formData();
  const paymentMethod = String(form.get("payment_method") || "qris").trim();

  const result = await backendRequest<{ payment_url?: string; redirect_url?: string }>("/api/v1/me/checkout", {
    method: "POST",
    token: session.token,
    body: {
      payment_method: paymentMethod,
    },
  });

  if (!result.data) {
    return redirect(`/checkout?error=${encodeURIComponent(result.error || "Checkout gagal")}`);
  }

  const paymentURL = result.data.payment_url || result.data.redirect_url;
  if (paymentURL) {
    return Response.redirect(paymentURL, 302);
  }

  return redirect(`/orders?message=${encodeURIComponent("Checkout berhasil dibuat")}`);
};
