import { getSessionCookie } from "better-auth/cookies";
import { type NextRequest, NextResponse } from "next/server";

const publicRoutes = [
	"/",
	"/blog",
	"/suporte",
	"/glossario",
	"/contato",
	"/termos-de-uso",
	"/politica-de-privacidade",
	"/politica-de-cookies",
	"/lgpd",
	"/exclusao-de-dados",
	"/login",
	"/cadastro",
	"/esqueci-senha",
	"/redefinir-senha",
	"/verificar-email",
	"/lista-de-espera",
	"/api/auth",
	"/api/webhooks",
	"/api/health",
	"/api/brasil",
	"/api/admin/content",
	"/api/data-deletion",
];

const authRoutes = [
	"/login",
	"/cadastro",
	"/esqueci-senha",
	"/redefinir-senha",
	"/verificar-email",
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
		return NextResponse.redirect(new URL("/painel", request.url));
	}

	if (isPublicRoute(pathname)) {
		return NextResponse.next();
	}

	if (!sessionCookie) {
		const signInUrl = new URL("/login", request.url);
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
