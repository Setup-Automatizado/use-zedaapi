import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

/**
 * Rotas publicas que nao requerem autenticacao
 */
const publicRoutes = [
	"/login",
	"/register",
	"/forgot-password",
	"/reset-password",
	"/verify-2fa",
];

/**
 * Rotas de autenticacao - redireciona para dashboard se ja logado
 */
const authRoutes = ["/login"];

/**
 * Prefixos de rotas que devem ser ignorados pelo proxy
 */
const ignoredPrefixes = ["/api", "/_next", "/favicon.ico", "/images", "/fonts"];

/**
 * Proxy function for Next.js 16
 * Handles authentication and route protection
 */
export default async function proxy(request: NextRequest) {
	const { pathname } = request.nextUrl;

	// Ignorar rotas estaticas e API
	if (ignoredPrefixes.some((prefix) => pathname.startsWith(prefix))) {
		return NextResponse.next();
	}

	// Verificar se existe cookie de sessao do Better Auth
	// Nota: O prefixo do cookie e configurado em lib/auth.ts como "whatsapp-manager"
	// Quando SECURE_COOKIES=true, Better Auth adiciona prefixo "__Secure-"
	const sessionCookie =
		request.cookies.get("__Secure-whatsapp-manager.session_token") ||
		request.cookies.get("whatsapp-manager.session_token");
	const hasSession = !!sessionCookie?.value;

	// Verificar se e uma rota publica
	const isPublicRoute = publicRoutes.some((route) =>
		pathname.startsWith(route),
	);
	const isAuthRoute = authRoutes.some((route) => pathname.startsWith(route));

	// Se usuario logado tentando acessar rota de auth, redirecionar para dashboard
	if (hasSession && isAuthRoute) {
		return NextResponse.redirect(new URL("/dashboard", request.url));
	}

	// Se rota publica, permitir acesso
	if (isPublicRoute) {
		return NextResponse.next();
	}

	// Se nao tem sessao e nao e rota publica, redirecionar para login
	if (!hasSession) {
		const loginUrl = new URL("/login", request.url);
		// Preservar a URL original para redirect apos login
		if (pathname !== "/" && pathname !== "/dashboard") {
			loginUrl.searchParams.set("callbackUrl", pathname);
		}
		return NextResponse.redirect(loginUrl);
	}

	return NextResponse.next();
}

export const config = {
	matcher: [
		/*
		 * Match all request paths except:
		 * - api (API routes)
		 * - _next/static (static files)
		 * - _next/image (image optimization files)
		 * - favicon.ico (favicon file)
		 * - public folder
		 */
		"/((?!api|_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)",
	],
};
