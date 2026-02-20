// =============================================================================
// NFS-e Nacional — Certificate A1 Management
// =============================================================================

import * as forge from "node-forge";
import type { NfseConfigData } from "./types";

interface CertificateCache {
	certPem: string;
	keyPem: string;
	expiresAt: number;
	loadedAt: number;
}

// Module-level singleton: in serverless environments (e.g. Vercel), this
// persists across warm invocations but resets on cold starts.
let certCache: CertificateCache | null = null;
const CACHE_TTL_MS = 30 * 60 * 1000;

/**
 * Load certificate PEM (cert + key) from PFX stored in S3.
 * Caches the result in memory to avoid re-downloading on every request.
 */
export async function loadCertificate(config: NfseConfigData): Promise<{
	certPem: string;
	keyPem: string;
}> {
	// Check cache (also invalidate if the certificate itself has expired)
	if (certCache && certCache.expiresAt && Date.now() >= certCache.expiresAt) {
		certCache = null;
	}
	if (certCache && Date.now() - certCache.loadedAt < CACHE_TTL_MS) {
		return { certPem: certCache.certPem, keyPem: certCache.keyPem };
	}

	// Validate certificate expiration
	if (new Date() >= config.certificateExpiresAt) {
		throw new Error(
			`Certificado A1 expirado em ${config.certificateExpiresAt.toISOString()}. Renove o certificado.`,
		);
	}

	// Decrypt the password
	const { decrypt } = await import("@/lib/crypto/encryption");
	const password = await decrypt(config.certificatePassword);

	// Read PFX directly from S3 via SDK (no public URL needed)
	const { getFile, extractKeyFromUrl } =
		await import("@/lib/services/storage/s3-client");
	const s3Key = extractKeyFromUrl(config.certificatePfxUrl);
	const pfxFileBuffer = await getFile(s3Key);
	if (!pfxFileBuffer) {
		throw new Error(`Certificado PFX nao encontrado no S3: ${s3Key}`);
	}

	const pfxDer = pfxFileBuffer.toString("binary");

	// Parse PKCS#12
	const p12Asn1 = forge.asn1.fromDer(pfxDer);
	const p12 = forge.pkcs12.pkcs12FromAsn1(p12Asn1, password);

	// Extract certificate
	const certBagType = forge.pki.oids.certBag;
	const certBags = p12.getBags({ bagType: certBagType });
	const certBag = certBagType ? certBags[certBagType] : undefined;
	if (!certBag || certBag.length === 0 || !certBag[0]?.cert) {
		throw new Error("Certificado nao encontrado no arquivo PFX");
	}
	const certPem = forge.pki.certificateToPem(certBag[0].cert);

	// Extract private key
	const keyBagType = forge.pki.oids.pkcs8ShroudedKeyBag;
	const keyBags = p12.getBags({ bagType: keyBagType });
	const keyBag = keyBagType ? keyBags[keyBagType] : undefined;
	if (!keyBag || keyBag.length === 0 || !keyBag[0]?.key) {
		throw new Error("Chave privada nao encontrada no arquivo PFX");
	}
	const keyPem = forge.pki.privateKeyToPem(keyBag[0].key);

	// Update cache
	certCache = {
		certPem,
		keyPem,
		expiresAt: config.certificateExpiresAt.getTime(),
		loadedAt: Date.now(),
	};

	return { certPem, keyPem };
}

/**
 * Invalidate the certificate cache (e.g., after config change).
 */
export function invalidateCertificateCache(): void {
	certCache = null;
}
