export const COOKIE_CONSENT_KEY = "ze_cookie_consent";
export const COOKIE_CONSENT_VERSION = "1.0";

export interface CookiePreferences {
	essential: true;
	analytics: boolean;
	marketing: boolean;
	functionality: boolean;
	timestamp: string;
	version: string;
}

export const DEFAULT_PREFERENCES: CookiePreferences = {
	essential: true,
	analytics: false,
	marketing: false,
	functionality: false,
	timestamp: "",
	version: COOKIE_CONSENT_VERSION,
};

export function getCookieConsent(): CookiePreferences | null {
	if (typeof document === "undefined") return null;

	const raw = document.cookie
		.split("; ")
		.find((row) => row.startsWith(`${COOKIE_CONSENT_KEY}=`))
		?.split("=")
		.slice(1)
		.join("=");

	if (!raw) return null;

	try {
		const parsed = JSON.parse(decodeURIComponent(raw)) as CookiePreferences;

		if (parsed.version !== COOKIE_CONSENT_VERSION) return null;

		return parsed;
	} catch {
		return null;
	}
}

export function setCookieConsent(prefs: CookiePreferences): void {
	if (typeof document === "undefined") return;

	const value = encodeURIComponent(JSON.stringify(prefs));
	const maxAge = 365 * 24 * 60 * 60;
	const secure =
		typeof window !== "undefined" && window.location.protocol === "https:"
			? "; Secure"
			: "";

	document.cookie = `${COOKIE_CONSENT_KEY}=${value}; Path=/; Max-Age=${maxAge}; SameSite=Lax${secure}`;
}

export function hasConsented(): boolean {
	const consent = getCookieConsent();
	return consent !== null;
}

export function acceptAllCookies(): CookiePreferences {
	const prefs: CookiePreferences = {
		essential: true,
		analytics: true,
		marketing: true,
		functionality: true,
		timestamp: new Date().toISOString(),
		version: COOKIE_CONSENT_VERSION,
	};
	setCookieConsent(prefs);
	return prefs;
}

export function rejectOptionalCookies(): CookiePreferences {
	const prefs: CookiePreferences = {
		...DEFAULT_PREFERENCES,
		timestamp: new Date().toISOString(),
	};
	setCookieConsent(prefs);
	return prefs;
}
