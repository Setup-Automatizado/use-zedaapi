import type { Metadata } from "next";
import { LegalLayout } from "@/components/landing/legal-layout";

export const metadata: Metadata = {
	title: "Política de Cookies - Zé da API",
	description:
		"Política de Cookies da plataforma Zé da API Manager. Entenda como utilizamos cookies e tecnologias de rastreamento.",
};

export default function PoliticaDeCookiesPage() {
	return (
		<LegalLayout
			title="Política de Cookies"
			lastUpdated="Fevereiro de 2026"
		>
			<p>
				Esta Política de Cookies (&quot;Política&quot;) descreve como a{" "}
				<strong>Setup Automatizado Ltda</strong> (&quot;Empresa&quot;,
				&quot;nós&quot;), inscrita no CNPJ 54.246.473/0001-00, utiliza
				cookies e tecnologias similares na plataforma{" "}
				<strong>Zé da API Manager</strong> (&quot;Plataforma&quot;),
				disponível em{" "}
				<a
					href="https://zedaapi.com"
					target="_blank"
					rel="noopener noreferrer"
				>
					zedaapi.com
				</a>
				.
			</p>
			<p>
				Esta Política complementa nossa{" "}
				<a href="/politica-de-privacidade">Política de Privacidade</a> e
				deve ser lida em conjunto com ela.
			</p>

			<hr />

			{/* 1 */}
			<h2 id="o-que-sao">1. O que são Cookies</h2>
			<p>
				<strong>1.1.</strong> Cookies são pequenos arquivos de texto
				armazenados no navegador do Usuário quando este acessa um
				website. Eles permitem que o website reconheça o navegador em
				visitas subsequentes, armazenando preferências e informações de
				sessão.
			</p>
			<p>
				<strong>1.2.</strong> Tecnologias similares incluem local
				storage, session storage, pixels de rastreamento (web beacons) e
				scripts de analytics. Nesta Política, todas essas tecnologias
				são referidas genericamente como &quot;Cookies&quot;.
			</p>
			<p>
				<strong>1.3.</strong> Cookies podem ser de &quot;primeira
				parte&quot; (definidos diretamente pela Plataforma) ou de
				&quot;terceira parte&quot; (definidos por serviços integrados à
				Plataforma).
			</p>

			{/* 2 */}
			<h2 id="cookies-utilizados">2. Cookies que Utilizamos</h2>
			<p>
				Organizamos os cookies utilizados na Plataforma nas seguintes
				categorias:
			</p>

			<h3>2.1. Cookies Essenciais (Estritamente Necessários)</h3>
			<p>
				Estes cookies são indispensáveis para o funcionamento da
				Plataforma. Sem eles, recursos básicos como autenticação e
				navegação não funcionam. Não requerem consentimento, pois são
				tecnicamente necessários.
			</p>
			<table>
				<thead>
					<tr>
						<th>Cookie</th>
						<th>Finalidade</th>
						<th>Duração</th>
						<th>Tipo</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>
							<code>better-auth.session_token</code>
						</td>
						<td>
							Manutenção da sessão de autenticação do Usuário na
							Plataforma
						</td>
						<td>7 dias (renovável)</td>
						<td>Primeira parte</td>
					</tr>
					<tr>
						<td>
							<code>cookie_consent</code>
						</td>
						<td>
							Registra as preferências de cookies do Usuário para
							não solicitar novamente
						</td>
						<td>365 dias</td>
						<td>Primeira parte</td>
					</tr>
					<tr>
						<td>
							<code>__csrf_token</code>
						</td>
						<td>
							Proteção contra ataques CSRF (Cross-Site Request
							Forgery)
						</td>
						<td>Sessão</td>
						<td>Primeira parte</td>
					</tr>
					<tr>
						<td>
							<code>locale</code>
						</td>
						<td>Armazena a preferência de idioma do Usuário</td>
						<td>365 dias</td>
						<td>Primeira parte</td>
					</tr>
				</tbody>
			</table>

			<h3>2.2. Cookies de Funcionalidade</h3>
			<p>
				Estes cookies permitem que a Plataforma lembre escolhas feitas
				pelo Usuário para oferecer uma experiência personalizada. Podem
				ser desabilitados, mas isso afetará a experiência de uso.
			</p>
			<table>
				<thead>
					<tr>
						<th>Cookie</th>
						<th>Finalidade</th>
						<th>Duração</th>
						<th>Tipo</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>
							<code>theme</code>
						</td>
						<td>
							Armazena a preferência de tema (claro/escuro) do
							Usuário
						</td>
						<td>365 dias</td>
						<td>Primeira parte</td>
					</tr>
					<tr>
						<td>
							<code>sidebar_state</code>
						</td>
						<td>
							Armazena o estado (expandido/recolhido) da barra
							lateral do dashboard
						</td>
						<td>365 dias</td>
						<td>Primeira parte</td>
					</tr>
				</tbody>
			</table>

			<h3>2.3. Cookies de Análise (Analytics)</h3>
			<p>
				Estes cookies nos ajudam a entender como os Usuários interagem
				com a Plataforma, permitindo melhorar a experiência de uso. Os
				dados coletados são agregados e anonimizados.
			</p>
			<table>
				<thead>
					<tr>
						<th>Cookie</th>
						<th>Finalidade</th>
						<th>Duração</th>
						<th>Tipo</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>
							<code>_analytics_id</code>
						</td>
						<td>
							Identificação anônima do Usuário para métricas de
							uso agregadas
						</td>
						<td>365 dias</td>
						<td>Primeira parte</td>
					</tr>
					<tr>
						<td>
							<code>_analytics_session</code>
						</td>
						<td>
							Rastreamento da sessão de navegação para análise de
							comportamento
						</td>
						<td>30 minutos</td>
						<td>Primeira parte</td>
					</tr>
				</tbody>
			</table>

			<h3>2.4. Cookies de Marketing</h3>
			<p>
				Estes cookies são utilizados para rastrear a eficácia de
				campanhas de marketing e oferecer conteúdo relevante ao Usuário.
				Atualmente, utilizamos os seguintes cookies de marketing:
			</p>
			<table>
				<thead>
					<tr>
						<th>Cookie</th>
						<th>Finalidade</th>
						<th>Duração</th>
						<th>Tipo</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>
							<code>utm_source</code>
						</td>
						<td>
							Rastreia a origem do tráfego de campanhas de
							marketing
						</td>
						<td>30 dias</td>
						<td>Primeira parte</td>
					</tr>
					<tr>
						<td>
							<code>utm_medium</code>
						</td>
						<td>
							Rastreia o meio da campanha de marketing (e-mail,
							social, etc.)
						</td>
						<td>30 dias</td>
						<td>Primeira parte</td>
					</tr>
					<tr>
						<td>
							<code>utm_campaign</code>
						</td>
						<td>Identifica a campanha de marketing específica</td>
						<td>30 dias</td>
						<td>Primeira parte</td>
					</tr>
				</tbody>
			</table>

			{/* 3 */}
			<h2 id="cookies-terceiros">3. Cookies de Terceiros</h2>
			<p>
				<strong>3.1.</strong> Alguns serviços integrados à Plataforma
				podem definir seus próprios cookies:
			</p>

			<h3>3.1.1. Stripe (Processamento de Pagamento)</h3>
			<p>
				Quando o Usuário acessa a página de checkout ou gerenciamento de
				assinatura, o Stripe pode definir cookies para prevenção de
				fraude e processamento seguro de pagamentos. Estes cookies são
				regidos pela{" "}
				<a
					href="https://stripe.com/privacy"
					target="_blank"
					rel="noopener noreferrer"
				>
					Política de Privacidade do Stripe
				</a>
				.
			</p>

			<p>
				<strong>3.2.</strong> Não temos controle direto sobre os cookies
				definidos por terceiros. Recomendamos consultar as respectivas
				políticas de privacidade desses serviços para mais informações.
			</p>

			{/* 4 */}
			<h2 id="gerenciar-cookies">4. Como Gerenciar Cookies</h2>
			<p>
				<strong>4.1.</strong> O Usuário pode gerenciar suas preferências
				de cookies de diferentes formas:
			</p>

			<h3>4.1.1. Banner de Consentimento</h3>
			<p>
				Na primeira visita à Plataforma, é exibido um banner de
				consentimento que permite ao Usuário aceitar ou rejeitar cookies
				não essenciais. As preferências podem ser alteradas a qualquer
				momento clicando em &quot;Alterar preferências de cookies&quot;
				no rodapé da Plataforma.
			</p>

			<h3>4.1.2. Configurações do Navegador</h3>
			<p>
				A maioria dos navegadores permite gerenciar cookies através de
				suas configurações. Instruções para os navegadores mais comuns:
			</p>
			<ul>
				<li>
					<strong>Google Chrome</strong>: Configurações &gt;
					Privacidade e Segurança &gt; Cookies e outros dados do site;
				</li>
				<li>
					<strong>Mozilla Firefox</strong>: Opções &gt; Privacidade e
					Segurança &gt; Cookies e dados de sites;
				</li>
				<li>
					<strong>Safari</strong>: Preferências &gt; Privacidade &gt;
					Cookies e dados de sites;
				</li>
				<li>
					<strong>Microsoft Edge</strong>: Configurações &gt; Cookies
					e permissões do site.
				</li>
			</ul>

			<p>
				<strong>4.2.</strong> A desativação de cookies essenciais pode
				impedir o funcionamento adequado da Plataforma, incluindo a
				impossibilidade de realizar login ou manter a sessão ativa.
			</p>
			<p>
				<strong>4.3.</strong> A desativação de cookies de análise e
				marketing não afeta o funcionamento básico da Plataforma, mas
				pode limitar a personalização da experiência.
			</p>

			{/* 5 */}
			<h2 id="base-legal">5. Base Legal para Cookies</h2>
			<p>
				<strong>5.1.</strong> Os cookies essenciais são utilizados com
				base no <strong>legítimo interesse</strong> (LGPD, Art. 7o, IX),
				por serem estritamente necessários para o funcionamento da
				Plataforma.
			</p>
			<p>
				<strong>5.2.</strong> Os cookies de funcionalidade, análise e
				marketing são utilizados com base no{" "}
				<strong>consentimento</strong> (LGPD, Art. 7o, I) do Usuário,
				obtido através do banner de consentimento exibido na primeira
				visita.
			</p>
			<p>
				<strong>5.3.</strong> O Usuário pode revogar o consentimento
				para cookies não essenciais a qualquer momento, conforme
				descrito na Seção 4 acima, sem prejuízo da legalidade do
				tratamento realizado anteriormente.
			</p>

			{/* 6 */}
			<h2 id="alteracoes-cookies">6. Alterações na Política</h2>
			<p>
				<strong>6.1.</strong> Esta Política poderá ser atualizada
				periodicamente para refletir mudanças nos cookies utilizados ou
				na legislação aplicável. Alterações substanciais serão
				comunicadas através do banner de consentimento ou notificação na
				Plataforma.
			</p>
			<p>
				<strong>6.2.</strong> A data de última atualização é indicada no
				topo deste documento.
			</p>
			<p>
				<strong>6.3.</strong> Após qualquer alteração significativa nos
				cookies utilizados, solicitaremos novamente o consentimento do
				Usuário quando aplicável.
			</p>

			<hr />

			<p>
				Em caso de dúvidas sobre esta Política de Cookies, entre em
				contato:
			</p>
			<ul>
				<li>
					<strong>E-mail</strong>:{" "}
					<a href="mailto:privacidade@zedaapi.com">
						privacidade@zedaapi.com
					</a>
				</li>
				<li>
					<strong>E-mail geral</strong>:{" "}
					<a href="mailto:contato@zedaapi.com">contato@zedaapi.com</a>
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
