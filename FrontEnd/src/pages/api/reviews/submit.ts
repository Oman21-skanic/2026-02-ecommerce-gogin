import type { APIRoute } from "astro";
import { backendRequest } from "../../../lib/api/backend";
import { readSession } from "../../../lib/auth/session";

export const prerender = false;

export const POST: APIRoute = async ({ request, cookies, redirect }) => {
	const session = readSession(cookies);
	if (!session) {
		return redirect("/auth/login?next=/reviews");
	}

	const form = await request.formData();
	const rating = Number(form.get("rating") || "5");
	const comment = String(form.get("comment") || "").trim();

	if (!comment) {
		return redirect("/reviews?error=Komentar tidak boleh kosong");
	}

	const result = await backendRequest<null>("/api/v1/me/reviews", {
		method: "POST",
		token: session.token,
		body: {
			rating,
			comment,
		},
	});

	if (result.error) {
		return redirect(`/reviews?error=${encodeURIComponent(result.error)}`);
	}

	return redirect("/orders?message=Review berhasil dikirim. Terima kasih!");
};
