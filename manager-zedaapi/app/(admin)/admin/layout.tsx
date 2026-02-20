import { requireAdmin } from "@/lib/auth-server";
import { AdminShell } from "@/components/layout/admin-shell";

export default async function AdminLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	await requireAdmin();

	return <AdminShell>{children}</AdminShell>;
}
