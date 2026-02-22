"use client";

import { useState, useMemo, useRef, useCallback } from "react";
import Link from "next/link";
import { motion, useInView } from "framer-motion";
import {
	SearchIcon,
	RocketIcon,
	CreditCardIcon,
	ShieldCheckIcon,
	CodeIcon,
	HeadphonesIcon,
	BuildingIcon,
	ScaleIcon,
	ArrowRightIcon,
	XIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "@/components/ui/accordion";

// ── Types ──────────────────────────

interface FaqItem {
	question: string;
	answer: string;
}

interface FaqCategory {
	id: string;
	label: string;
	icon: typeof RocketIcon;
	items: FaqItem[];
}

// ── FAQ Data ──────────────────────────

const faqCategories: FaqCategory[] = [
	{
		id: "getting-started",
		label: "Primeiros Passos",
		icon: RocketIcon,
		items: [
			{
				question: "O que é o Zé da API e para quem é indicado?",
				answer: "O Zé da API é uma plataforma brasileira de API REST para WhatsApp. Conecte números via QR Code e envie/receba mensagens em escala. Usado para automação de atendimento, notificações transacionais, chatbots, confirmação de pedidos, alertas de sistema e integrações com CRMs, ERPs e ferramentas no-code como n8n, Make e Zapier. Ideal para startups, e-commerces, SaaS, clínicas, escolas e qualquer empresa que precisa de WhatsApp automatizado.",
			},
			{
				question: "Preciso de um número WhatsApp dedicado?",
				answer: "Sim, cada instância usa um número WhatsApp. Recomendamos um chip dedicado à operação para evitar conflitos com uso pessoal. Um chip pré-pago de R$ 10 resolve. Você pode usar múltiplos números simultaneamente no mesmo painel — cada um opera de forma independente com QR Code e token próprios.",
			},
			{
				question: "Quanto tempo leva para começar a enviar mensagens?",
				answer: "Menos de 5 minutos. Crie a conta, ative uma instância, escaneie o QR Code com o WhatsApp do celular e envie a primeira mensagem via API. Sem aprovação manual, sem burocracia e sem espera. A documentação interativa inclui exemplos prontos para copiar e colar em Node.js, Python, PHP e Go.",
			},
			{
				question: "Preciso de cartão de crédito para testar?",
				answer: "Não. Todos os planos incluem 7 dias grátis sem cartão de crédito. Quando decidir assinar, aceitamos cartão de crédito internacional (via Stripe) e PIX/Boleto para clientes brasileiros (via Sicredi). Você pode cancelar a qualquer momento durante o trial sem ser cobrado.",
			},
			{
				question: "Preciso de conhecimento técnico para usar?",
				answer: "Depende do uso. Para integrações no-code (n8n, Make, Zapier), não é necessário programar — basta arrastar e conectar os nodes. Para integração via API, um conhecimento básico de HTTP/REST é suficiente. A documentação inclui exemplos passo a passo com curl, e temos Postman collection pronta. Nosso suporte em português também ajuda na configuração inicial.",
			},
			{
				question: "Posso enviar mensagens em massa?",
				answer: "Sim, com responsabilidade. A API suporta envio em escala para grandes volumes, mas você deve seguir as políticas do WhatsApp para evitar banimentos. Recomendamos: respeitar horários comerciais, não enviar spam, manter uma lista de opt-in e tratar opt-outs imediatamente. A plataforma inclui rate limiting configurável por instância para proteger seu número.",
			},
		],
	},
	{
		id: "pricing",
		label: "Preços e Pagamento",
		icon: CreditCardIcon,
		items: [
			{
				question: "Como funciona a cobrança? Existe taxa por mensagem?",
				answer: "Zero cobrança por mensagem enviada ou recebida. O modelo é de preço fixo mensal por faixa de instâncias (números WhatsApp). Escolha a faixa que atende seu volume e envie sem limite — até 10 milhões de mensagens por mês comprovados. Sem surpresas na fatura, sem excedentes, sem taxas ocultas.",
			},
			{
				question: "Posso cancelar a qualquer momento?",
				answer: "Sim, sem fidelidade e sem multa. Cancele pelo painel com um clique — seu acesso continua ativo até o fim do período já pago. Não cobramos taxa de cancelamento e não fazemos retenção forçada. Se precisar voltar depois, é só reativar.",
			},
			{
				question: "Quais formas de pagamento são aceitas?",
				answer: "Cartão de crédito internacional (Visa, Mastercard, Amex) processado via Stripe com segurança PCI DSS Level 1. Para clientes brasileiros, também aceitamos PIX (aprovação instantânea) e Boleto Bancário (compensação em até 2 dias úteis), ambos processados via parceria com o Sicredi.",
			},
			{
				question: "Como funciona o plano anual?",
				answer: "No plano anual você economiza 20% em relação ao mensal. O valor é cobrado em parcela única no início do ciclo. Você mantém todas as funcionalidades do plano mensal, com a vantagem do desconto. Caso cancele antes do fim do ano, o acesso continua até o período pago expirar.",
			},
			{
				question: "E se eu precisar de mais instâncias durante o mês?",
				answer: "Basta ajustar no painel a qualquer momento. O upgrade é imediato e o valor é calculado pro-rata para o período restante. Se fizer downgrade, a diferença é creditada para o próximo ciclo. Sem burocracia, sem espera — tudo automatizado.",
			},
			{
				question: "Emitem Nota Fiscal?",
				answer: "Sim. Emitimos NFS-e (Nota Fiscal de Serviço Eletrônica) automaticamente para todos os pagamentos. A nota é enviada por e-mail após a confirmação do pagamento. Para clientes PJ, basta informar o CNPJ e razão social no painel que a emissão é feita com os dados fiscais corretos.",
			},
		],
	},
	{
		id: "technical",
		label: "Técnico e API",
		icon: CodeIcon,
		items: [
			{
				question: "Quais tipos de mensagem a API suporta?",
				answer: "Texto, imagem, áudio (incluindo PTT — push-to-talk como áudio do WhatsApp), vídeo, documentos/PDF, localização, contatos (vCard), botões de resposta rápida, listas de seleção, stickers e mensagens com link preview. Todos os tipos são enviados via um único endpoint REST com JSON. A documentação completa está em api.zedaapi.com/docs.",
			},
			{
				question: "Como funcionam os webhooks?",
				answer: "Cada mensagem recebida, leitura, status de entrega e evento de grupo é enviado instantaneamente ao seu servidor via HTTP POST (webhook). Latência mediana de 47ms. Você configura a URL de destino por instância no painel. Inclui retry automático com backoff exponencial (3 tentativas) e Dead Letter Queue para mensagens que falharam.",
			},
			{
				question: "Posso integrar com n8n, Make ou Zapier?",
				answer: "Sim. Temos node nativo para n8n (@setup-automatizado/n8n-nodes-zedaapi) disponível no npm. Para Make e Zapier, a integração é feita via módulo HTTP genérico — a API REST padrão funciona com qualquer ferramenta que fale HTTP/JSON. Postman collection inclusa na documentação para facilitar os testes.",
			},
			{
				question: "Qual a latência média da API?",
				answer: "Latência mediana de 47ms para envio de mensagens (tempo entre a chamada API e a confirmação do servidor). O tempo total até a entrega no WhatsApp do destinatário depende da rede do usuário, mas tipicamente é menor que 1 segundo. Monitoramos latência em tempo real com Prometheus e alertas automáticos.",
			},
			{
				question: "Existe limite de mensagens por minuto?",
				answer: "Não há limite fixo imposto pela API. O rate limiting é configurável por instância no painel. Na prática, o WhatsApp tem seus próprios limites (que variam por conta e reputação do número). Recomendamos não ultrapassar 60 mensagens/minuto por número para manter a saúde da conta. Para volumes maiores, use múltiplas instâncias.",
			},
			{
				question: "A API funciona com WhatsApp Business?",
				answer: "Sim, funciona tanto com WhatsApp pessoal quanto com WhatsApp Business. Estamos em fase avançada de integração com a Cloud API oficial da Meta (v24.0), que permitirá usar ambos os provedores simultaneamente (coexistence mode). Atualmente operamos via whatsmeow, biblioteca open-source robusta e confiável.",
			},
			{
				question: "Quais linguagens de programação são suportadas?",
				answer: "Qualquer linguagem que faça requisições HTTP funciona. A documentação inclui exemplos completos em Node.js/TypeScript, Python, PHP e Go. A API é REST pura com JSON — não precisa de SDK proprietário. Para Node.js, um simples fetch() resolve. Para Python, requests. Para PHP, Guzzle ou Http facade do Laravel.",
			},
		],
	},
	{
		id: "security",
		label: "Segurança e LGPD",
		icon: ShieldCheckIcon,
		items: [
			{
				question: "Como funciona a segurança dos dados?",
				answer: "Criptografia TLS 1.3 em trânsito e AES-256-GCM em repouso para dados sensíveis. Tokens de autenticação armazenados com hash criptográfico. Rate limiting por instância para prevenir abuso. Infraestrutura com monitoramento 24/7, alertas automáticos e logs de auditoria. Nenhum dado de mensagem é armazenado permanentemente — processamos e entregamos.",
			},
			{
				question: "A plataforma é compatível com a LGPD?",
				answer: "Sim. Operamos em conformidade com a LGPD (Lei Geral de Proteção de Dados). Mantemos DPO (Data Protection Officer) dedicado, política de privacidade transparente, mecanismo de exclusão de dados via API e painel, e processamos dados pessoais apenas para a finalidade contratada. Nossos servidores são hospedados em infraestrutura brasileira.",
			},
			{
				question: "Vocês armazenam o conteúdo das mensagens?",
				answer: "Não armazenamos o conteúdo das mensagens permanentemente. As mensagens são processadas em memória, entregues via webhook ao seu servidor e descartadas. Metadados (timestamps, status de entrega) são mantidos temporariamente para fins operacionais e métricas, respeitando o período de retenção configurado. Mídias são armazenadas em S3 com URLs temporárias que expiram.",
			},
			{
				question: "Como funciona a autenticação da API?",
				answer: "Cada instância possui um Client-Token exclusivo de no mínimo 16 caracteres. O token é enviado no header da requisição (Client-Token). Tokens não são logados, não são expostos em erros e são armazenados com hash criptográfico. Você pode rotacionar tokens a qualquer momento pelo painel sem interrupção do serviço.",
			},
		],
	},
	{
		id: "reliability",
		label: "SLA e Infraestrutura",
		icon: BuildingIcon,
		items: [
			{
				question: "Qual o SLA de disponibilidade?",
				answer: "Starter e Pro: 99,5% de uptime (até 3,6h/mês de downtime permitido). Business e Scale: 99,9% (até 43min/mês). Enterprise: 99,95% (até 22min/mês). Infraestrutura redundante com failover automático, health checks a cada 30 segundos e monitoramento com Prometheus + Sentry. Status em tempo real disponível na página de status.",
			},
			{
				question: "O que acontece se o celular ficar offline?",
				answer: "As mensagens ficam em fila (FIFO) e são entregues assim que o celular reconecta. A fila mantém ordenação estrita por instância para evitar mensagens fora de ordem. Se o celular ficar offline por mais de 14 dias, o WhatsApp desconecta a sessão e um novo QR Code será necessário. O painel mostra o status de conexão em tempo real.",
			},
			{
				question: "Existe redundância na infraestrutura?",
				answer: "Sim. Banco de dados PostgreSQL com réplicas, Redis para cache e filas com persistência, NATS JetStream para mensageria distribuída com replay, S3 compatível para armazenamento de mídia e circuit breaker em todas as integrações externas. Failover automático sem intervenção manual. Monitoramento 24/7 com alertas em múltiplos canais.",
			},
		],
	},
	{
		id: "support",
		label: "Suporte",
		icon: HeadphonesIcon,
		items: [
			{
				question: "Como funciona o suporte técnico?",
				answer: "Equipe 100% brasileira, suporte em português. E-mail com resposta em até 24h para todos os planos. Planos Business+ incluem suporte prioritário via WhatsApp com SLA de resposta de 4 horas. Documentação completa com exemplos em 4 linguagens, Postman collection, guias de integração passo a passo e FAQ detalhado.",
			},
			{
				question: "Vocês ajudam na integração inicial?",
				answer: "Sim. Oferecemos onboarding assistido para todos os clientes. A documentação interativa em api.zedaapi.com/docs permite testar endpoints diretamente no navegador. Para integrações complexas ou de grande volume, agende uma call técnica gratuita com nosso time de engenharia. Planos Business+ incluem arquiteto de soluções dedicado.",
			},
			{
				question: "Onde encontro a documentação da API?",
				answer: "Documentação completa e interativa (Swagger/OpenAPI) em api.zedaapi.com/docs. Inclui todos os endpoints, parâmetros, exemplos de requisição/resposta, códigos de erro e exemplos em Node.js, Python, PHP e Go. Também disponibilizamos Postman collection para importação direta e node nativo para n8n.",
			},
		],
	},
	{
		id: "legal",
		label: "Legal e Compliance",
		icon: ScaleIcon,
		items: [
			{
				question: "Vocês são parceiros oficiais da Meta/WhatsApp?",
				answer: "Atualmente operamos via whatsmeow, uma biblioteca open-source robusta e amplamente utilizada. Estamos em fase avançada de integração com a Cloud API oficial da Meta (v24.0), o que permitirá usar ambos os provedores de forma coexistente. O usuário é responsável por cumprir as Políticas Comerciais e os Termos de Serviço do WhatsApp.",
			},
			{
				question: "Existe risco de banimento do número?",
				answer: "O risco de banimento pelo WhatsApp existe quando as políticas de uso são violadas (spam, mensagens não solicitadas, conteúdo proibido). Para minimizar o risco: use listas de opt-in, respeite opt-outs imediatamente, envie conteúdo relevante, respeite horários comerciais e não envie em volume excessivo para um número novo. A plataforma inclui rate limiting configurável e boas práticas na documentação.",
			},
			{
				question: "Posso usar para enviar mensagens de marketing?",
				answer: "Sim, desde que os destinatários tenham dado consentimento prévio (opt-in). As mensagens de marketing devem seguir as políticas do WhatsApp: conteúdo relevante, possibilidade de opt-out fácil e respeito aos limites de frequência. Recomendamos segmentar suas listas e personalizar as mensagens para melhores taxas de entrega e menor risco.",
			},
			{
				question:
					"Como funciona a exclusão de dados (direito ao esquecimento)?",
				answer: "Atendemos ao Art. 18 da LGPD. O titular pode solicitar exclusão via endpoint dedicado na API ou pelo painel do administrador. A exclusão é processada em até 48 horas úteis e abrange todos os dados pessoais: metadados de mensagens, logs de webhook e registros de instância. Emitimos confirmação por e-mail ao término do processo.",
			},
		],
	},
];

const totalQuestions = faqCategories.reduce((a, c) => a + c.items.length, 0);

// ── Component ──────────────────────────

export function FAQ() {
	const [searchQuery, setSearchQuery] = useState("");
	const [activeCategory, setActiveCategory] = useState<string | null>(null);
	const ref = useRef<HTMLDivElement>(null);
	const isInView = useInView(ref, { once: true, margin: "-80px" });

	const handleSearch = useCallback(
		(e: React.ChangeEvent<HTMLInputElement>) => {
			setSearchQuery(e.target.value);
		},
		[],
	);

	const handleClearSearch = useCallback(() => {
		setSearchQuery("");
	}, []);

	const handleCategoryClick = useCallback((catId: string | null) => {
		setActiveCategory((prev) => (prev === catId ? null : catId));
	}, []);

	const handleClearAll = useCallback(() => {
		setSearchQuery("");
		setActiveCategory(null);
	}, []);

	const filteredCategories = useMemo(() => {
		const query = searchQuery.toLowerCase().trim();

		let result = faqCategories;

		if (activeCategory) {
			result = result.filter((cat) => cat.id === activeCategory);
		}

		if (query) {
			result = result
				.map((cat) => ({
					...cat,
					items: cat.items.filter(
						(item) =>
							item.question.toLowerCase().includes(query) ||
							item.answer.toLowerCase().includes(query),
					),
				}))
				.filter((cat) => cat.items.length > 0);
		}

		return result;
	}, [searchQuery, activeCategory]);

	const totalResults = filteredCategories.reduce(
		(acc, cat) => acc + cat.items.length,
		0,
	);

	const hasActiveFilters = searchQuery !== "" || activeCategory !== null;

	return (
		<section id="faq" className="relative py-20 sm:py-28" ref={ref}>
			<div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				{/* Header */}
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					animate={isInView ? { opacity: 1, y: 0 } : {}}
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
						Respostas objetivas e completas para as dúvidas mais
						comuns. Não encontrou a sua?{" "}
						<Link
							href="/contato"
							className="font-medium text-primary hover:underline underline-offset-4"
						>
							Fale conosco
						</Link>
						.
					</p>
				</motion.div>

				{/* Search + Category Filters */}
				<motion.div
					initial={{ opacity: 0, y: 16 }}
					animate={isInView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.5, delay: 0.1 }}
					className="mx-auto mt-10 max-w-3xl space-y-5"
				>
					{/* Search Bar */}
					<div className="relative">
						<SearchIcon className="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
						<input
							type="search"
							value={searchQuery}
							onChange={handleSearch}
							placeholder="Buscar pergunta... ex: webhook, preço, n8n, LGPD"
							className="h-11 w-full rounded-xl border border-border bg-background pl-10 pr-20 text-sm text-foreground placeholder:text-muted-foreground/60 transition-all duration-200 focus:border-primary/50 focus:outline-none focus:ring-2 focus:ring-primary/20"
							aria-label="Buscar perguntas frequentes"
						/>
						{searchQuery && (
							<button
								type="button"
								onClick={handleClearSearch}
								className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-1 rounded-md bg-muted px-2 py-1 text-xs font-medium text-muted-foreground hover:bg-muted/80 hover:text-foreground transition-colors"
								aria-label="Limpar busca"
							>
								<XIcon className="size-3" />
								Limpar
							</button>
						)}
					</div>

					{/* Category Pills */}
					<div className="flex flex-wrap items-center justify-center gap-2">
						<button
							type="button"
							onClick={() => handleCategoryClick(null)}
							className={`rounded-full border px-3.5 py-1.5 text-xs font-medium transition-all duration-200 ${
								!activeCategory
									? "border-primary/40 bg-primary/10 text-primary shadow-sm shadow-primary/10"
									: "border-border bg-muted/50 text-muted-foreground hover:text-foreground hover:border-border hover:bg-muted"
							}`}
						>
							Todos ({totalQuestions})
						</button>
						{faqCategories.map((cat) => {
							const Icon = cat.icon;
							const isActive = activeCategory === cat.id;
							return (
								<button
									key={cat.id}
									type="button"
									onClick={() => handleCategoryClick(cat.id)}
									className={`flex items-center gap-1.5 rounded-full border px-3.5 py-1.5 text-xs font-medium transition-all duration-200 ${
										isActive
											? "border-primary/40 bg-primary/10 text-primary shadow-sm shadow-primary/10"
											: "border-border bg-muted/50 text-muted-foreground hover:text-foreground hover:border-border hover:bg-muted"
									}`}
								>
									<Icon className="size-3" />
									{cat.label}
								</button>
							);
						})}
					</div>

					{/* Results count when filtering */}
					{hasActiveFilters && (
						<p className="text-center text-xs text-muted-foreground animate-in fade-in duration-200">
							{totalResults === 0
								? "Nenhum resultado encontrado. Tente outros termos."
								: `${totalResults} resultado${totalResults !== 1 ? "s" : ""} encontrado${totalResults !== 1 ? "s" : ""}`}
						</p>
					)}
				</motion.div>

				{/* FAQ Categories + Accordions */}
				<div className="mx-auto mt-10 max-w-3xl space-y-8">
					{filteredCategories.map((category) => {
						const Icon = category.icon;
						return (
							<div
								key={category.id}
								className="space-y-3 animate-in fade-in slide-in-from-bottom-2 duration-300"
							>
								{/* Category Header */}
								<div className="flex items-center gap-2.5">
									<div className="flex size-7 items-center justify-center rounded-lg bg-primary/10">
										<Icon className="size-3.5 text-primary" />
									</div>
									<h3 className="text-sm font-semibold text-foreground">
										{category.label}
									</h3>
									<div className="h-px flex-1 bg-border/50" />
									<span className="text-[10px] font-medium text-muted-foreground/60">
										{category.items.length}{" "}
										{category.items.length === 1
											? "pergunta"
											: "perguntas"}
									</span>
								</div>

								{/* Accordion */}
								<Accordion
									type="single"
									collapsible
									key={`accordion-${category.id}-${searchQuery}-${activeCategory}`}
								>
									{category.items.map((item, index) => (
										<AccordionItem
											key={`${category.id}-${index}`}
											value={`${category.id}-${index}`}
										>
											<AccordionTrigger className="text-left">
												{item.question}
											</AccordionTrigger>
											<AccordionContent>
												<p className="text-sm text-muted-foreground leading-relaxed whitespace-pre-line">
													{item.answer}
												</p>
											</AccordionContent>
										</AccordionItem>
									))}
								</Accordion>
							</div>
						);
					})}

					{/* Empty state */}
					{filteredCategories.length === 0 && (
						<div className="flex flex-col items-center justify-center py-16 text-center animate-in fade-in duration-300">
							<SearchIcon className="size-10 text-muted-foreground/30" />
							<p className="mt-4 text-sm font-medium text-muted-foreground">
								Nenhuma pergunta encontrada
							</p>
							<p className="mt-1 text-xs text-muted-foreground/60">
								Tente buscar por outros termos ou{" "}
								<button
									type="button"
									onClick={handleClearAll}
									className="font-medium text-primary hover:underline underline-offset-4"
								>
									limpe os filtros
								</button>
							</p>
						</div>
					)}
				</div>

				{/* Bottom CTA */}
				<motion.div
					initial={{ opacity: 0, y: 16 }}
					animate={isInView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.5, delay: 0.3 }}
					className="mx-auto mt-14 max-w-xl text-center"
				>
					<div className="rounded-2xl border border-border/60 bg-muted/20 p-6 sm:p-8">
						<p className="text-base font-semibold text-foreground">
							Ainda tem dúvidas?
						</p>
						<p className="mt-2 text-sm text-muted-foreground leading-relaxed">
							Nossa equipe técnica brasileira responde em até 24h.
							Planos Business+ com suporte prioritário em até 4h.
						</p>
						<div className="mt-5 flex flex-col items-center gap-2.5 sm:flex-row sm:justify-center sm:gap-3">
							<Button
								size="lg"
								asChild
								className="h-11 px-6 text-sm shadow-lg shadow-primary/20 hover:shadow-xl hover:shadow-primary/25 transition-all duration-300"
							>
								<Link href="/contato">
									Falar com Especialista
									<ArrowRightIcon
										className="size-4"
										data-icon="inline-end"
									/>
								</Link>
							</Button>
							<Button
								variant="outline"
								size="lg"
								asChild
								className="h-11 px-6 text-sm"
							>
								<a
									href="https://api.zedaapi.com/docs"
									target="_blank"
									rel="noopener noreferrer"
								>
									Ver Documentação
								</a>
							</Button>
						</div>
					</div>
				</motion.div>
			</div>

			{/* JSON-LD Structured Data for SEO */}
			<script
				type="application/ld+json"
				// biome-ignore lint/security/noDangerouslySetInnerHtml: safe static JSON-LD
				dangerouslySetInnerHTML={{
					__html: JSON.stringify({
						"@context": "https://schema.org",
						"@type": "FAQPage",
						mainEntity: faqCategories.flatMap((cat) =>
							cat.items.map((item) => ({
								"@type": "Question",
								name: item.question,
								acceptedAnswer: {
									"@type": "Answer",
									text: item.answer,
								},
							})),
						),
					}),
				}}
			/>
		</section>
	);
}
