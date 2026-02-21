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
					"/painel",
					"/painel/",
					"/api/",
					"/login",
					"/cadastro",
					"/esqueci-senha",
					"/verificar-email",
					"/dois-fatores",
				],
			},
		],
		sitemap: "https://zedaapi.com/sitemap.xml",
	};
}
