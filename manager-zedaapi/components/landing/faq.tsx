"use client";

import { motion } from "framer-motion";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "@/components/ui/accordion";

const faqItems = [
	{
		question: "O que é o Zé da API e para quem é indicado?",
		answer: "O Zé da API é a plataforma brasileira de API profissional para WhatsApp. Você cria instâncias, conecta números via QR Code e usa nossa API REST para enviar e receber mensagens em escala. Atendemos desde startups que enviam centenas de mensagens por dia até operações enterprise com milhões de disparos mensais. Casos de uso mais comuns: automação de atendimento, notificações transacionais (pedidos, pagamentos, entregas), chatbots com IA, campanhas de marketing, integração com CRMs, ERPs, e-commerces e sistemas legados.",
	},
	{
		question: "Preciso de um número WhatsApp dedicado?",
		answer: "Sim, cada instância usa um número WhatsApp ativo. Pode ser qualquer número — mas recomendamos fortemente um chip dedicado à operação. Por quê? Números dedicados evitam conflitos com uso pessoal, permitem que a equipe gerencie o canal sem depender de um celular específico e reduzem o risco de bloqueios. Dica: um chip pré-pago de R$ 10 resolve. Muitos clientes usam números fixos com WhatsApp Business.",
	},
	{
		question: "Quanto tempo leva para começar a enviar mensagens?",
		answer: "Cronômetro na mão: 30 segundos para criar a conta, 1 minuto para criar a instância e escanear o QR Code, 2 minutos para enviar a primeira mensagem via API. Total real: menos de 5 minutos do zero à primeira mensagem entregue no WhatsApp. Sem aprovação manual, sem espera por suporte, sem configuração de servidor.",
	},
	{
		question: "Preciso de cartão de crédito para testar?",
		answer: "Não. Zero compromisso financeiro para começar. Todos os planos oferecem 7 dias de teste grátis sem exigir cartão de crédito, dados bancários ou qualquer forma de pagamento. Teste com calma, valide sua integração e só depois decida. Quando quiser assinar, aceitamos cartão de crédito internacional (Stripe) e PIX ou Boleto Híbrido (Sicredi) para clientes brasileiros.",
	},
	{
		question: "Qual o SLA de disponibilidade?",
		answer: "Definimos SLAs sérios por plano: Starter e Pro garantem 99,5% (até 3h36min de downtime/mês). Business e Scale sobem para 99,9% (até 43min/mês). Enterprise oferece 99,95% e Ultimate garante 99,99% (menos de 4,5 minutos/mês). Na prática, mantemos uptime acima do SLA contratado. Toda a infraestrutura opera com redundância geográfica, monitoramento 24/7, failover automático e alertas proativos antes que qualquer impacto chegue ao seu sistema.",
	},
	{
		question: "Como funciona a segurança dos dados?",
		answer: "Segurança não é feature, é fundação. Criptografia TLS 1.3 em trânsito e AES-256-GCM em repouso. Tokens de API gerados com entropia criptográfica e armazenados com hash — nunca em texto plano. Aderência total à LGPD (Lei 13.709/2018) com DPO dedicado (contato: privacidade@zedaapi.com). Seguimos o OWASP Top 10 como baseline de segurança. Seus dados nunca são compartilhados com terceiros, e você pode solicitar exclusão completa a qualquer momento conforme a LGPD.",
	},
	{
		question: "Posso cancelar a qualquer momento?",
		answer: 'Sim, sem fidelidade, sem multa, sem pegadinhas. Cancele diretamente pelo painel com um clique — nenhuma ligação para "retenção", nenhum formulário escondido. Seu acesso permanece ativo até o final do período já pago, e não cobramos taxa de cancelamento. Acreditamos que você fica porque quer, não porque está preso a um contrato.',
	},
	{
		question: "Vocês são parceiros oficiais da Meta/WhatsApp?",
		answer: "Transparência total: hoje o Zé da API opera via whatsmeow, uma biblioteca de código aberto amplamente utilizada no mercado. Paralelamente, estamos em fase avançada de integração com a Cloud API oficial da Meta (v24.0), o que trará suporte a funcionalidades exclusivas como templates oficiais, WhatsApp Flows e Business-Scoped User IDs. O usuário é responsável por cumprir as políticas de uso do WhatsApp/Meta, e oferecemos orientações claras sobre boas práticas na nossa documentação.",
	},
	{
		question: "Como funciona o suporte técnico?",
		answer: "Equipe 100% brasileira, 100% em português. Todos os planos incluem suporte por e-mail (suporte@zedaapi.com) com resposta em até 24h úteis. A partir do plano Business, o suporte é prioritário via WhatsApp com resposta em até 4h. Planos Scale e superiores contam com suporte 24/7, gerente de conta dedicado e onboarding assistido. Além do suporte humano, mantemos documentação completa, exemplos de código em 4 linguagens, coleção Postman e base de conhecimento atualizada semanalmente.",
	},
	{
		question: "Posso integrar com n8n, Make ou Zapier?",
		answer: "Sim — e com qualquer outra ferramenta que fale HTTP. Oferecemos nodes nativos para n8n, integração com Make (Integromat) e Zapier para automações no-code. Nossa API REST é compatível com toda linguagem de programação: Node.js, Python, PHP, Go, Java, Ruby, C# e mais. Disponibilizamos SDKs oficiais para Node.js e Python, coleção Postman importável e webhooks configuráveis por evento. Se sua ferramenta faz requisições HTTP, ela funciona com o Zé da API.",
	},
] as const;

export function FAQ() {
	return (
		<section id="faq" className="relative py-20 sm:py-28">
			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				{/* Header */}
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					whileInView={{ opacity: 1, y: 0 }}
					viewport={{ once: true, margin: "-80px" }}
					transition={{ duration: 0.5 }}
					className="mx-auto max-w-2xl text-center"
				>
					<p className="text-sm font-semibold uppercase tracking-widest text-primary">
						FAQ
					</p>
					<h2 className="mt-3 text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
						Perguntas Frequentes
					</h2>
					<p className="mt-4 text-base text-muted-foreground sm:text-lg leading-relaxed">
						Respostas objetivas para as dúvidas mais comuns. Não
						encontrou a sua?{" "}
						<a
							href="#contato"
							className="font-medium text-primary hover:underline underline-offset-4"
						>
							Fale conosco
						</a>
						.
					</p>
				</motion.div>

				{/* Accordion */}
				<motion.div
					initial={{ opacity: 0, y: 24 }}
					whileInView={{ opacity: 1, y: 0 }}
					viewport={{ once: true, margin: "-60px" }}
					transition={{ duration: 0.5, delay: 0.15 }}
					className="mx-auto mt-14 max-w-3xl"
				>
					<Accordion type="single" collapsible>
						{faqItems.map((item, index) => (
							<AccordionItem key={index} value={`item-${index}`}>
								<AccordionTrigger>
									{item.question}
								</AccordionTrigger>
								<AccordionContent>
									<p className="text-sm text-muted-foreground leading-relaxed">
										{item.answer}
									</p>
								</AccordionContent>
							</AccordionItem>
						))}
					</Accordion>
				</motion.div>
			</div>
		</section>
	);
}
