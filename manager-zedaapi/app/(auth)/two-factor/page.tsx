import type { Metadata } from "next";
import { TwoFactorForm } from "@/components/auth/two-factor-form";

export const metadata: Metadata = {
	title: "Verificacao 2FA | Zé da API Manager",
	description: "Verificacao em duas etapas",
};

export default function TwoFactorPage() {
	return <TwoFactorForm />;
}
