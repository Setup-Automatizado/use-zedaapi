/**
 * Phone Number Formatting Utilities
 *
 * Centralized phone number formatting for international numbers.
 * Supports multiple country formats with proper spacing and grouping.
 *
 * @module lib/phone
 */

/**
 * Country-specific phone formatting rules
 */
interface CountryFormat {
	/** Country code (e.g., "55" for Brazil) */
	code: string;
	/** Expected length after country code */
	lengths: number[];
	/** Format function */
	format: (digits: string) => string;
}

/**
 * Country formatting configurations
 */
const COUNTRY_FORMATS: CountryFormat[] = [
	// Brazil: +55 DD NNNNN-NNNN or +55 DD NNNN-NNNN
	{
		code: "55",
		lengths: [10, 11], // 2 DDD + 8-9 local
		format: (digits: string) => {
			const areaCode = digits.substring(0, 2);
			const localNumber = digits.substring(2);
			if (localNumber.length === 9) {
				return `+55 ${areaCode} ${localNumber.substring(0, 5)}-${localNumber.substring(5)}`;
			}
			if (localNumber.length === 8) {
				return `+55 ${areaCode} ${localNumber.substring(0, 4)}-${localNumber.substring(4)}`;
			}
			return `+55 ${areaCode} ${localNumber}`;
		},
	},
	// Argentina mobile: +54 9 XXXX XX-XXXX
	{
		code: "549",
		lengths: [10, 11, 12], // Area + local varies
		format: (digits: string) => {
			if (digits.length >= 10) {
				const areaCode = digits.substring(0, 4);
				const localNumber = digits.substring(4);
				if (localNumber.length >= 6) {
					return `+54 9 ${areaCode} ${localNumber.substring(0, 2)}-${localNumber.substring(2)}`;
				}
			}
			return `+54 9 ${digits}`;
		},
	},
	// Argentina landline: +54 XX XXXX-XXXX
	{
		code: "54",
		lengths: [10, 11],
		format: (digits: string) => {
			if (digits.length >= 10) {
				const areaCode = digits.substring(0, 2);
				const localNumber = digits.substring(2);
				return `+54 ${areaCode} ${localNumber.substring(0, 4)}-${localNumber.substring(4)}`;
			}
			return `+54 ${digits}`;
		},
	},
	// USA/Canada: +1 XXX XXX-XXXX
	{
		code: "1",
		lengths: [10],
		format: (digits: string) => {
			if (digits.length === 10) {
				return `+1 ${digits.substring(0, 3)} ${digits.substring(3, 6)}-${digits.substring(6)}`;
			}
			return `+1 ${digits}`;
		},
	},
	// Mexico: +52 XXX XXX XXXX
	{
		code: "52",
		lengths: [10],
		format: (digits: string) => {
			if (digits.length === 10) {
				return `+52 ${digits.substring(0, 3)} ${digits.substring(3, 6)} ${digits.substring(6)}`;
			}
			return `+52 ${digits}`;
		},
	},
	// UK: +44 XXXX XXXXXX
	{
		code: "44",
		lengths: [10, 11],
		format: (digits: string) => {
			if (digits.length >= 10) {
				return `+44 ${digits.substring(0, 4)} ${digits.substring(4)}`;
			}
			return `+44 ${digits}`;
		},
	},
	// Germany: +49 XXX XXXXXXXX
	{
		code: "49",
		lengths: [10, 11, 12],
		format: (digits: string) => {
			if (digits.length >= 10) {
				return `+49 ${digits.substring(0, 3)} ${digits.substring(3)}`;
			}
			return `+49 ${digits}`;
		},
	},
	// Spain: +34 XXX XXX XXX
	{
		code: "34",
		lengths: [9],
		format: (digits: string) => {
			if (digits.length === 9) {
				return `+34 ${digits.substring(0, 3)} ${digits.substring(3, 6)} ${digits.substring(6)}`;
			}
			return `+34 ${digits}`;
		},
	},
	// Portugal: +351 XXX XXX XXX
	{
		code: "351",
		lengths: [9],
		format: (digits: string) => {
			if (digits.length === 9) {
				return `+351 ${digits.substring(0, 3)} ${digits.substring(3, 6)} ${digits.substring(6)}`;
			}
			return `+351 ${digits}`;
		},
	},
	// Italy: +39 XXX XXX XXXX
	{
		code: "39",
		lengths: [10, 11],
		format: (digits: string) => {
			if (digits.length >= 10) {
				return `+39 ${digits.substring(0, 3)} ${digits.substring(3, 6)} ${digits.substring(6)}`;
			}
			return `+39 ${digits}`;
		},
	},
	// France: +33 X XX XX XX XX
	{
		code: "33",
		lengths: [9],
		format: (digits: string) => {
			if (digits.length === 9) {
				return `+33 ${digits[0]} ${digits.substring(1, 3)} ${digits.substring(3, 5)} ${digits.substring(5, 7)} ${digits.substring(7)}`;
			}
			return `+33 ${digits}`;
		},
	},
	// Chile: +56 X XXXX XXXX
	{
		code: "56",
		lengths: [9],
		format: (digits: string) => {
			if (digits.length === 9) {
				return `+56 ${digits[0]} ${digits.substring(1, 5)} ${digits.substring(5)}`;
			}
			return `+56 ${digits}`;
		},
	},
	// Colombia: +57 XXX XXX XXXX
	{
		code: "57",
		lengths: [10],
		format: (digits: string) => {
			if (digits.length === 10) {
				return `+57 ${digits.substring(0, 3)} ${digits.substring(3, 6)} ${digits.substring(6)}`;
			}
			return `+57 ${digits}`;
		},
	},
	// Peru: +51 XXX XXX XXX
	{
		code: "51",
		lengths: [9],
		format: (digits: string) => {
			if (digits.length === 9) {
				return `+51 ${digits.substring(0, 3)} ${digits.substring(3, 6)} ${digits.substring(6)}`;
			}
			return `+51 ${digits}`;
		},
	},
	// India: +91 XXXXX XXXXX
	{
		code: "91",
		lengths: [10],
		format: (digits: string) => {
			if (digits.length === 10) {
				return `+91 ${digits.substring(0, 5)} ${digits.substring(5)}`;
			}
			return `+91 ${digits}`;
		},
	},
];

/**
 * Format a phone number for display
 *
 * Automatically detects country based on prefix and applies appropriate formatting.
 * Falls back to a generic international format if country is not recognized.
 *
 * @param phone - Raw phone number (can include + prefix or country code)
 * @param fallback - Value to return if phone is empty (default: empty string)
 * @returns Formatted phone number string
 *
 * @example
 * formatPhoneNumber("5511999887766") // "+55 11 99988-7766"
 * formatPhoneNumber("14155551234")   // "+1 415 555-1234"
 * formatPhoneNumber("")              // ""
 * formatPhoneNumber("", "-")         // "-"
 */
export function formatPhoneNumber(
	phone: string | undefined | null,
	fallback = "",
): string {
	if (!phone) return fallback;

	// Remove all non-digit characters
	const cleaned = phone.replace(/\D/g, "");

	if (cleaned.length < 6) return phone;

	// Sort by code length descending to match longer codes first (e.g., "549" before "54")
	const sortedFormats = [...COUNTRY_FORMATS].sort(
		(a, b) => b.code.length - a.code.length,
	);

	// Try to match a country format
	for (const country of sortedFormats) {
		if (cleaned.startsWith(country.code)) {
			const nationalNumber = cleaned.substring(country.code.length);
			// Check if length is valid for this country
			if (
				country.lengths.some(
					(len) =>
						nationalNumber.length === len ||
						nationalNumber.length === len + 1 ||
						nationalNumber.length === len - 1,
				)
			) {
				return country.format(nationalNumber);
			}
		}
	}

	// Default international format for unrecognized countries
	// Try to extract country code (1-3 digits) and format the rest
	if (cleaned.length > 10) {
		// Assume 2-digit country code
		const countryCode = cleaned.substring(0, 2);
		const rest = cleaned.substring(2);

		// Format as +XX XXXXX-XXXX or similar
		if (rest.length >= 8) {
			const split = Math.ceil(rest.length / 2);
			return `+${countryCode} ${rest.substring(0, split)}-${rest.substring(split)}`;
		}
		return `+${countryCode} ${rest}`;
	}

	// For short numbers, just return with + prefix
	return `+${cleaned}`;
}

/**
 * Format phone for input display (Brazilian format)
 *
 * Used in input fields where user is typing a phone number.
 * Formats as (XX) XXXXX-XXXX as user types.
 *
 * @param value - Current input value
 * @returns Formatted string for display in input
 *
 * @example
 * formatPhoneInput("11")       // "11"
 * formatPhoneInput("11999")    // "(11) 999"
 * formatPhoneInput("11999887766") // "(11) 99988-7766"
 */
export function formatPhoneInput(value: string): string {
	// Remove all non-digits
	const digits = value.replace(/\D/g, "");

	// Format as (XX) XXXXX-XXXX
	if (digits.length <= 2) return digits;
	if (digits.length <= 7) return `(${digits.slice(0, 2)}) ${digits.slice(2)}`;
	return `(${digits.slice(0, 2)}) ${digits.slice(2, 7)}-${digits.slice(7, 11)}`;
}

/**
 * Extract digits from a formatted phone number
 *
 * @param phone - Formatted phone number
 * @returns Only the digits
 *
 * @example
 * extractDigits("+55 11 99988-7766") // "5511999887766"
 * extractDigits("(11) 99988-7766")   // "11999887766"
 */
export function extractDigits(phone: string | undefined | null): string {
	if (!phone) return "";
	return phone.replace(/\D/g, "");
}

/**
 * Validate phone number length
 *
 * @param phone - Phone number (raw or formatted)
 * @param minDigits - Minimum number of digits (default: 10)
 * @param maxDigits - Maximum number of digits (default: 15)
 * @returns True if valid length
 */
export function isValidPhoneLength(
	phone: string | undefined | null,
	minDigits = 10,
	maxDigits = 15,
): boolean {
	const digits = extractDigits(phone);
	return digits.length >= minDigits && digits.length <= maxDigits;
}

/**
 * Get country code from phone number
 *
 * @param phone - Phone number with country code
 * @returns Country code or undefined if not recognized
 *
 * @example
 * getCountryCode("5511999887766") // "55"
 * getCountryCode("14155551234")   // "1"
 */
export function getCountryCode(phone: string | undefined | null): string | undefined {
	if (!phone) return undefined;

	const cleaned = phone.replace(/\D/g, "");

	// Sort by code length descending
	const sortedFormats = [...COUNTRY_FORMATS].sort(
		(a, b) => b.code.length - a.code.length,
	);

	for (const country of sortedFormats) {
		if (cleaned.startsWith(country.code)) {
			return country.code;
		}
	}

	return undefined;
}
