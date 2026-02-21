import type { Metadata } from "next";
import { Suspense } from "react";
import { VerifyEmailNotice } from "@/components/auth/verify-email-notice";

export const metadata: Metadata = {
	title: "Verificar e-mail | Zé da API Manager",
	description: "Verifique seu e-mail para ativar sua conta",
};

export default function VerifyEmailPage() {
	return (
		<Suspense>
			<VerifyEmailNotice />
		</Suspense>
	);
}
