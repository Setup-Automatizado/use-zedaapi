import type { Metadata } from "next";
import { LegalLayout } from "@/components/landing/legal-layout";

export const metadata: Metadata = {
	title: "Termos de Uso - Zé da API",
	description:
		"Termos e Condições de Uso da plataforma Zé da API Manager. Leia atentamente antes de utilizar nossos serviços de WhatsApp API.",
};

export default function TermosDeUsoPage() {
	return (
		<LegalLayout title="Termos de Uso" lastUpdated="Fevereiro de 2026">
			<p>
				Bem-vindo ao <strong>Zé da API Manager</strong>. Estes Termos de
				Uso (&quot;Termos&quot;) constituem um contrato vinculante que
				regula o acesso e a utilização da plataforma de gerenciamento de
				WhatsApp API oferecida por{" "}
				<strong>Setup Automatizado Ltda</strong> (&quot;Empresa&quot;,
				&quot;nós&quot;, &quot;nosso&quot; ou &quot;Zé da API&quot;),
				pessoa jurídica de direito privado, inscrita no CNPJ/MF sob o
				n.o 54.246.473/0001-00, com sede na cidade do Rio de Janeiro,
				Estado do Rio de Janeiro, República Federativa do Brasil.
			</p>
			<p>
				Ao acessar ou utilizar a plataforma Zé da API Manager
				(&quot;Plataforma&quot; ou &quot;Serviço&quot;), disponível em{" "}
				<a
					href="https://zedaapi.com"
					target="_blank"
					rel="noopener noreferrer"
				>
					zedaapi.com
				</a>
				, você (&quot;Usuário&quot;, &quot;Cliente&quot; ou
				&quot;você&quot;) declara ter lido, compreendido e concordado
				integralmente com estes Termos. Caso não concorde com qualquer
				disposição, não utilize a Plataforma.
			</p>
			<p>
				Estes Termos foram elaborados em conformidade com a legislação
				brasileira vigente, incluindo a <strong>Lei 13.709/2018</strong>{" "}
				(LGPD), o{" "}
				<strong>Marco Civil da Internet (Lei 12.965/2014)</strong>, o{" "}
				<strong>Código de Defesa do Consumidor (Lei 8.078/1990)</strong>
				, o <strong>Decreto 7.962/2013</strong> (Comércio Eletrônico), a{" "}
				<strong>Lei 12.737/2012</strong> (Delitos Informáticos), a{" "}
				<strong>Lei 9.610/1998</strong> (Direitos Autorais) e a{" "}
				<strong>Lei 9.279/1996</strong> (Propriedade Industrial).
			</p>

			<hr />

			{/* 1 */}
			<h2 id="definicoes">1. Definições e Interpretação</h2>
			<p>
				<strong>1.1.</strong> Para os fins destes Termos, consideram-se
				as seguintes definições:
			</p>
			<p>
				<strong>1.1.1.</strong> <strong>Plataforma</strong>: o software
				SaaS (Software as a Service) denominado &quot;Zé da API
				Manager&quot;, incluindo todas as funcionalidades, APIs,
				webhooks, documentação técnica, interfaces de usuário, painel
				administrativo e quaisquer atualizações ou novas versões
				disponibilizadas pela Empresa.
			</p>
			<p>
				<strong>1.1.2.</strong> <strong>Instância WhatsApp</strong>:
				cada conexão individual configurada pelo Usuário para
				comunicação via WhatsApp Business API através da Plataforma,
				vinculada a um número de telefone específico.
			</p>
			<p>
				<strong>1.1.3.</strong> <strong>API</strong>: a interface de
				programação de aplicações (Application Programming Interface)
				fornecida pela Plataforma, incluindo endpoints REST, webhooks e
				demais mecanismos de integração programática.
			</p>
			<p>
				<strong>1.1.4.</strong> <strong>Webhook</strong>: mecanismo de
				notificação em tempo real configurado pelo Usuário para
				recebimento automático de eventos da Plataforma em um endpoint
				HTTP(S) de sua propriedade.
			</p>
			<p>
				<strong>1.1.5.</strong> <strong>Dados do Usuário</strong>: todas
				as informações, mensagens, mídias, metadados, configurações e
				demais conteúdos inseridos, processados ou gerados pelo Usuário
				através da Plataforma.
			</p>
			<p>
				<strong>1.1.6.</strong> <strong>WhatsApp/Meta</strong>: WhatsApp
				LLC e/ou Meta Platforms, Inc., e suas subsidiárias e afiliadas,
				provedoras da infraestrutura subjacente de mensageria utilizada
				pela Plataforma.
			</p>
			<p>
				<strong>1.1.7.</strong> <strong>Conta</strong>: o cadastro
				individual do Usuário na Plataforma, protegido por credenciais
				de autenticação (e-mail e senha) e, opcionalmente, por
				autenticação de dois fatores (2FA).
			</p>
			<p>
				<strong>1.1.8.</strong> <strong>Plano</strong>: a modalidade de
				assinatura contratada pelo Usuário, que define os limites de
				instâncias, funcionalidades e níveis de serviço disponíveis.
			</p>
			<p>
				<strong>1.1.9.</strong> <strong>Token de API</strong>: chave
				criptográfica de autenticação utilizada para acesso programático
				à API da Plataforma.
			</p>
			<p>
				<strong>1.1.10.</strong> <strong>NFS-e</strong>: Nota Fiscal de
				Serviço Eletrônica, documento fiscal emitido eletronicamente
				pela Empresa para cada pagamento realizado.
			</p>
			<p>
				<strong>1.2.</strong> Termos no singular incluem o plural e
				vice-versa. Referências a &quot;dias&quot; significam dias
				corridos, salvo quando expressamente indicado &quot;dias
				úteis&quot;. Referências a legislação incluem suas alterações,
				regulamentações e normas complementares.
			</p>

			{/* 2 */}
			<h2 id="aceitacao">2. Aceitação dos Termos</h2>
			<p>
				<strong>2.1.</strong> Ao criar uma conta, acessar ou utilizar
				qualquer funcionalidade da Plataforma, o Usuário manifesta sua
				aceitação integral e irrestrita a estes Termos, à{" "}
				<a href="/politica-de-privacidade">Política de Privacidade</a>,
				à <a href="/politica-de-cookies">Política de Cookies</a> e a
				quaisquer políticas complementares publicadas pela Empresa, que
				constituem parte integrante e indissociável destes Termos.
			</p>
			<p>
				<strong>2.2.</strong> O Usuário declara ser maior de 18
				(dezoito) anos e possuir capacidade civil plena para celebrar
				este contrato, nos termos dos artigos 1o e 2o do Código Civil
				Brasileiro (Lei 10.406/2002). No caso de pessoa jurídica, o
				representante declara possuir poderes legais e contratuais
				suficientes para vincular a organização aos presentes Termos.
			</p>
			<p>
				<strong>2.3.</strong> A utilização da Plataforma por menores de
				18 (dezoito) anos é expressamente vedada, em conformidade com o
				artigo 14 da LGPD e com o Estatuto da Criança e do Adolescente
				(Lei 8.069/1990).
			</p>
			<p>
				<strong>2.4.</strong> A aceitação destes Termos constitui
				contrato válido e vinculante entre o Usuário e a Empresa, nos
				termos do artigo 104 do Código Civil Brasileiro e do artigo 7o,
				inciso XIII, do Marco Civil da Internet.
			</p>

			{/* 3 */}
			<h2 id="descricao-servico">3. Descrição do Serviço</h2>
			<p>
				<strong>3.1.</strong> A Plataforma Zé da API Manager é um
				serviço SaaS de gerenciamento de instâncias WhatsApp Business
				API que permite ao Usuário:
			</p>
			<ul>
				<li>
					Criar, configurar e gerenciar múltiplas instâncias WhatsApp
					com conexão via QR Code ou pareamento;
				</li>
				<li>
					Enviar e receber mensagens de texto, mídia (imagens, vídeos,
					áudios, documentos), localização, contatos e mensagens
					interativas via API REST;
				</li>
				<li>
					Configurar webhooks para recebimento de eventos em tempo
					real, incluindo mensagens recebidas, status de entrega,
					leitura e presença;
				</li>
				<li>
					Monitorar status de conexão, métricas de uso, entregas e
					qualidade de serviço em painel de controle;
				</li>
				<li>
					Integrar com sistemas de terceiros através de API REST
					documentada com autenticação por token;
				</li>
				<li>
					Gerenciar contatos, grupos, listas de transmissão e
					configurações de negócio no WhatsApp;
				</li>
				<li>
					Acessar painel administrativo para gerenciamento de conta,
					assinaturas, chaves de API e configurações.
				</li>
			</ul>
			<p>
				<strong>3.2.</strong> A Empresa atua exclusivamente como{" "}
				<strong>provedora de tecnologia intermediária</strong>. O
				Serviço depende da infraestrutura da Meta (WhatsApp), e
				funcionalidades podem ser afetadas, limitadas, modificadas ou
				interrompidas por alterações, indisponibilidades, atualizações
				de política ou restrições impostas pela Meta a seu exclusivo e
				absoluto critério, sem que isso gere qualquer responsabilidade
				para a Empresa.
			</p>
			<p>
				<strong>3.3.</strong> A Empresa reserva-se o direito de
				adicionar, modificar, suspender ou descontinuar funcionalidades
				da Plataforma, mediante notificação prévia de 30 (trinta) dias
				para alterações substanciais que afetem negativamente o uso do
				Serviço pelo Usuário. Melhorias, correções de segurança e
				atualizações de manutenção podem ser implementadas sem aviso
				prévio.
			</p>
			<p>
				<strong>3.4.</strong> A Empresa não garante que a Plataforma
				atenderá a requisitos específicos, particulares ou
				extraordinários do Usuário que não estejam expressamente
				previstos na documentação do Serviço.
			</p>

			{/* 4 */}
			<h2 id="cadastro-conta">4. Cadastro e Conta</h2>
			<p>
				<strong>4.1.</strong> Para utilizar a Plataforma, o Usuário deve
				criar uma conta fornecendo informações verídicas, completas e
				atualizadas, incluindo nome completo, endereço de e-mail válido
				e, quando aplicável, dados da pessoa jurídica (razão social,
				CNPJ e endereço fiscal).
			</p>
			<p>
				<strong>4.2.</strong> O Usuário é o único e exclusivo
				responsável por manter a confidencialidade de suas credenciais
				de acesso (e-mail, senha, tokens de API, chaves de autenticação)
				e por todas as atividades realizadas em sua conta, inclusive
				aquelas realizadas por terceiros que tenham obtido acesso, com
				ou sem autorização.
			</p>
			<p>
				<strong>4.3.</strong> A verificação de e-mail é obrigatória para
				ativação da conta. A Plataforma oferece autenticação de dois
				fatores (2FA) como camada adicional de segurança, cuja ativação
				é <strong>enfaticamente recomendada</strong> pela Empresa.
			</p>
			<p>
				<strong>4.4.</strong> A Plataforma pode utilizar sistema de
				lista de espera (waitlist) para controle de acesso a novos
				cadastros, a critério exclusivo da Empresa.
			</p>
			<p>
				<strong>4.5.</strong> O Usuário deve notificar a Empresa
				imediatamente ao tomar conhecimento de qualquer uso não
				autorizado de sua conta, comprometimento de credenciais ou
				qualquer outra violação de segurança, pelo e-mail{" "}
				<a href="mailto:suporte@zedaapi.com">suporte@zedaapi.com</a>.
			</p>
			<p>
				<strong>4.6.</strong> A Empresa não será responsável por perdas,
				danos, custos ou despesas de qualquer natureza decorrentes do
				descumprimento das obrigações de segurança previstas nesta seção
				por parte do Usuário, incluindo, mas não se limitando a, acessos
				não autorizados decorrentes de negligência na guarda de
				credenciais.
			</p>
			<p>
				<strong>4.7.</strong> A Empresa reserva-se o direito de recusar
				ou cancelar cadastros que contenham informações falsas,
				incompletas ou que violem estes Termos.
			</p>

			{/* 5 */}
			<h2 id="planos-precos">5. Planos e Preços</h2>
			<p>
				<strong>5.1.</strong> A Plataforma oferece diferentes planos de
				assinatura (&quot;Planos&quot;), com os seguintes valores e
				limites de referência:
			</p>
			<table>
				<thead>
					<tr>
						<th>Plano</th>
						<th>Valor Mensal</th>
						<th>Instâncias</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>Starter</td>
						<td>R$ 9,00</td>
						<td>1</td>
					</tr>
					<tr>
						<td>Growth</td>
						<td>R$ 29,00</td>
						<td>5</td>
					</tr>
					<tr>
						<td>Professional</td>
						<td>R$ 59,00</td>
						<td>15</td>
					</tr>
					<tr>
						<td>Business</td>
						<td>R$ 99,00</td>
						<td>30</td>
					</tr>
					<tr>
						<td>Enterprise</td>
						<td>R$ 149,00</td>
						<td>60</td>
					</tr>
					<tr>
						<td>Ultimate</td>
						<td>R$ 199,00</td>
						<td>100+</td>
					</tr>
				</tbody>
			</table>
			<p>
				<strong>5.2.</strong> Os valores, limites e funcionalidades de
				cada Plano estão detalhados na página de preços da Plataforma e
				podem ser atualizados mediante aviso prévio de 30 (trinta) dias.
				Aumentos de preço não se aplicam ao ciclo de faturamento já
				contratado.
			</p>
			<p>
				<strong>5.3.</strong> Planos anuais podem oferecer descontos,
				conforme divulgado na Plataforma. O desconto é aplicável apenas
				no ato da contratação anual e não é retroativo.
			</p>
			<p>
				<strong>5.4.</strong> A Empresa pode oferecer período de teste
				gratuito (&quot;Trial&quot;), com duração e condições definidas
				no momento da contratação. Ao término do Trial, o Usuário deverá
				contratar um Plano pago para continuar utilizando a Plataforma.
				A não contratação resultará na suspensão do acesso às
				funcionalidades restritas.
			</p>
			<p>
				<strong>5.5.</strong> A Empresa reserva-se o direito de oferecer
				planos promocionais, descontos por tempo limitado ou condições
				especiais, que serão regidos por termos específicos comunicados
				no momento da oferta.
			</p>
			<p>
				<strong>5.6.</strong> O upgrade de Plano terá efeito imediato,
				com cobrança proporcional (pro rata) ao período restante. O
				downgrade terá efeito no próximo ciclo de faturamento, sem
				direito a reembolso da diferença.
			</p>

			{/* 6 */}
			<h2 id="pagamento">6. Pagamento e Faturamento</h2>
			<p>
				<strong>6.1.</strong> Os pagamentos são processados por meio de
				provedores de pagamento terceirizados homologados:
			</p>
			<ul>
				<li>
					<strong>Stripe, Inc.</strong> (EUA): para pagamentos com
					cartão de crédito internacional e métodos de pagamento
					internacionais, em conformidade com o padrão PCI-DSS
					(Payment Card Industry Data Security Standard);
				</li>
				<li>
					<strong>PIX via Sicredi</strong> (Brasil): para
					transferências instantâneas via Sistema de Pagamentos
					Instantâneos do Banco Central do Brasil, com liquidação
					imediata;
				</li>
				<li>
					<strong>Boleto Bancário Híbrido via Sicredi</strong>{" "}
					(Brasil): para pagamentos via boleto bancário com QR Code
					PIX integrado, com compensação em até 3 (três) dias úteis.
				</li>
			</ul>
			<p>
				<strong>6.2.</strong> A Empresa <strong>NÃO armazena</strong>{" "}
				dados completos de cartão de crédito em seus servidores. Todas
				as informações de pagamento são processadas, tokenizadas e
				armazenadas exclusivamente pelos provedores de pagamento, em
				conformidade com o padrão PCI-DSS Level 1.
			</p>
			<p>
				<strong>6.3.</strong> A cobrança é realizada de forma recorrente
				(mensal ou anual, conforme o Plano contratado), na data de
				aniversário da contratação. O Usuário autoriza expressamente a
				cobrança recorrente ao contratar o Plano.
			</p>
			<p>
				<strong>6.4.</strong> Em caso de falha no pagamento, a Empresa
				realizará até 3 (três) tentativas adicionais de cobrança nos
				dias 3, 5 e 7 após a data original. Persistindo a inadimplência
				após a terceira tentativa, o acesso à Plataforma será suspenso
				até a regularização do pagamento, sem exclusão imediata dos
				dados.
			</p>
			<p>
				<strong>6.5.</strong> Sobre valores em atraso incidirão: (a)
				multa moratória de 2% (dois por cento); (b) juros de mora de 1%
				(um por cento) ao mês, calculados pro rata die; e (c) correção
				monetária pelo IPCA/IBGE ou índice que o substitua.
			</p>
			<p>
				<strong>6.6.</strong> A Empresa emitirá Nota Fiscal de Serviço
				Eletrônica (NFS-e) para cada pagamento confirmado, conforme
				legislação tributária municipal e federal aplicável, incluindo o
				Código Tributário Nacional (Lei 5.172/1966) e a Lei Complementar
				116/2003. A NFS-e será enviada ao e-mail cadastrado pelo
				Usuário.
			</p>
			<p>
				<strong>6.7.</strong> Todos os valores são expressos em Reais
				(BRL) e incluem os impostos devidos. Para pagamentos
				internacionais via Stripe, a conversão cambial será realizada
				pelo provedor de pagamento conforme suas taxas vigentes.
			</p>
			<p>
				<strong>6.8. Direito de Arrependimento.</strong> Em conformidade
				com o artigo 49 do Código de Defesa do Consumidor (Lei
				8.078/1990), o Usuário consumidor pessoa física que contratar o
				Serviço fora do estabelecimento comercial (contratação online)
				poderá exercer o direito de arrependimento no prazo de 7 (sete)
				dias corridos a contar da contratação ou do recebimento da
				confirmação, o que ocorrer por último, com direito ao reembolso
				integral dos valores pagos, incluindo eventuais encargos.
			</p>
			<p>
				<strong>6.9.</strong> Solicitações de reembolso fora do prazo do
				artigo 49 do CDC serão analisadas individualmente pela Empresa,
				a seu exclusivo critério, e poderão ser concedidas em casos
				excepcionais e devidamente justificados.
			</p>

			{/* 7 */}
			<h2 id="uso-permitido">7. Uso Permitido e Restrições</h2>
			<p>
				<strong>7.1.</strong> A Plataforma deve ser utilizada
				exclusivamente para fins lícitos e em estrita conformidade com a
				legislação brasileira vigente, incluindo, mas não se limitando
				a:
			</p>
			<ul>
				<li>
					Lei 12.965/2014 (Marco Civil da Internet) e seu Decreto
					regulamentador 8.771/2016;
				</li>
				<li>
					Lei 13.709/2018 (Lei Geral de Proteção de Dados Pessoais -
					LGPD);
				</li>
				<li>Lei 8.078/1990 (Código de Defesa do Consumidor - CDC);</li>
				<li>Decreto 7.962/2013 (Comércio Eletrônico);</li>
				<li>
					Lei 12.737/2012 (Delitos Informáticos - Lei Carolina
					Dieckmann);
				</li>
				<li>
					Lei 12.735/2012 (Tipificação de Condutas Realizadas Mediante
					Uso de Sistemas Eletrônicos);
				</li>
				<li>Lei 7.716/1989 (Crimes de Preconceito e Discriminação);</li>
				<li>
					Código Penal Brasileiro (Decreto-Lei 2.848/1940), em
					especial os dispositivos relativos a estelionato, falsidade
					ideológica e crimes contra a honra.
				</li>
			</ul>
			<p>
				<strong>7.2.</strong> É expressamente proibido ao Usuário
				utilizar a Plataforma para:
			</p>
			<ul>
				<li>
					Envio de mensagens não solicitadas (spam), correntes,
					pirâmides financeiras ou qualquer forma de comunicação em
					massa sem o consentimento prévio e expresso dos
					destinatários;
				</li>
				<li>
					Realizar engenharia reversa, descompilar, desmontar,
					desobfuscar ou tentar de qualquer forma acessar o
					código-fonte, algoritmos ou estrutura interna da Plataforma;
				</li>
				<li>
					Sublicenciar, revender, redistribuir, alugar ou
					comercializar o acesso à Plataforma ou a qualquer de suas
					funcionalidades sem autorização expressa e por escrito da
					Empresa;
				</li>
				<li>
					Utilizar a Plataforma para atividades ilegais, fraudulentas,
					enganosas ou que violem direitos de terceiros;
				</li>
				<li>
					Transmitir vírus, trojans, worms, malware, ransomware ou
					qualquer código malicioso ou destrutivo através da
					Plataforma;
				</li>
				<li>
					Sobrecarregar, estressar ou interferir intencionalmente na
					infraestrutura da Plataforma, realizar ataques de negação de
					serviço (DoS/DDoS), scraping não autorizado ou exploração de
					vulnerabilidades;
				</li>
				<li>
					Coletar, armazenar ou processar dados pessoais de terceiros
					sem o devido consentimento ou base legal adequada, conforme
					exigido pela LGPD;
				</li>
				<li>
					Disseminar discurso de ódio, conteúdo discriminatório,
					pornográfico (especialmente envolvendo menores), violento,
					difamatório, calunioso ou que incite à prática de crimes;
				</li>
				<li>
					Utilizar a Plataforma para assédio, intimidação, ameaça,
					perseguição (stalking) ou qualquer forma de violência
					psicológica contra terceiros;
				</li>
				<li>
					Compartilhar tokens de API, credenciais de acesso ou chaves
					de autenticação com terceiros não autorizados;
				</li>
				<li>
					Contornar, burlar ou desabilitar mecanismos de segurança,
					rate limits, restrições de acesso ou quaisquer medidas de
					proteção implementadas pela Plataforma;
				</li>
				<li>
					Utilizar a Plataforma para fins de competitive intelligence,
					benchmarking ou atividades que visem beneficiar concorrentes
					diretos da Empresa.
				</li>
			</ul>
			<p>
				<strong>7.3.</strong> A violação das restrições previstas nesta
				seção poderá resultar na suspensão imediata da conta, rescisão
				contratual sem direito a reembolso, e responsabilização civil e
				criminal do Usuário, conforme legislação aplicável.
			</p>

			{/* 8 */}
			<h2 id="uso-aceitavel-whatsapp">
				8. Política de Uso Aceitável do WhatsApp
			</h2>
			<p>
				<strong>
					ESTA SEÇÃO CONTÉM DISPOSIÇÕES DE SUMA IMPORTÂNCIA E DEVE SER
					LIDA COM ESPECIAL E REDOBRADA ATENÇÃO PELO USUÁRIO.
				</strong>
			</p>
			<p>
				<strong>8.1.</strong> O Usuário reconhece e concorda que a
				utilização da Plataforma está sujeita, além destes Termos e da
				legislação brasileira, às políticas, termos e condições do
				WhatsApp e da Meta Platforms, Inc., incluindo, mas não se
				limitando a:
			</p>
			<ul>
				<li>
					<a
						href="https://www.whatsapp.com/legal/business-terms"
						target="_blank"
						rel="noopener noreferrer"
					>
						Termos de Serviço do WhatsApp Business
					</a>
					;
				</li>
				<li>
					<a
						href="https://www.whatsapp.com/legal/commerce-policy"
						target="_blank"
						rel="noopener noreferrer"
					>
						Política Comercial do WhatsApp
					</a>
					;
				</li>
				<li>
					<a
						href="https://www.whatsapp.com/legal/business-policy"
						target="_blank"
						rel="noopener noreferrer"
					>
						Política de Mensagens do WhatsApp Business
					</a>
					;
				</li>
				<li>
					<a
						href="https://developers.facebook.com/terms/"
						target="_blank"
						rel="noopener noreferrer"
					>
						Termos de Serviço da Meta para Desenvolvedores
					</a>
					;
				</li>
				<li>
					<a
						href="https://developers.facebook.com/devpolicy/"
						target="_blank"
						rel="noopener noreferrer"
					>
						Política de Dados da Meta Platform
					</a>
					.
				</li>
			</ul>
			<p>
				<strong>8.2.</strong> O Usuário é o{" "}
				<strong>ÚNICO E EXCLUSIVO RESPONSÁVEL</strong> pelo cumprimento
				integral e irrestrito de todas as políticas, termos de uso,
				diretrizes, regras e restrições do WhatsApp e da Meta. A Empresa
				atua{" "}
				<strong>
					exclusivamente como provedora de tecnologia intermediária
				</strong>{" "}
				e <strong>NÃO exerce, em nenhuma hipótese</strong>, controle,
				supervisão, moderação, edição ou verificação sobre o conteúdo
				das mensagens enviadas, recebidas ou processadas pelo Usuário
				através da Plataforma.
			</p>
			<p>
				<strong>8.3.</strong> São expressamente proibidos e passíveis de
				suspensão imediata e irrevogável da conta, sem direito a
				reembolso:
			</p>
			<ul>
				<li>
					Envio de mensagens em massa não solicitadas (bulk messaging)
					sem prévio opt-in documentado dos destinatários;
				</li>
				<li>
					Envio de mensagens com conteúdo spam, golpes, fraudes,
					phishing, engenharia social ou esquemas financeiros ilegais;
				</li>
				<li>
					Assédio, intimidação, ameaça, chantagem, extorsão ou
					perseguição de qualquer pessoa;
				</li>
				<li>
					Distribuição de conteúdo ilegal, pornográfico envolvendo
					menores, gore, violento extremo ou que promova ódio,
					discriminação ou intolerância;
				</li>
				<li>
					Utilização de números de telefone obtidos sem consentimento,
					adquiridos ilegalmente ou provenientes de bases de dados não
					autorizadas;
				</li>
				<li>
					Envio de mensagens que violem leis de proteção ao
					consumidor, privacidade, proteção de dados ou direitos de
					terceiros;
				</li>
				<li>
					Automação que viole os limites de rate-limit, quality rating
					ou políticas de envio do WhatsApp;
				</li>
				<li>
					Criação de contas falsas, utilização de identidades
					fictícias ou impersonação de terceiros;
				</li>
				<li>
					Utilização da Plataforma para fins políticos de
					desinformação, manipulação de eleições ou disseminação de
					fake news;
				</li>
				<li>
					Comercialização de produtos ou serviços proibidos pelas
					políticas do WhatsApp, incluindo armas, drogas, medicamentos
					sem receita, produtos falsificados ou conteúdo
					regulamentado.
				</li>
			</ul>
			<p>
				<strong>8.4.</strong> A Empresa reserva-se o direito de
				suspender ou encerrar imediatamente o acesso do Usuário à
				Plataforma caso tome conhecimento, por qualquer meio, de
				violações reais ou potenciais às políticas do WhatsApp/Meta, sem
				necessidade de aviso prévio, notificação judicial ou
				extrajudicial, e sem direito a reembolso de valores já pagos.
			</p>
			<p>
				<strong>8.5.</strong>{" "}
				<strong>
					A EMPRESA EXIME-SE INTEGRAL, IRRESTRITA E IRREVOGAVELMENTE
					DE QUALQUER RESPONSABILIDADE
				</strong>
				, direta ou indireta, solidária ou subsidiária, por:
			</p>
			<ul>
				<li>
					Banimento, suspensão, restrição, limitação ou encerramento
					de contas WhatsApp do Usuário pela Meta, por qualquer
					motivo;
				</li>
				<li>
					Consequências legais, financeiras, reputacionais,
					regulatórias ou de qualquer outra natureza decorrentes do
					uso indevido, irregular ou ilegal da Plataforma pelo
					Usuário;
				</li>
				<li>
					Alterações unilaterais nas políticas, APIs, termos de
					serviço, funcionalidades, limites ou condições do
					WhatsApp/Meta, inclusive descontinuação de serviços;
				</li>
				<li>
					Perdas, danos, lucros cessantes ou danos emergentes
					decorrentes da interrupção, degradação ou indisponibilidade
					dos serviços do WhatsApp por parte da Meta;
				</li>
				<li>
					Violações de políticas do WhatsApp/Meta cometidas pelo
					Usuário ou por terceiros que atuem em seu nome ou por sua
					conta;
				</li>
				<li>
					Multas, penalidades, sanções ou ações judiciais impostas
					pela Meta, por autoridades reguladoras ou por terceiros em
					razão do comportamento do Usuário;
				</li>
				<li>
					Indisponibilidade de funcionalidades específicas do WhatsApp
					em determinadas regiões, dispositivos ou versões de
					aplicativo.
				</li>
			</ul>
			<p>
				<strong>8.6.</strong> O Usuário concorda em indenizar, defender
				e manter a Empresa, seus sócios, diretores, funcionários,
				agentes, prepostos e prestadores de serviço completamente
				isentos e indenes de quaisquer reclamações, ações judiciais ou
				administrativas, perdas, danos, custos e despesas (incluindo
				honorários advocatícios e custas processuais) decorrentes de seu
				descumprimento das políticas do WhatsApp/Meta ou de qualquer
				violação descrita nesta seção.
			</p>

			{/* 9 */}
			<h2 id="api-webhooks">9. API e Webhooks</h2>
			<p>
				<strong>9.1.</strong> A Plataforma disponibiliza API REST
				documentada para integração programática. O acesso à API está
				condicionado à utilização de tokens de autenticação (API Keys)
				fornecidos pela Plataforma, com mínimo de 16 caracteres de
				complexidade criptográfica.
			</p>
			<p>
				<strong>9.2.</strong> Limites de requisição (rate limits) são
				aplicados por Plano e por instância, conforme documentação
				técnica disponível em{" "}
				<a
					href="https://zedaapi.com/docs"
					target="_blank"
					rel="noopener noreferrer"
				>
					zedaapi.com/docs
				</a>
				. O excesso de requisições resultará em respostas HTTP 429 (Too
				Many Requests). Tentativas sistemáticas de contornar rate limits
				configuram violação destes Termos.
			</p>
			<p>
				<strong>9.3.</strong> Webhooks configurados pelo Usuário devem
				estar acessíveis publicamente, responder em tempo hábil (máximo
				de 20 segundos), utilizar protocolo HTTPS com certificado
				SSL/TLS válido e retornar código HTTP 2xx para confirmação de
				recebimento.
			</p>
			<p>
				<strong>9.4.</strong> A Empresa não garante a entrega de
				webhooks em caso de indisponibilidade, instabilidade ou
				configuração inadequada do endpoint do Usuário. Eventos não
				entregues são reenviados automaticamente com backoff exponencial
				e fila de dead-letter, conforme limites do Plano contratado.
			</p>
			<p>
				<strong>9.5.</strong> A Plataforma utiliza filas de mensagens
				com ordenação FIFO (First In, First Out) por instância para
				garantir a sequência correta de envio. O Usuário não deve contar
				com entrega instantânea, reconhecendo que podem haver latências
				inerentes ao processamento assíncrono.
			</p>
			<p>
				<strong>9.6.</strong> O Usuário é responsável pela segurança de
				seus tokens de API e deve tratá-los como credenciais sensíveis.
				Tokens comprometidos devem ser revogados e regenerados
				imediatamente. A Empresa não se responsabiliza por acessos
				indevidos decorrentes do comprometimento de tokens por
				negligência do Usuário.
			</p>

			{/* 10 */}
			<h2 id="propriedade-intelectual">10. Propriedade Intelectual</h2>
			<p>
				<strong>10.1.</strong> A Plataforma, incluindo, mas não se
				limitando a, seu código-fonte, código-objeto, arquitetura,
				design, layout, interfaces de usuário, documentação técnica e
				comercial, marcas, logotipos, ícones, nomes de domínio e demais
				elementos visuais, gráficos e funcionais, são de propriedade
				exclusiva da Empresa e protegidos pela legislação brasileira de
				propriedade intelectual, incluindo:
			</p>
			<ul>
				<li>Lei 9.610/1998 (Direitos Autorais e Direitos Conexos);</li>
				<li>Lei 9.279/1996 (Propriedade Industrial);</li>
				<li>
					Lei 9.609/1998 (Proteção da Propriedade Intelectual de
					Programa de Computador).
				</li>
			</ul>
			<p>
				<strong>10.2.</strong> A contratação de Plano confere ao Usuário
				licença <strong>limitada</strong>,{" "}
				<strong>não exclusiva</strong>, <strong>intransferível</strong>,{" "}
				<strong>não sublicenciável</strong> e <strong>revogável</strong>{" "}
				para utilizar a Plataforma durante a vigência da assinatura,
				exclusivamente para os fins previstos nestes Termos. Esta
				licença não confere ao Usuário qualquer direito de propriedade
				sobre a Plataforma ou seus componentes.
			</p>
			<p>
				<strong>10.3.</strong> O Usuário mantém a titularidade integral
				sobre os dados e conteúdos por ele inseridos na Plataforma. Ao
				utilizar o Serviço, o Usuário concede à Empresa licença
				limitada, não exclusiva e pelo prazo de vigência do contrato
				para processar, armazenar e transmitir tais dados exclusivamente
				para a prestação do Serviço contratado.
			</p>
			<p>
				<strong>10.4.</strong> A marca &quot;Zé da API&quot;, o logotipo
				e suas variações são marcas de propriedade da Empresa. Qualquer
				uso não autorizado, reprodução, imitação ou associação indevida
				constitui violação de direitos de propriedade industrial,
				sujeita às sanções previstas na Lei 9.279/1996.
			</p>
			<p>
				<strong>10.5.</strong> A Empresa poderá utilizar o nome e
				logotipo do Usuário pessoa jurídica como referência comercial
				(case study, lista de clientes) em materiais de marketing, salvo
				oposição expressa comunicada por escrito.
			</p>

			{/* 11 */}
			<h2 id="dados-privacidade">11. Dados e Privacidade</h2>
			<p>
				<strong>11.1.</strong> O tratamento de dados pessoais pela
				Empresa é regido integralmente pela{" "}
				<a href="/politica-de-privacidade">Política de Privacidade</a>,
				que constitui parte integrante e indissociável destes Termos.
			</p>
			<p>
				<strong>11.2.</strong> A Empresa atua como{" "}
				<strong>Controladora</strong> dos dados cadastrais, financeiros
				e de uso da Plataforma, e como <strong>Operadora</strong> dos
				dados pessoais de terceiros processados em nome do Usuário
				através das instâncias WhatsApp, nos termos dos artigos 5o,
				incisos VI e VII, da Lei 13.709/2018 (LGPD).
			</p>
			<p>
				<strong>11.3.</strong> Os direitos dos titulares de dados
				pessoais previstos nos artigos 17 a 22 da LGPD podem ser
				exercidos conforme detalhado na página{" "}
				<a href="/lgpd">LGPD - Direitos do Titular</a> e na página de{" "}
				<a href="/exclusao-de-dados">Exclusão de Dados</a>.
			</p>
			<p>
				<strong>11.4.</strong> O Usuário é o{" "}
				<strong>único e exclusivo responsável</strong> por obter e
				manter as bases legais adequadas (consentimento, legítimo
				interesse, execução contratual ou demais hipóteses do artigo 7o
				da LGPD) para o tratamento de dados pessoais de terceiros
				através da Plataforma. O Usuário declara que todos os dados de
				terceiros processados por meio de suas instâncias foram
				coletados de forma lícita e com base legal válida.
			</p>
			<p>
				<strong>11.5.</strong> Em caso de incidente de segurança que
				envolva dados pessoais processados através da Plataforma, a
				Empresa notificará o Usuário em prazo razoável, conforme o
				artigo 48 da LGPD, para que este possa adotar as medidas
				cabíveis perante os titulares e a Autoridade Nacional de
				Proteção de Dados (ANPD).
			</p>

			{/* 12 */}
			<h2 id="disponibilidade-sla">12. Disponibilidade e SLA</h2>
			<p>
				<strong>12.1.</strong> A Empresa envidará seus melhores esforços
				comercialmente razoáveis para manter a Plataforma disponível de
				forma contínua. Os níveis de disponibilidade garantida (SLA -
				Service Level Agreement) variam conforme o Plano contratado:
			</p>
			<table>
				<thead>
					<tr>
						<th>Plano</th>
						<th>SLA de Disponibilidade</th>
						<th>Tempo Máximo de Indisponibilidade/Mês</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>Starter / Growth</td>
						<td>99,5%</td>
						<td>3h 39min</td>
					</tr>
					<tr>
						<td>Professional / Business</td>
						<td>99,9%</td>
						<td>43min 28s</td>
					</tr>
					<tr>
						<td>Enterprise / Ultimate</td>
						<td>99,99%</td>
						<td>4min 21s</td>
					</tr>
				</tbody>
			</table>
			<p>
				<strong>12.2.</strong> Os percentuais de SLA são calculados
				mensalmente, excluindo-se: (a) janelas de manutenção programada,
				comunicadas com no mínimo 48 (quarenta e oito) horas de
				antecedência por e-mail ou notificação na Plataforma; (b)
				eventos de força maior; (c) indisponibilidades decorrentes de
				serviços de terceiros (WhatsApp/Meta, provedores de
				infraestrutura); e (d) falhas causadas por ações do próprio
				Usuário.
			</p>
			<p>
				<strong>12.3.</strong> Em caso de descumprimento do SLA
				atribuível exclusivamente à Empresa, o Usuário poderá solicitar
				créditos proporcionais ao período de indisponibilidade,
				limitados a 30% (trinta por cento) do valor mensal do Plano no
				mês afetado. Tais créditos constituem a única e exclusiva
				compensação disponível.
			</p>
			<p>
				<strong>12.4.</strong> A Empresa{" "}
				<strong>NÃO se responsabiliza</strong> por indisponibilidades
				decorrentes de fatores externos à sua esfera de controle,
				incluindo, mas não se limitando a: instabilidades nos serviços
				do WhatsApp/Meta, falhas em provedores de infraestrutura cloud,
				interrupções de conectividade de rede, problemas com DNS,
				ataques cibernéticos a terceiros ou falhas nos provedores de
				pagamento.
			</p>

			{/* 13 */}
			<h2 id="limitacao-responsabilidade">
				13. Limitação de Responsabilidade
			</h2>
			<p>
				<strong>
					ESTA SEÇÃO CONTÉM LIMITAÇÕES IMPORTANTES DE RESPONSABILIDADE
					QUE SE APLICAM NA MÁXIMA EXTENSÃO PERMITIDA PELA LEGISLAÇÃO
					BRASILEIRA. LEIA COM ATENÇÃO REDOBRADA.
				</strong>
			</p>
			<p>
				<strong>13.1.</strong> A responsabilidade total e cumulativa da
				Empresa perante o Usuário, por <strong>qualquer causa</strong>{" "}
				relacionada a estes Termos, ao uso da Plataforma, a incidentes
				de segurança, a falhas técnicas ou a qualquer outro evento, será
				limitada ao{" "}
				<strong>
					MENOR valor entre: (a) o montante efetivamente pago pelo
					Usuário à Empresa nos 12 (doze) meses imediatamente
					anteriores ao evento que originou a reclamação; e (b) R$
					10.000,00 (dez mil reais)
				</strong>
				. Este limite aplica-se independentemente da natureza da
				reclamação (contratual, extracontratual, por ato ilícito ou
				qualquer outra).
			</p>
			<p>
				<strong>13.2.</strong> A Empresa{" "}
				<strong>NÃO SERÁ RESPONSÁVEL</strong>, em nenhuma hipótese e sob
				nenhum fundamento jurídico, por:
			</p>
			<ul>
				<li>
					Danos indiretos, incidentais, consequentes, especiais,
					punitivos, exemplares ou morais;
				</li>
				<li>
					Lucros cessantes, perda de receita, perda de negócios, perda
					de oportunidades comerciais, perda de goodwill ou perda de
					contratos;
				</li>
				<li>
					Perda, corrupção ou destruição de dados, exceto na medida em
					que tal perda decorra de dolo ou culpa grave comprovada e
					exclusiva da Empresa;
				</li>
				<li>
					Custos de aquisição de serviços substitutos ou alternativos;
				</li>
				<li>
					Reclamações de terceiros contra o Usuário, incluindo
					clientes, fornecedores, parceiros, autoridades reguladoras
					ou judiciais;
				</li>
				<li>
					Alterações, suspensões, degradações ou descontinuação de
					serviços por parte do WhatsApp/Meta;
				</li>
				<li>
					Banimento, restrição, suspensão ou encerramento de contas
					WhatsApp do Usuário por violação de políticas da Meta;
				</li>
				<li>
					Falhas, atrasos ou impossibilidade na entrega de mensagens,
					que são de responsabilidade exclusiva da infraestrutura do
					WhatsApp;
				</li>
				<li>
					Indisponibilidades, interrupções ou degradações decorrentes
					de ataques cibernéticos (DDoS, ransomware, etc.), desde que
					a Empresa tenha adotado medidas de segurança razoáveis e
					compatíveis com o estado da arte;
				</li>
				<li>
					Danos decorrentes de uso indevido, negligente ou fraudulento
					da Plataforma pelo Usuário ou por terceiros com acesso às
					credenciais do Usuário;
				</li>
				<li>
					Decisões comerciais, estratégicas ou operacionais tomadas
					pelo Usuário com base em informações, métricas ou dados
					fornecidos pela Plataforma;
				</li>
				<li>
					Incompatibilidade entre a Plataforma e sistemas, hardware,
					software ou configurações de terceiros utilizados pelo
					Usuário.
				</li>
			</ul>
			<p>
				<strong>13.3.</strong> A Plataforma é fornecida &quot;NO ESTADO
				EM QUE SE ENCONTRA&quot; (as is) e &quot;CONFORME
				DISPONÍVEL&quot; (as available). A Empresa{" "}
				<strong>NÃO OFERECE GARANTIAS DE QUALQUER ESPÉCIE</strong>,
				expressas, implícitas, legais ou de outra natureza, incluindo
				garantias de comercialidade, adequação a um fim específico, não
				violação, disponibilidade ininterrupta ou ausência de erros.
				Nenhuma informação ou orientação, verbal ou escrita, fornecida
				pela Empresa ou por seus representantes, constitui garantia não
				expressamente prevista nestes Termos.
			</p>
			<p>
				<strong>13.4.</strong> As limitações e exclusões de
				responsabilidade previstas nesta seção aplicam-se na{" "}
				<strong>máxima extensão permitida</strong> pela legislação
				brasileira. No que couber, reconhece-se a aplicação do Código de
				Defesa do Consumidor (Lei 8.078/1990) às relações de consumo,
				sendo as cláusulas destes Termos interpretadas de acordo com
				seus preceitos quando o Usuário for consumidor pessoa física.
			</p>
			<p>
				<strong>13.5.</strong> Caso qualquer limitação prevista nesta
				seção seja considerada inválida ou inexequível por tribunal
				competente, a responsabilidade da Empresa será limitada ao menor
				valor legalmente permitido.
			</p>

			{/* 14 */}
			<h2 id="indenizacao">14. Indenização</h2>
			<p>
				<strong>14.1.</strong> O Usuário concorda em{" "}
				<strong>indenizar, defender e manter</strong> a Empresa, seus
				sócios, quotistas, diretores, administradores, funcionários,
				colaboradores, agentes, prepostos, prestadores de serviço,
				contratados e representantes legais,{" "}
				<strong>integralmente isentos e indenes</strong> de quaisquer e
				todos os pedidos, reclamações, demandas, ações (judiciais,
				administrativas ou arbitrais), procedimentos, perdas,
				responsabilidades, danos, custos e despesas (incluindo, sem
				limitação, honorários advocatícios, custas processuais, perícias
				e demais encargos) decorrentes de ou relacionados a:
			</p>
			<ul>
				<li>
					Violação de quaisquer disposições destes Termos ou de
					qualquer política, contrato ou documento aqui referenciado;
				</li>
				<li>
					Uso da Plataforma pelo Usuário ou por quaisquer terceiros
					que acessem a Plataforma através da conta do Usuário;
				</li>
				<li>
					Violação de direitos de terceiros, incluindo, mas não se
					limitando a, direitos de propriedade intelectual, proteção
					de dados, privacidade, honra, imagem ou personalidade;
				</li>
				<li>
					Violação das políticas, termos de uso, diretrizes ou
					condições do WhatsApp, da Meta ou de qualquer plataforma de
					terceiros;
				</li>
				<li>
					Violação de qualquer legislação, regulamento, norma ou
					determinação judicial ou administrativa aplicável;
				</li>
				<li>
					Conteúdo de mensagens enviadas, recebidas, armazenadas ou
					processadas pelo Usuário através da Plataforma;
				</li>
				<li>
					Tratamento irregular de dados pessoais de terceiros
					realizado pelo Usuário através da Plataforma;
				</li>
				<li>
					Quaisquer reclamações de destinatários de mensagens enviadas
					pelo Usuário, incluindo denúncias de spam, assédio ou
					comunicação indesejada.
				</li>
			</ul>
			<p>
				<strong>14.2.</strong> A obrigação de indenização prevista nesta
				seção <strong>subsiste</strong> por prazo indeterminado após o
				encerramento da conta, da relação contratual ou da utilização da
				Plataforma, permanecendo válida e exigível enquanto persistirem
				os fatos geradores.
			</p>
			<p>
				<strong>14.3.</strong> A Empresa notificará o Usuário
				prontamente sobre qualquer reclamação ou ação que possa dar
				origem a obrigações de indenização sob esta seção, e cooperará
				razoavelmente com o Usuário na defesa de tais reclamações, às
				custas do Usuário.
			</p>

			{/* 15 */}
			<h2 id="suspensao-rescisao">15. Suspensão e Rescisão</h2>
			<p>
				<strong>15.1.</strong> O Usuário pode cancelar sua assinatura a
				qualquer momento através das configurações da Plataforma ou
				mediante solicitação ao e-mail{" "}
				<a href="mailto:suporte@zedaapi.com">suporte@zedaapi.com</a>. O
				cancelamento terá efeito ao final do período de faturamento
				vigente, permanecendo o acesso ativo e funcional até essa data,
				sem geração de novas cobranças.
			</p>
			<p>
				<strong>15.2.</strong> A Empresa pode suspender ou encerrar o
				acesso do Usuário à Plataforma, imediatamente e sem aviso
				prévio, nas seguintes hipóteses:
			</p>
			<ul>
				<li>
					Violação destes Termos, da Política de Privacidade, da
					Política de Cookies ou de quaisquer políticas do
					WhatsApp/Meta;
				</li>
				<li>
					Inadimplência por período superior a 15 (quinze) dias
					corridos após a terceira tentativa de cobrança;
				</li>
				<li>
					Utilização da Plataforma para fins ilegais, fraudulentos,
					abusivos ou que representem risco a terceiros;
				</li>
				<li>
					Determinação judicial, arbitral ou de autoridade
					administrativa competente;
				</li>
				<li>
					Comportamento que represente risco à segurança, integridade,
					disponibilidade ou reputação da Plataforma ou da Empresa;
				</li>
				<li>
					Fornecimento de informações falsas ou fraudulentas no
					cadastro ou durante a utilização da Plataforma;
				</li>
				<li>
					Uso da Plataforma em desacordo com a finalidade para a qual
					foi desenvolvida.
				</li>
			</ul>
			<p>
				<strong>15.3.</strong> Em caso de encerramento por violação dos
				Termos ou por conduta ilícita do Usuário,{" "}
				<strong>não haverá direito a reembolso</strong> de valores
				pagos, parcial ou integralmente, sem prejuízo da obrigação de
				indenização prevista na Seção 14.
			</p>
			<p>
				<strong>15.4.</strong> Após o encerramento voluntário da conta,
				a Empresa reterá os dados do Usuário por até 30 (trinta) dias
				para possibilitar eventual reativação. Após este período, os
				dados serão permanentemente anonimizados ou excluídos,
				ressalvadas as obrigações legais de retenção:
			</p>
			<ul>
				<li>
					Dados financeiros e NFS-e: 5 (cinco) anos (CTN, Art. 173;
					Lei 5.172/1966);
				</li>
				<li>
					Registros de acesso (IP, data/hora): 6 (seis) meses (Marco
					Civil da Internet, Art. 15, Lei 12.965/2014);
				</li>
				<li>
					Dados necessários para litígios em curso ou iminentes: até
					conclusão definitiva.
				</li>
			</ul>
			<p>
				<strong>15.5.</strong> As disposições relativas a Propriedade
				Intelectual (Seção 10), Limitação de Responsabilidade (Seção
				13), Indenização (Seção 14), Lei Aplicável e Foro (Seção 20) e
				demais cláusulas que, por sua natureza, devam subsistir,
				permanecerão em vigor após o encerramento da relação contratual.
			</p>

			{/* 16 */}
			<h2 id="forca-maior">16. Força Maior</h2>
			<p>
				<strong>16.1.</strong> Nenhuma das partes será responsável por
				atrasos, falhas ou impossibilidade no cumprimento de suas
				obrigações quando decorrentes de eventos de força maior ou caso
				fortuito, conforme definidos no artigo 393 do Código Civil
				Brasileiro, incluindo, mas não se limitando a:
			</p>
			<ul>
				<li>
					Desastres naturais (terremotos, enchentes, furacões,
					tsunamis, erupções vulcânicas);
				</li>
				<li>
					Epidemias, pandemias ou emergências sanitárias declaradas
					por autoridades competentes;
				</li>
				<li>
					Guerras, invasões, atos de terrorismo, insurreições,
					revoluções ou golpes de estado;
				</li>
				<li>
					Restrições governamentais, embargos, sanções, bloqueios
					comerciais ou regulações emergenciais;
				</li>
				<li>
					Falhas generalizadas de infraestrutura de telecomunicações,
					energia elétrica, Internet ou serviços de data center;
				</li>
				<li>
					Ataques cibernéticos de grande escala (DDoS massivo,
					ransomware, APT) que afetem provedores de infraestrutura
					crítica;
				</li>
				<li>
					Alterações legislativas, regulatórias ou de política pública
					que impossibilitem a prestação do Serviço;
				</li>
				<li>Greves generalizadas que afetem serviços essenciais;</li>
				<li>
					Decisões judiciais ou administrativas que determinem a
					suspensão ou bloqueio do Serviço.
				</li>
			</ul>
			<p>
				<strong>16.2.</strong> A parte afetada deverá notificar a outra
				parte em até 5 (cinco) dias úteis após tomar conhecimento do
				evento de força maior, descrevendo sua natureza, duração
				estimada e impacto nas obrigações contratuais, e envidará
				esforços razoáveis e de boa-fé para mitigar seus efeitos.
			</p>
			<p>
				<strong>16.3.</strong> Caso o evento de força maior persista por
				mais de 60 (sessenta) dias consecutivos, qualquer das partes
				poderá rescindir o contrato sem penalidades, mediante
				notificação por escrito.
			</p>

			{/* 17 */}
			<h2 id="programa-afiliados">17. Programa de Afiliados</h2>
			<p>
				<strong>17.1.</strong> A Empresa pode oferecer um Programa de
				Afiliados que permite a Usuários elegíveis indicar novos
				clientes e receber comissões sobre as contratações efetivadas. A
				participação no Programa está sujeita a termos e condições
				específicos disponibilizados na Plataforma.
			</p>
			<p>
				<strong>17.2.</strong> O afiliado compromete-se a promover a
				Plataforma de forma ética, lícita e em conformidade com estes
				Termos, abstendo-se de utilizar práticas enganosas, spam, falsas
				promessas ou qualquer método que viole a legislação brasileira
				ou as políticas do WhatsApp/Meta.
			</p>
			<p>
				<strong>17.3.</strong> As comissões, percentuais, prazos de
				pagamento e demais condições do Programa estão descritos nos
				termos específicos do Programa de Afiliados, que constituem
				parte integrante destes Termos quando aplicáveis.
			</p>
			<p>
				<strong>17.4.</strong> A Empresa reserva-se o direito de
				modificar, suspender ou encerrar o Programa de Afiliados a
				qualquer momento, mediante aviso prévio de 30 (trinta) dias, sem
				que isso gere direito a indenização por parte dos afiliados,
				ressalvadas as comissões já devidas.
			</p>
			<p>
				<strong>17.5.</strong> Fraudes no Programa de Afiliados,
				incluindo auto-referência, contas fictícias, manipulação de
				links de indicação ou qualquer conduta desonesta, resultarão no
				cancelamento imediato da participação, estorno de comissões e
				possível encerramento da conta na Plataforma.
			</p>

			{/* 18 */}
			<h2 id="confidencialidade">18. Confidencialidade</h2>
			<p>
				<strong>18.1.</strong> As partes comprometem-se a manter a
				confidencialidade de todas as informações técnicas, comerciais,
				financeiras e estratégicas a que tenham acesso em razão da
				relação contratual (&quot;Informações Confidenciais&quot;),
				incluindo, mas não se limitando a: especificações técnicas da
				API, tokens de acesso, dados de performance, termos comerciais
				personalizados e informações sobre a infraestrutura.
			</p>
			<p>
				<strong>18.2.</strong> As Informações Confidenciais não devem
				ser divulgadas a terceiros sem o consentimento prévio e por
				escrito da parte divulgadora, exceto quando exigido por lei,
				regulamento ou determinação judicial ou administrativa.
			</p>
			<p>
				<strong>18.3.</strong> A obrigação de confidencialidade subsiste
				por 2 (dois) anos após o término da relação contratual.
			</p>

			{/* 19 */}
			<h2 id="alteracoes-termos">19. Alterações dos Termos</h2>
			<p>
				<strong>19.1.</strong> A Empresa reserva-se o direito de
				alterar, atualizar ou revisar estes Termos a qualquer momento.
				Alterações substanciais serão comunicadas com no mínimo 30
				(trinta) dias de antecedência por e-mail e/ou notificação na
				Plataforma, em conformidade com o artigo 6o, inciso III, do
				Código de Defesa do Consumidor e o artigo 7o, inciso XI, do
				Marco Civil da Internet.
			</p>
			<p>
				<strong>19.2.</strong> O uso continuado da Plataforma após a
				entrada em vigor das alterações constitui aceitação tácita e
				irrevogável dos novos Termos.
			</p>
			<p>
				<strong>19.3.</strong> Caso o Usuário não concorde com as
				alterações, poderá encerrar sua conta antes da data de vigência
				dos novos Termos, sem penalidades, mediante solicitação conforme
				a Seção 15.
			</p>
			<p>
				<strong>19.4.</strong> Versões anteriores destes Termos ficarão
				disponíveis mediante solicitação ao e-mail{" "}
				<a href="mailto:contato@zedaapi.com">contato@zedaapi.com</a> ou
				poderão ser consultadas no histórico de versões disponibilizado
				pela Empresa.
			</p>

			{/* 20 */}
			<h2 id="lei-foro">20. Lei Aplicável e Foro</h2>
			<p>
				<strong>20.1.</strong> Estes Termos são regidos e interpretados
				exclusivamente pela legislação da República Federativa do
				Brasil.
			</p>
			<p>
				<strong>20.2.</strong> Fica eleito o foro da{" "}
				<strong>
					Comarca do Rio de Janeiro, Estado do Rio de Janeiro
				</strong>
				, para dirimir quaisquer controvérsias decorrentes destes Termos
				ou da utilização da Plataforma, com renúncia expressa a qualquer
				outro, por mais privilegiado que seja.
			</p>
			<p>
				<strong>20.3.</strong> Nos casos em que o Usuário for consumidor
				pessoa física, nos termos do artigo 2o do Código de Defesa do
				Consumidor (Lei 8.078/1990), prevalecerá o foro de seu
				domicílio, conforme artigo 101, inciso I, da mesma Lei, em
				atenção ao princípio da facilitação da defesa do consumidor.
			</p>
			<p>
				<strong>20.4.</strong> Antes de iniciar qualquer procedimento
				judicial, as partes comprometem-se a buscar a solução amigável
				da controvérsia por meio de negociação direta, pelo prazo mínimo
				de 30 (trinta) dias, contados da notificação da parte contrária
				sobre a existência do conflito.
			</p>

			{/* 21 */}
			<h2 id="disposicoes-gerais">21. Disposições Gerais</h2>
			<p>
				<strong>21.1. Integralidade.</strong> Estes Termos, juntamente
				com a Política de Privacidade, a Política de Cookies, a página
				LGPD - Direitos do Titular e demais políticas referenciadas,
				constituem o acordo integral entre as partes em relação ao
				objeto aqui descrito, substituindo quaisquer entendimentos,
				negociações, acordos anteriores, escritos ou verbais, sobre o
				mesmo assunto.
			</p>
			<p>
				<strong>21.2. Independência das Cláusulas.</strong> Se qualquer
				disposição destes Termos for considerada inválida, nula,
				anulável ou inexequível por tribunal ou autoridade competente,
				as demais disposições permanecerão em pleno vigor e efeito,
				devendo a disposição inválida ser substituída por outra válida
				que mais se aproxime da intenção original das partes.
			</p>
			<p>
				<strong>21.3. Cessão.</strong> O Usuário não poderá ceder,
				transferir, sub-rogar ou de qualquer forma alienar seus direitos
				e obrigações decorrentes destes Termos sem o consentimento
				prévio e por escrito da Empresa. A Empresa poderá ceder
				livremente seus direitos e obrigações a sociedades do mesmo
				grupo econômico, a sucessores em caso de fusão, aquisição,
				cisão, incorporação ou reorganização societária, ou a terceiros
				que assumam a operação da Plataforma.
			</p>
			<p>
				<strong>21.4. Renúncia.</strong> A tolerância, liberalidade ou
				não exercício, por qualquer das partes, de qualquer direito ou
				faculdade previstos nestes Termos não constituirá renúncia,
				novação ou precedente, podendo o direito ser exercido a qualquer
				tempo, na forma e nos prazos da legislação aplicável.
			</p>
			<p>
				<strong>21.5. Comunicações.</strong> Comunicações oficiais
				relativas a estes Termos serão realizadas por e-mail para os
				endereços cadastrados ou por notificação na Plataforma. As
				comunicações enviadas por e-mail considerar-se-ão recebidas na
				data do envio, salvo prova em contrário.
			</p>
			<p>
				<strong>21.6. Relação entre as Partes.</strong> Nada nestes
				Termos cria ou implica relação de parceria, joint venture,
				consórcio, representação comercial, emprego, trabalho, agência,
				mandato ou franquia entre a Empresa e o Usuário. Cada parte é e
				permanece como contratante independente.
			</p>
			<p>
				<strong>21.7. Interpretação.</strong> Em caso de dúvida na
				interpretação destes Termos, prevalecerá a interpretação mais
				restritiva em favor da Empresa no que se refere à limitação de
				responsabilidade, e a interpretação mais favorável ao consumidor
				nas hipóteses previstas no CDC, nos termos do artigo 47 da Lei
				8.078/1990.
			</p>
			<p>
				<strong>21.8. Idioma.</strong> Estes Termos foram redigidos em
				português do Brasil. Em caso de tradução para outros idiomas,
				prevalecerá a versão em português.
			</p>
			<p>
				<strong>21.9. Registro.</strong> Estes Termos poderão ser
				registrados em Cartório de Títulos e Documentos para fins de
				publicidade e oponibilidade a terceiros, a critério da Empresa.
			</p>

			<hr />

			<p>
				Em caso de dúvidas sobre estes Termos, entre em contato conosco:
			</p>
			<ul>
				<li>
					<strong>E-mail</strong>:{" "}
					<a href="mailto:contato@zedaapi.com">contato@zedaapi.com</a>
				</li>
				<li>
					<strong>Suporte</strong>:{" "}
					<a href="mailto:suporte@zedaapi.com">suporte@zedaapi.com</a>
				</li>
				<li>
					<strong>WhatsApp</strong>:{" "}
					<a
						href="https://wa.me/5521971532700"
						target="_blank"
						rel="noopener noreferrer"
					>
						+55 21 97153-2700
					</a>
				</li>
			</ul>
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
