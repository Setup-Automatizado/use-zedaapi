import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { auth } from "@/lib/auth";

export async function getAuthSession() {
	const session = await auth.api.getSession({
		headers: await headers(),
	});
	return session;
}

export async function requireAuth() {
	const session = await getAuthSession();
	if (!session) {
		redirect("/sign-in");
	}
	return session;
}

export async function requireAdmin() {
	const session = await requireAuth();
	if (session.user.role !== "admin") {
		redirect("/dashboard");
	}
	return session;
}
