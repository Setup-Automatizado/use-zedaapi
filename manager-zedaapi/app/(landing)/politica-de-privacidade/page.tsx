import type { Metadata } from "next";
import { LegalLayout } from "@/components/landing/legal-layout";

export const metadata: Metadata = {
	title: "Política de Privacidade - Zé da API",
	description:
		"Política de Privacidade da plataforma Zé da API Manager. Saiba como coletamos, usamos e protegemos seus dados pessoais em conformidade com a LGPD.",
};

export default function PoliticaDePrivacidadePage() {
	return (
		<LegalLayout
			title="Política de Privacidade"
			lastUpdated="Fevereiro de 2026"
		>
			<p>
				A <strong>Setup Automatizado Ltda</strong>, inscrita no CNPJ
				54.246.473/0001-00, com sede em Rio de Janeiro, RJ, Brasil, na
				qualidade de controladora de dados pessoais, apresenta esta
				Política de Privacidade (&quot;Política&quot;) que descreve como
				coletamos, utilizamos, armazenamos, compartilhamos e protegemos
				os dados pessoais dos usuários da plataforma{" "}
				<strong>Zé da API Manager</strong> (&quot;Plataforma&quot;).
			</p>
			<p>
				Esta Política foi elaborada em conformidade com a{" "}
				<strong>Lei 13.709/2018</strong> (Lei Geral de Proteção de Dados
				Pessoais - LGPD), o{" "}
				<strong>Marco Civil da Internet (Lei 12.965/2014)</strong> e
				demais normas aplicáveis do ordenamento jurídico brasileiro.
			</p>

			<hr />

			{/* 1 */}
			<h2 id="introducao">1. Introdução e Compromisso</h2>
			<p>
				<strong>1.1.</strong> Respeitamos a privacidade de nossos
				usuários e estamos comprometidos com a proteção de seus dados
				pessoais. Esta Política detalha nossas práticas de tratamento de
				dados de forma transparente, clara e em linguagem acessível,
				conforme determina o artigo 9o da LGPD.
			</p>
			<p>
				<strong>1.2.</strong> Ao utilizar a Plataforma, o Usuário
				declara ter lido e compreendido esta Política. Recomendamos a
				leitura periódica deste documento, que pode ser atualizado para
				refletir mudanças em nossas práticas ou na legislação.
			</p>

			{/* 2 */}
			<h2 id="controlador">2. Controlador dos Dados</h2>
			<p>
				<strong>2.1.</strong> O controlador dos dados pessoais tratados
				pela Plataforma é:
			</p>
			<ul>
				<li>
					<strong>Razão Social</strong>: Setup Automatizado Assessoria
					em Tecnologia LTDA
				</li>
				<li>
					<strong>CNPJ</strong>: 54.246.473/0001-00
				</li>
				<li>
					<strong>Endereço</strong>: Rio de Janeiro, RJ, Brasil
				</li>
				<li>
					<strong>Website</strong>:{" "}
					<a
						href="https://zedaapi.com"
						target="_blank"
						rel="noopener noreferrer"
					>
						zedaapi.com
					</a>
				</li>
			</ul>
			<p>
				<strong>2.2.</strong> Encarregado de Proteção de Dados (DPO):
			</p>
			<ul>
				<li>
					<strong>E-mail</strong>:{" "}
					<a href="mailto:privacidade@zedaapi.com">
						privacidade@zedaapi.com
					</a>
				</li>
				<li>
					<strong>Canal de atendimento</strong>: disponível de segunda
					a sexta-feira, das 9h às 18h (horário de Brasília), com
					prazo de resposta de até 15 (quinze) dias úteis, conforme a
					LGPD.
				</li>
			</ul>

			{/* 3 */}
			<h2 id="dados-coletados">3. Dados Pessoais Coletados</h2>
			<p>
				Coletamos os seguintes tipos de dados pessoais, organizados por
				categoria:
			</p>

			<h3>3.1. Dados Cadastrais</h3>
			<p>
				Coletados no momento do registro e manutenção da conta na
				Plataforma:
			</p>
			<ul>
				<li>Nome completo;</li>
				<li>Endereço de e-mail;</li>
				<li>CPF ou CNPJ (para emissão de nota fiscal);</li>
				<li>Número de telefone;</li>
				<li>Endereço (quando necessário para faturamento);</li>
				<li>Nome da empresa e cargo (quando aplicável).</li>
			</ul>

			<h3>3.2. Dados Financeiros</h3>
			<p>Relacionados à contratação e pagamento dos serviços:</p>
			<ul>
				<li>
					Dados de pagamento processados por <strong>Stripe</strong> e{" "}
					<strong>Sicredi</strong> (não armazenamos dados completos de
					cartão de crédito em nossos servidores);
				</li>
				<li>Histórico de transações e faturas;</li>
				<li>Plano contratado e status da assinatura;</li>
				<li>Dados fiscais para emissão de NFS-e.</li>
			</ul>

			<h3>3.3. Dados de Uso</h3>
			<p>Coletados automaticamente durante a utilização da Plataforma:</p>
			<ul>
				<li>Endereço IP;</li>
				<li>Tipo de navegador (User Agent) e sistema operacional;</li>
				<li>Páginas visitadas e tempo de permanência;</li>
				<li>Data e hora de acesso;</li>
				<li>Logs de ações realizadas na Plataforma;</li>
				<li>Informações de dispositivo e resolução de tela.</li>
			</ul>

			<h3>3.4. Dados da API e Instâncias</h3>
			<p>Relacionados ao uso técnico da Plataforma:</p>
			<ul>
				<li>
					Dados das instâncias WhatsApp configuradas (números, status,
					configurações);
				</li>
				<li>URLs de webhooks configurados;</li>
				<li>Chaves de API e tokens de acesso;</li>
				<li>
					Logs de chamadas à API (endpoints, parâmetros, respostas);
				</li>
				<li>
					Métricas de uso (mensagens enviadas/recebidas, mídias
					processadas).
				</li>
			</ul>

			<h3>3.5. Dados de Comunicação</h3>
			<p>
				Quando o Usuário entra em contato conosco por canais de suporte:
			</p>
			<ul>
				<li>Conteúdo de e-mails e mensagens de suporte;</li>
				<li>Registros de atendimento.</li>
			</ul>

			<h3>3.6. Cookies e Tecnologias de Rastreamento</h3>
			<p>
				Utilizamos cookies e tecnologias similares conforme detalhado em
				nossa <a href="/politica-de-cookies">Política de Cookies</a>.
			</p>

			{/* 4 */}
			<h2 id="finalidades">
				4. Finalidades e Bases Legais do Tratamento
			</h2>
			<p>
				Tratamos dados pessoais com base nas seguintes finalidades e
				fundamentos legais previstos no artigo 7o da LGPD:
			</p>

			<table>
				<thead>
					<tr>
						<th>Finalidade</th>
						<th>Base Legal (LGPD Art. 7o)</th>
						<th>Dados Utilizados</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>Criação e manutenção de conta</td>
						<td>Execução contratual (inciso V)</td>
						<td>Cadastrais</td>
					</tr>
					<tr>
						<td>Prestação dos serviços contratados</td>
						<td>Execução contratual (inciso V)</td>
						<td>Cadastrais, API, Uso</td>
					</tr>
					<tr>
						<td>Processamento de pagamentos</td>
						<td>Execução contratual (inciso V)</td>
						<td>Cadastrais, Financeiros</td>
					</tr>
					<tr>
						<td>Emissão de notas fiscais (NFS-e)</td>
						<td>Obrigação legal (inciso II)</td>
						<td>Cadastrais, Financeiros</td>
					</tr>
					<tr>
						<td>Cumprimento de obrigações fiscais e tributárias</td>
						<td>Obrigação legal (inciso II)</td>
						<td>Cadastrais, Financeiros</td>
					</tr>
					<tr>
						<td>Suporte técnico e atendimento ao cliente</td>
						<td>Execução contratual (inciso V)</td>
						<td>Cadastrais, Comunicação, Uso</td>
					</tr>
					<tr>
						<td>Segurança da Plataforma e prevenção a fraudes</td>
						<td>Legítimo interesse (inciso IX)</td>
						<td>Uso, IP, Logs</td>
					</tr>
					<tr>
						<td>Melhoria dos serviços e experiência do usuário</td>
						<td>Legítimo interesse (inciso IX)</td>
						<td>Uso, Navegação</td>
					</tr>
					<tr>
						<td>
							Envio de comunicações sobre o serviço
							(transacionais)
						</td>
						<td>Execução contratual (inciso V)</td>
						<td>Cadastrais</td>
					</tr>
					<tr>
						<td>Envio de comunicações de marketing</td>
						<td>Consentimento (inciso I)</td>
						<td>Cadastrais</td>
					</tr>
					<tr>
						<td>
							Cumprimento de decisões judiciais ou administrativas
						</td>
						<td>Obrigação legal (inciso II)</td>
						<td>Conforme solicitado</td>
					</tr>
					<tr>
						<td>Analytics e métricas agregadas</td>
						<td>Legítimo interesse (inciso IX)</td>
						<td>Uso (anonimizados)</td>
					</tr>
				</tbody>
			</table>

			<p>
				<strong>4.1.</strong> Quando a base legal for o consentimento, o
				Usuário poderá revogá-lo a qualquer momento, sem prejuízo da
				legalidade do tratamento realizado anteriormente. A revogação
				pode ser feita através das configurações da conta ou pelo e-mail{" "}
				<a href="mailto:privacidade@zedaapi.com">
					privacidade@zedaapi.com
				</a>
				.
			</p>
			<p>
				<strong>4.2.</strong> O legítimo interesse como base legal é
				utilizado exclusivamente quando não prevalecem os direitos e
				liberdades fundamentais do titular que exijam a proteção dos
				dados pessoais, conforme avaliação documentada pela Empresa
				(Relatório de Impacto à Proteção de Dados).
			</p>

			{/* 5 */}
			<h2 id="compartilhamento">5. Compartilhamento de Dados</h2>
			<p>
				<strong>5.1.</strong> Podemos compartilhar dados pessoais com
				terceiros nas seguintes hipóteses:
			</p>

			<h3>5.1.1. Processadores de Pagamento</h3>
			<ul>
				<li>
					<strong>Stripe, Inc.</strong> (EUA): processamento de
					pagamentos com cartão de crédito e métodos internacionais.
					Dados compartilhados: nome, e-mail, dados de pagamento.
					Política de privacidade: stripe.com/privacy;
				</li>
				<li>
					<strong>Sicredi</strong> (Brasil): processamento de
					pagamentos via PIX e boleto bancário. Dados compartilhados:
					nome, CPF/CNPJ, valor.
				</li>
			</ul>

			<h3>5.1.2. Provedores de Infraestrutura</h3>
			<ul>
				<li>
					Provedores de hospedagem cloud para armazenamento e
					processamento de dados;
				</li>
				<li>
					Provedores de CDN (Content Delivery Network) para
					distribuição de conteúdo estático;
				</li>
				<li>
					Provedores de armazenamento de objetos (S3-compatível) para
					mídias.
				</li>
			</ul>

			<h3>5.1.3. Provedores de Comunicação</h3>
			<ul>
				<li>
					Provedores de e-mail (SMTP) para envio de e-mails
					transacionais e de marketing;
				</li>
				<li>Serviços de notificação para alertas da Plataforma.</li>
			</ul>

			<h3>5.1.4. Autoridades Competentes</h3>
			<ul>
				<li>
					Quando exigido por lei, regulamento, processo judicial ou
					solicitação governamental válida;
				</li>
				<li>
					Para proteção dos direitos, propriedade ou segurança da
					Empresa, de seus usuários ou do público.
				</li>
			</ul>

			<p>
				<strong>5.2.</strong> Todos os terceiros com quem compartilhamos
				dados são contratualmente obrigados a manter a confidencialidade
				e a segurança dos dados, utilizando-os exclusivamente para as
				finalidades contratadas.
			</p>
			<p>
				<strong>5.3.</strong> A Empresa <strong>NÃO</strong> vende,
				aluga ou comercializa dados pessoais de seus usuários a
				terceiros para fins de marketing.
			</p>

			{/* 6 */}
			<h2 id="transferencia-internacional">
				6. Transferência Internacional de Dados
			</h2>
			<p>
				<strong>6.1.</strong> Alguns de nossos provedores de serviço
				estão localizados fora do Brasil, notadamente nos Estados Unidos
				(Stripe). Nesses casos, a transferência internacional de dados
				ocorre em conformidade com o Capítulo V da LGPD (artigos 33 a
				36), com base nas seguintes garantias:
			</p>
			<ul>
				<li>Cláusulas contratuais padrão de proteção de dados;</li>
				<li>
					Certificações e selos reconhecidos pela ANPD (quando
					disponíveis);
				</li>
				<li>
					Medidas técnicas e organizacionais complementares
					(criptografia em trânsito e em repouso).
				</li>
			</ul>
			<p>
				<strong>6.2.</strong> Os dados pessoais transferidos
				internacionalmente permanecem sujeitos às proteções previstas
				nesta Política e na legislação brasileira.
			</p>

			{/* 7 */}
			<h2 id="retencao">7. Retenção de Dados</h2>
			<p>
				<strong>7.1.</strong> Mantemos os dados pessoais pelo tempo
				necessário para cumprir as finalidades para as quais foram
				coletados, conforme os seguintes períodos:
			</p>
			<table>
				<thead>
					<tr>
						<th>Tipo de Dado</th>
						<th>Período de Retenção</th>
						<th>Justificativa</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>Dados cadastrais</td>
						<td>Vigência da conta + 6 meses</td>
						<td>
							Possibilitar reativação e cumprimento contratual
						</td>
					</tr>
					<tr>
						<td>Dados financeiros e NFS-e</td>
						<td>5 anos após a transação</td>
						<td>Obrigação fiscal/tributária (CTN Art. 173)</td>
					</tr>
					<tr>
						<td>Logs de acesso (IP, data/hora)</td>
						<td>6 meses</td>
						<td>Marco Civil da Internet (Art. 15)</td>
					</tr>
					<tr>
						<td>Logs de API e webhooks</td>
						<td>90 dias</td>
						<td>Suporte técnico e resolução de incidentes</td>
					</tr>
					<tr>
						<td>Dados de comunicação (suporte)</td>
						<td>2 anos</td>
						<td>Histórico de atendimento e qualidade</td>
					</tr>
					<tr>
						<td>Dados de instâncias WhatsApp</td>
						<td>Vigência da conta + 30 dias</td>
						<td>Cumprimento contratual</td>
					</tr>
					<tr>
						<td>Cookies de sessão</td>
						<td>Até o encerramento da sessão</td>
						<td>Funcionalidade essencial</td>
					</tr>
				</tbody>
			</table>
			<p>
				<strong>7.2.</strong> Ao término do período de retenção, os
				dados serão anonimizados ou permanentemente excluídos, exceto
				quando a retenção for exigida por lei.
			</p>

			{/* 8 */}
			<h2 id="direitos-titular">8. Direitos do Titular</h2>
			<p>
				<strong>8.1.</strong> Em conformidade com os artigos 17 a 22 da
				LGPD, o titular dos dados pessoais tem os seguintes direitos:
			</p>
			<ul>
				<li>Confirmação da existência de tratamento;</li>
				<li>Acesso aos dados;</li>
				<li>
					Correção de dados incompletos, inexatos ou desatualizados;
				</li>
				<li>
					Anonimização, bloqueio ou eliminação de dados desnecessários
					ou tratados em desconformidade com a LGPD;
				</li>
				<li>
					Portabilidade dos dados a outro fornecedor de serviço ou
					produto;
				</li>
				<li>
					Eliminação dos dados pessoais tratados com base no
					consentimento;
				</li>
				<li>
					Informação sobre entidades públicas e privadas com as quais
					o controlador compartilhou dados;
				</li>
				<li>
					Informação sobre a possibilidade de não consentir e sobre as
					consequências da negativa;
				</li>
				<li>Revogação do consentimento.</li>
			</ul>
			<p>
				<strong>8.2.</strong> Para exercer seus direitos, acesse a
				página <a href="/lgpd">LGPD - Direitos do Titular</a> ou entre
				em contato com nosso Encarregado de Dados pelo e-mail{" "}
				<a href="mailto:privacidade@zedaapi.com">
					privacidade@zedaapi.com
				</a>
				.
			</p>
			<p>
				<strong>8.3.</strong> Responderemos às solicitações em até{" "}
				<strong>15 (quinze) dias úteis</strong>, conforme previsto na
				legislação. Em casos de alta complexidade, o prazo poderá ser
				estendido mediante comunicação ao titular.
			</p>

			{/* 9 */}
			<h2 id="seguranca">9. Segurança dos Dados</h2>
			<p>
				<strong>9.1.</strong> Adotamos medidas técnicas e
				organizacionais adequadas para proteger os dados pessoais contra
				acesso não autorizado, destruição, perda, alteração ou qualquer
				forma de tratamento inadequado, incluindo:
			</p>
			<ul>
				<li>
					<strong>Criptografia em trânsito</strong>: todas as
					comunicações são realizadas via HTTPS/TLS;
				</li>
				<li>
					<strong>Criptografia em repouso</strong>: dados sensíveis
					armazenados com AES-256-GCM;
				</li>
				<li>
					<strong>Autenticação</strong>: suporte a autenticação de
					dois fatores (2FA);
				</li>
				<li>
					<strong>Controle de acesso</strong>: princípio do menor
					privilégio para acesso a dados;
				</li>
				<li>
					<strong>Tokens de API</strong>: chaves de autenticação com
					mínimo de 16 caracteres;
				</li>
				<li>
					<strong>Monitoramento</strong>: logs de acesso e detecção de
					atividades anômalas;
				</li>
				<li>
					<strong>Backup</strong>: rotinas automatizadas de backup com
					criptografia;
				</li>
				<li>
					<strong>Infraestrutura</strong>: servidores em ambientes
					protegidos com redundância geográfica.
				</li>
			</ul>
			<p>
				<strong>9.2.</strong> Apesar de nossos esforços, nenhum sistema
				de segurança é completamente inviolável. O Usuário deve adotar
				boas práticas de segurança, como utilizar senhas fortes, ativar
				2FA e não compartilhar credenciais de acesso.
			</p>

			{/* 10 */}
			<h2 id="incidentes">10. Incidentes de Segurança</h2>
			<p>
				<strong>10.1.</strong> Em caso de incidente de segurança que
				possa acarretar risco ou dano relevante aos titulares, a
				Empresa:
			</p>
			<ul>
				<li>
					Comunicará a Autoridade Nacional de Proteção de Dados (ANPD)
					em prazo razoável, conforme o artigo 48 da LGPD;
				</li>
				<li>
					Notificará os titulares afetados, informando a natureza dos
					dados pessoais afetados, as medidas tomadas e as
					recomendações para mitigar os possíveis efeitos;
				</li>
				<li>
					Adotará medidas imediatas para conter o incidente e
					minimizar seus efeitos.
				</li>
			</ul>
			<p>
				<strong>10.2.</strong> A Empresa mantém um plano de resposta a
				incidentes documentado e atualizado, com procedimentos para
				identificação, contenção, erradicação e recuperação.
			</p>

			{/* 11 */}
			<h2 id="menores">11. Dados de Menores</h2>
			<p>
				<strong>11.1.</strong> A Plataforma não é destinada a menores de
				18 (dezoito) anos. Não coletamos intencionalmente dados pessoais
				de crianças ou adolescentes.
			</p>
			<p>
				<strong>11.2.</strong> Caso tomemos conhecimento de que
				coletamos dados de menor de idade sem o devido consentimento dos
				pais ou responsáveis legais, tomaremos medidas imediatas para
				excluir tais dados, conforme o artigo 14 da LGPD.
			</p>

			{/* 12 */}
			<h2 id="alteracoes">12. Alterações na Política</h2>
			<p>
				<strong>12.1.</strong> Esta Política poderá ser atualizada
				periodicamente para refletir mudanças em nossas práticas de
				tratamento de dados ou alterações na legislação aplicável.
			</p>
			<p>
				<strong>12.2.</strong> Alterações substanciais serão comunicadas
				por e-mail e/ou notificação na Plataforma com antecedência
				mínima de 15 (quinze) dias antes de sua entrada em vigor.
			</p>
			<p>
				<strong>12.3.</strong> A data de última atualização será sempre
				indicada no topo deste documento.
			</p>

			{/* 13 */}
			<h2 id="contato-dpo">13. Contato do DPO</h2>
			<p>
				Para questões relacionadas à privacidade e proteção de dados
				pessoais, entre em contato com nosso Encarregado de Proteção de
				Dados:
			</p>
			<ul>
				<li>
					<strong>E-mail do DPO</strong>:{" "}
					<a href="mailto:privacidade@zedaapi.com">
						privacidade@zedaapi.com
					</a>
				</li>
				<li>
					<strong>E-mail geral</strong>:{" "}
					<a href="mailto:contato@zedaapi.com">contato@zedaapi.com</a>
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
				Caso não esteja satisfeito com nossa resposta, o titular pode
				apresentar reclamação à{" "}
				<strong>Autoridade Nacional de Proteção de Dados (ANPD)</strong>
				:
			</p>
			<ul>
				<li>
					<strong>Website</strong>:{" "}
					<a
						href="https://www.gov.br/anpd"
						target="_blank"
						rel="noopener noreferrer"
					>
						gov.br/anpd
					</a>
				</li>
			</ul>

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
