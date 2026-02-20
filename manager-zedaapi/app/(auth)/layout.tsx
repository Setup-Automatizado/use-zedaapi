import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { getAuthSession } from "@/lib/auth-server";

export const metadata: Metadata = {
	title: "Zé da API Manager",
	description: "Gerencie suas instancias WhatsApp API",
};

export default async function AuthLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	const session = await getAuthSession();

	if (session) {
		redirect("/dashboard");
	}

	return (
		<div className="flex min-h-svh items-center justify-center bg-background p-4">
			<div className="w-full max-w-[400px]">
				<div className="mb-8 flex flex-col items-center gap-2">
					<div className="flex size-10 items-center justify-center rounded-xl bg-primary text-primary-foreground font-bold text-lg">
						Z
					</div>
					<span className="text-lg font-semibold tracking-tight">
						Zé da API Manager
					</span>
				</div>
				{children}
			</div>
		</div>
	);
}
