import type { Metadata } from "next";
import { WaitlistForm } from "@/components/auth/waitlist-form";

export const metadata: Metadata = {
	title: "Lista de espera | Zé da API Manager",
	description: "Entre na lista de espera do Zé da API Manager",
};

export default function WaitlistPage() {
	return <WaitlistForm />;
}
