"use client";

import * as React from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import {
	Breadcrumb,
	BreadcrumbItem,
	BreadcrumbLink,
	BreadcrumbList,
	BreadcrumbPage,
	BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

const routeNameMap: Record<string, string> = {
	"": "Dashboard",
	instances: "Instances",
	health: "Health",
	settings: "Settings",
	profile: "Profile",
	new: "New",
};

export function DynamicBreadcrumb() {
	const pathname = usePathname();

	const generateBreadcrumbs = () => {
		const paths = pathname.split("/").filter(Boolean);

		const breadcrumbs = [
			{
				label: "Dashboard",
				href: "/",
				isLast: paths.length === 0,
			},
		];

		let currentPath = "";
		paths.forEach((path, index) => {
			currentPath += `/${path}`;
			const isLast = index === paths.length - 1;
			const label =
				routeNameMap[path] ||
				path.charAt(0).toUpperCase() + path.slice(1);

			breadcrumbs.push({
				label,
				href: currentPath,
				isLast,
			});
		});

		return breadcrumbs;
	};

	const breadcrumbs = generateBreadcrumbs();

	if (breadcrumbs.length <= 1 && breadcrumbs[0].isLast) {
		return null;
	}

	return (
		<Breadcrumb>
			<BreadcrumbList>
				{breadcrumbs.map((crumb) => (
					<React.Fragment key={crumb.href}>
						<BreadcrumbItem>
							{crumb.isLast ? (
								<BreadcrumbPage>{crumb.label}</BreadcrumbPage>
							) : (
								<BreadcrumbLink asChild>
									<Link href={crumb.href}>{crumb.label}</Link>
								</BreadcrumbLink>
							)}
						</BreadcrumbItem>
						{!crumb.isLast && <BreadcrumbSeparator />}
					</React.Fragment>
				))}
			</BreadcrumbList>
		</Breadcrumb>
	);
}
