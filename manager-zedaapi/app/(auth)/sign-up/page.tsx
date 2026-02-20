import type { Metadata } from "next";
import { SignUpForm } from "@/components/auth/sign-up-form";

export const metadata: Metadata = {
	title: "Criar conta | Zé da API Manager",
	description: "Crie sua conta no Zé da API Manager",
};

export default function SignUpPage() {
	return <SignUpForm />;
}
