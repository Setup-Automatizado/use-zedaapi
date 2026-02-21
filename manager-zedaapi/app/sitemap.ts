import type { MetadataRoute } from "next";
import { db } from "@/lib/db";

export const dynamic = "force-dynamic";

const BASE_URL = "https://zedaapi.com";

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
	const [blogPosts, blogCategories, supportArticles, glossaryTerms] =
		await Promise.all([
			db.blogPost.findMany({
				where: { status: "published" },
				select: { slug: true, publishedAt: true, updatedAt: true },
				orderBy: { publishedAt: "desc" },
			}),
			db.blogCategory.findMany({
				select: { slug: true, updatedAt: true },
			}),
			db.supportArticle.findMany({
				where: { status: "published" },
				select: { slug: true, updatedAt: true },
			}),
			db.glossaryTerm.findMany({
				where: { status: "published" },
				select: { slug: true, updatedAt: true },
			}),
		]);

	const staticPages: MetadataRoute.Sitemap = [
		{
			url: BASE_URL,
			lastModified: new Date(),
			changeFrequency: "weekly",
			priority: 1,
		},
		{
			url: `${BASE_URL}/contato`,
			lastModified: new Date(),
			changeFrequency: "monthly",
			priority: 0.7,
		},
		{
			url: `${BASE_URL}/blog`,
			lastModified: new Date(),
			changeFrequency: "daily",
			priority: 0.9,
		},
		{
			url: `${BASE_URL}/suporte`,
			lastModified: new Date(),
			changeFrequency: "weekly",
			priority: 0.8,
		},
		{
			url: `${BASE_URL}/glossario`,
			lastModified: new Date(),
			changeFrequency: "weekly",
			priority: 0.7,
		},
		{
			url: `${BASE_URL}/termos-de-uso`,
			lastModified: new Date(),
			changeFrequency: "yearly",
			priority: 0.3,
		},
		{
			url: `${BASE_URL}/politica-de-privacidade`,
			lastModified: new Date(),
			changeFrequency: "yearly",
			priority: 0.3,
		},
		{
			url: `${BASE_URL}/politica-de-cookies`,
			lastModified: new Date(),
			changeFrequency: "yearly",
			priority: 0.3,
		},
		{
			url: `${BASE_URL}/lgpd`,
			lastModified: new Date(),
			changeFrequency: "yearly",
			priority: 0.3,
		},
		{
			url: `${BASE_URL}/exclusao-de-dados`,
			lastModified: new Date(),
			changeFrequency: "yearly",
			priority: 0.3,
		},
	];

	const blogPostPages: MetadataRoute.Sitemap = blogPosts.map((post) => ({
		url: `${BASE_URL}/blog/${post.slug}`,
		lastModified: post.publishedAt ?? post.updatedAt,
		changeFrequency: "weekly" as const,
		priority: 0.8,
	}));

	const blogCategoryPages: MetadataRoute.Sitemap = blogCategories.map(
		(cat) => ({
			url: `${BASE_URL}/blog/categoria/${cat.slug}`,
			lastModified: cat.updatedAt,
			changeFrequency: "weekly" as const,
			priority: 0.6,
		}),
	);

	const supportPages: MetadataRoute.Sitemap = supportArticles.map(
		(article) => ({
			url: `${BASE_URL}/suporte/${article.slug}`,
			lastModified: article.updatedAt,
			changeFrequency: "monthly" as const,
			priority: 0.7,
		}),
	);

	const glossaryPages: MetadataRoute.Sitemap = glossaryTerms.map((term) => ({
		url: `${BASE_URL}/glossario/${term.slug}`,
		lastModified: term.updatedAt,
		changeFrequency: "monthly" as const,
		priority: 0.5,
	}));

	return [
		...staticPages,
		...blogPostPages,
		...blogCategoryPages,
		...supportPages,
		...glossaryPages,
	];
}
