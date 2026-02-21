import { getSessionCookie } from "better-auth/cookies";
import { type NextRequest, NextResponse } from "next/server";

const publicRoutes = [
	"/",
	"/blog",
	"/suporte",
	"/contato",
	"/sign-in",
	"/sign-up",
	"/forgot-password",
	"/reset-password",
	"/verify-email",
	"/waitlist",
	"/api/auth",
	"/api/webhooks",
	"/api/health",
];

const authRoutes = [
	"/sign-in",
	"/sign-up",
	"/forgot-password",
	"/reset-password",
	"/verify-email",
];

function isPublicRoute(pathname: string): boolean {
	return publicRoutes.some(
		(route) => pathname === route || pathname.startsWith(`${route}/`),
	);
}

function isAuthRoute(pathname: string): boolean {
	return authRoutes.some(
		(route) => pathname === route || pathname.startsWith(`${route}/`),
	);
}

export async function proxy(request: NextRequest) {
	const { pathname } = request.nextUrl;

	const sessionCookie = getSessionCookie(request);

	// Redirect authenticated users away from auth pages to dashboard
	if (isAuthRoute(pathname) && sessionCookie) {
		return NextResponse.redirect(new URL("/dashboard", request.url));
	}

	if (isPublicRoute(pathname)) {
		return NextResponse.next();
	}

	if (!sessionCookie) {
		const signInUrl = new URL("/sign-in", request.url);
		signInUrl.searchParams.set("callbackUrl", pathname);
		return NextResponse.redirect(signInUrl);
	}

	// Admin role check is enforced server-side in the admin layout.
	// Proxy only verifies the session cookie exists.
	return NextResponse.next();
}

export const config = {
	matcher: [
		"/((?!_next/static|_next/image|favicon.ico|public|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico)$).*)",
	],
};
