import type { Metadata } from "next";
import { LegalLayout } from "@/components/landing/legal-layout";
import { DataDeletionForm } from "./data-deletion-form";

export const metadata: Metadata = {
	title: "Exclusão de Dados - Zé da API",
	description:
		"Solicite a exclusão de seus dados pessoais da plataforma Zé da API Manager, conforme previsto na LGPD (Art. 18, VI).",
};

export default function ExclusaoDeDadosPage() {
	return (
		<LegalLayout title="Exclusão de Dados" lastUpdated="Fevereiro de 2026">
			<p>
				Em conformidade com o <strong>artigo 18, inciso VI</strong> da
				Lei Geral de Proteção de Dados Pessoais (Lei 13.709/2018 -
				LGPD), você tem o direito de solicitar a eliminação dos dados
				pessoais tratados com base no consentimento. A{" "}
				<strong>Zé da API</strong> respeita este direito e disponibiliza
				este formulário para facilitar sua solicitação.
			</p>

			<hr />

			{/* 1 */}
			<h2 id="o-que-acontece">
				1. O que Acontece quando Você Solicita a Exclusão
			</h2>
			<p>
				Ao solicitar a exclusão de seus dados, as seguintes ações serão
				realizadas:
			</p>

			<h3>1.1. Dados Removidos (em até 30 dias)</h3>
			<ul>
				<li>
					<strong>Dados da conta</strong>: nome, e-mail, senha (hash),
					configurações de perfil, preferências e tokens de
					autenticação serão permanentemente excluídos;
				</li>
				<li>
					<strong>Instâncias WhatsApp</strong>: todas as instâncias
					conectadas serão desconectadas e desprovisionadas. Sessões
					ativas serão encerradas;
				</li>
				<li>
					<strong>Configurações de webhooks</strong>: URLs, segredos e
					configurações de notificação serão removidos;
				</li>
				<li>
					<strong>Chaves de API</strong>: todos os tokens de acesso
					serão revogados e excluídos;
				</li>
				<li>
					<strong>Assinatura</strong>: o plano ativo será cancelado
					imediatamente, sem cobranças futuras;
				</li>
				<li>
					<strong>Logs de uso</strong>: registros de atividade na
					Plataforma serão anonimizados ou excluídos;
				</li>
				<li>
					<strong>Mídias</strong>: arquivos de mídia armazenados
					(imagens, vídeos, documentos) serão permanentemente
					excluídos do armazenamento.
				</li>
			</ul>

			<h3>1.2. Dados Retidos por Obrigação Legal</h3>
			<p>
				Conforme previsto no artigo 16 da LGPD, alguns dados serão
				retidos mesmo após a solicitação de exclusão, por exigência
				legal:
			</p>
			<ul>
				<li>
					<strong>Registros financeiros e NFS-e</strong>: retidos por{" "}
					<strong>5 (cinco) anos</strong> após a transação, conforme o
					Código Tributário Nacional (Art. 173) e legislação fiscal
					aplicável. Inclui histórico de pagamentos, faturas emitidas
					e dados fiscais (CPF/CNPJ para fins tributários);
				</li>
				<li>
					<strong>Logs de acesso (IP e data/hora)</strong>: retidos
					por <strong>6 (seis) meses</strong>, conforme o Marco Civil
					da Internet (Lei 12.965/2014, Art. 15);
				</li>
				<li>
					<strong>Dados necessários para processos judiciais</strong>:
					caso haja litígio em curso ou iminente, dados relevantes
					podem ser retidos até a conclusão definitiva do processo.
				</li>
			</ul>
			<p>
				Após o término do prazo de retenção legal, esses dados serão
				automaticamente anonimizados ou excluídos.
			</p>

			<h3>1.3. Consequências Irreversíveis</h3>
			<p>
				<strong>Atenção:</strong> a exclusão de dados é uma ação{" "}
				<strong>permanente e irreversível</strong>. Após a conclusão do
				processo:
			</p>
			<ul>
				<li>
					Não será possível recuperar a conta ou os dados excluídos;
				</li>
				<li>
					Todas as instâncias WhatsApp serão definitivamente
					desconectadas;
				</li>
				<li>
					Integrações e webhooks deixarão de funcionar imediatamente;
				</li>
				<li>
					Para utilizar a Plataforma novamente, será necessário criar
					uma nova conta do zero.
				</li>
			</ul>

			<hr />

			{/* 2 */}
			<h2 id="formulario">2. Formulário de Solicitação</h2>
			<p>
				Preencha o formulário abaixo para solicitar a exclusão de seus
				dados. Processaremos sua solicitação em até{" "}
				<strong>30 (trinta) dias corridos</strong>. Você receberá uma
				confirmação por e-mail quando o processo for concluído.
			</p>

			<DataDeletionForm />

			<hr />

			<h2 id="alternativas">3. Alternativas à Exclusão</h2>
			<p>
				Antes de solicitar a exclusão completa, considere as seguintes
				alternativas:
			</p>
			<ul>
				<li>
					<strong>Cancelar a assinatura</strong>: você pode cancelar
					seu plano sem excluir seus dados. A conta permanecerá ativa
					com acesso limitado;
				</li>
				<li>
					<strong>Desativar instâncias</strong>: desconecte instâncias
					WhatsApp individualmente sem afetar o restante da conta;
				</li>
				<li>
					<strong>Revogar consentimento de marketing</strong>: pare de
					receber comunicações promocionais sem excluir sua conta;
				</li>
				<li>
					<strong>Exportar dados (portabilidade)</strong>: solicite
					uma cópia de seus dados antes da exclusão, conforme seu
					direito previsto no Art. 18, V da LGPD.
				</li>
			</ul>
			<p>
				Para exercer qualquer dessas alternativas ou outros direitos
				previstos na LGPD, acesse a página{" "}
				<a href="/lgpd">LGPD - Direitos do Titular</a> ou entre em
				contato com nosso DPO em{" "}
				<a href="mailto:privacidade@zedaapi.com">
					privacidade@zedaapi.com
				</a>
				.
			</p>

			<hr />

			<p>
				<strong>Setup Automatizado Ltda</strong>
				<br />
				CNPJ: 54.246.473/0001-00
				<br />
				Rio de Janeiro, RJ, Brasil
			</p>
		</LegalLayout>
	);
}
