import type { NextConfig } from "next";

const nextConfig: NextConfig = {
	output: "standalone",

	serverExternalPackages: [
		"node-forge",
		"xml-crypto",
		"@xmldom/xmldom",
		"nodemailer",
		"bullmq",
		"ioredis",
	],

	images: {
		remotePatterns: [
			{
				protocol: "https",
				hostname: "**",
			},
		],
	},
};

export default nextConfig;
