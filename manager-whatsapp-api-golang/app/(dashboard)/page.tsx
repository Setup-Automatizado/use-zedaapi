import { redirect } from "next/navigation";

/**
 * Redirects dashboard group root to /dashboard
 */
export default function DashboardIndexPage() {
	redirect("/dashboard");
}
