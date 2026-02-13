import { redirect } from "next/navigation";

/**
 * Pagina raiz redireciona para o dashboard
 * O middleware cuida da autenticacao
 */
export default function RootPage() {
	redirect("/dashboard");
}
