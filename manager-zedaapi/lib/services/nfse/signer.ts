// =============================================================================
// NFS-e Nacional — XMLDSIG Signer (RSA-SHA256)
// =============================================================================

import { DOMParser } from "@xmldom/xmldom";
import { SignedXml } from "xml-crypto";

const ID_PATTERN = /^[A-Za-z0-9_-]+$/;

/**
 * Sign a DPS XML with XMLDSIG enveloped signature using RSA-SHA256.
 * Required by NFS-e Nacional API.
 */
export function signDpsXml(
	xml: string,
	certPem: string,
	keyPem: string,
): string {
	const doc = new DOMParser().parseFromString(xml, "text/xml");

	// Find the infDPS element to get its Id attribute
	const infDPS = doc.getElementsByTagName("infDPS")[0];
	if (!infDPS) {
		throw new Error("Elemento infDPS nao encontrado no XML");
	}

	const infDPSId = infDPS.getAttribute("Id");
	if (!infDPSId) {
		throw new Error("Atributo Id nao encontrado em infDPS");
	}

	if (!ID_PATTERN.test(infDPSId)) {
		throw new Error(`Invalid infDPS Id attribute: ${infDPSId}`);
	}

	const sig = new SignedXml({
		privateKey: keyPem,
		publicCert: certPem,
		canonicalizationAlgorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
		signatureAlgorithm: "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
	});

	sig.addReference({
		xpath: `//*[@Id='${infDPSId}']`,
		digestAlgorithm: "http://www.w3.org/2001/04/xmlenc#sha256",
		transforms: [
			"http://www.w3.org/2000/09/xmldsig#enveloped-signature",
			"http://www.w3.org/2001/10/xml-exc-c14n#",
		],
	});

	sig.computeSignature(xml, {
		location: { reference: `//*[local-name(.)='infDPS']`, action: "after" },
	});

	return sig.getSignedXml();
}

/**
 * Sign a pedRegEvento XML with XMLDSIG enveloped signature using RSA-SHA256.
 * Same algorithm as DPS, but targets the infPedReg element instead of infDPS.
 */
export function signEventXml(
	xml: string,
	certPem: string,
	keyPem: string,
): string {
	const doc = new DOMParser().parseFromString(xml, "text/xml");

	const infPedReg = doc.getElementsByTagName("infPedReg")[0];
	if (!infPedReg) {
		throw new Error("Elemento infPedReg nao encontrado no XML");
	}

	const infPedRegId = infPedReg.getAttribute("Id");
	if (!infPedRegId) {
		throw new Error("Atributo Id nao encontrado em infPedReg");
	}

	if (!ID_PATTERN.test(infPedRegId)) {
		throw new Error(`Invalid infPedReg Id attribute: ${infPedRegId}`);
	}

	const sig = new SignedXml({
		privateKey: keyPem,
		publicCert: certPem,
		canonicalizationAlgorithm: "http://www.w3.org/2001/10/xml-exc-c14n#",
		signatureAlgorithm: "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
	});

	sig.addReference({
		xpath: `//*[@Id='${infPedRegId}']`,
		digestAlgorithm: "http://www.w3.org/2001/04/xmlenc#sha256",
		transforms: [
			"http://www.w3.org/2000/09/xmldsig#enveloped-signature",
			"http://www.w3.org/2001/10/xml-exc-c14n#",
		],
	});

	sig.computeSignature(xml, {
		location: {
			reference: `//*[local-name(.)='infPedReg']`,
			action: "after",
		},
	});

	return sig.getSignedXml();
}
