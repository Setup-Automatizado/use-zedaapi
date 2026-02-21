import type { Metadata } from "next";
import { ForgotPasswordForm } from "@/components/auth/forgot-password-form";

export const metadata: Metadata = {
	title: "Esqueceu a senha | Zé da API Manager",
	description: "Recupere o acesso a sua conta",
};

export default function ForgotPasswordPage() {
	return <ForgotPasswordForm />;
}
