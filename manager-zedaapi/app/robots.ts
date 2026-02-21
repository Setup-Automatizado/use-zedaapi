import type { MetadataRoute } from "next";

export default function robots(): MetadataRoute.Robots {
	return {
		rules: [
			{
				userAgent: "*",
				allow: "/",
				disallow: [
					"/admin",
					"/admin/",
					"/dashboard",
					"/dashboard/",
					"/api/",
					"/sign-in",
					"/sign-up",
					"/forgot-password",
					"/verify-email",
					"/two-factor",
				],
			},
		],
		sitemap: "https://zedaapi.com/sitemap.xml",
	};
}
