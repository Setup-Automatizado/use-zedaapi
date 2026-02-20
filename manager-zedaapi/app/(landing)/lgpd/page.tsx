import type { Metadata } from "next";
import { LegalLayout } from "@/components/landing/legal-layout";

export const metadata: Metadata = {
	title: "LGPD - Direitos do Titular - Zé da API",
	description:
		"Conheça seus direitos como titular de dados pessoais segundo a LGPD (Lei 13.709/2018). Saiba como exercer seus direitos na plataforma Zé da API.",
};

export default function LGPDPage() {
	return (
		<LegalLayout
			title="LGPD - Direitos do Titular"
			lastUpdated="Fevereiro de 2026"
		>
			<p>
				A <strong>Lei Geral de Proteção de Dados Pessoais</strong> (Lei
				13.709/2018 - LGPD) garante a você, titular de dados pessoais,
				uma série de direitos em relação ao tratamento de suas
				informações. A <strong>Setup Automatizado Ltda</strong>{" "}
				(&quot;Zé da API&quot;), na qualidade de controladora de dados,
				respeita integralmente esses direitos e facilita seu exercício
				conforme detalhado nesta página.
			</p>

			<hr />

			{/* 1 */}
			<h2 id="seus-direitos">1. Seus Direitos segundo a LGPD</h2>
			<p>
				<strong>1.1.</strong> A LGPD foi promulgada em 14 de agosto de
				2018 e entrou em vigor em setembro de 2020. Ela estabelece
				regras claras sobre coleta, armazenamento, tratamento e
				compartilhamento de dados pessoais, impondo padrões de proteção
				e penalidades significativas para o descumprimento.
			</p>
			<p>
				<strong>1.2.</strong> A lei se aplica a qualquer operação de
				tratamento de dados pessoais realizada por pessoa natural ou
				jurídica, de direito público ou privado, independentemente do
				meio, do país de sua sede ou do país onde estejam localizados os
				dados, desde que a operação de tratamento seja realizada no
				território nacional brasileiro.
			</p>

			{/* 2 */}
			<h2 id="lista-direitos">
				2. Lista Completa de Direitos (Arts. 17 a 22 da LGPD)
			</h2>
			<p>
				Como titular de dados pessoais tratados pela Zé da API, você tem
				direito a:
			</p>

			<h3>2.1. Confirmação de Tratamento (Art. 18, I)</h3>
			<p>
				Você tem o direito de obter a confirmação de que a Zé da API
				realiza tratamento de seus dados pessoais. Mediante solicitação,
				informaremos se possuímos dados pessoais seus em nossas bases e
				quais categorias de dados são tratadas.
			</p>

			<h3>2.2. Acesso aos Dados (Art. 18, II)</h3>
			<p>
				Você tem o direito de acessar todos os dados pessoais que
				mantemos sobre você. Forneceremos uma cópia completa de seus
				dados em formato legível, incluindo:
			</p>
			<ul>
				<li>Dados cadastrais (nome, e-mail, CPF/CNPJ);</li>
				<li>Histórico de transações;</li>
				<li>Registros de uso da Plataforma;</li>
				<li>Configurações de instâncias WhatsApp;</li>
				<li>Logs de acesso (conforme período de retenção).</li>
			</ul>

			<h3>2.3. Correção de Dados (Art. 18, III)</h3>
			<p>
				Você tem o direito de solicitar a correção de dados pessoais
				incompletos, inexatos ou desatualizados. Muitas correções podem
				ser feitas diretamente nas configurações de sua conta na
				Plataforma. Para dados que não possam ser alterados diretamente,
				envie uma solicitação ao nosso DPO.
			</p>

			<h3>2.4. Anonimização, Bloqueio ou Eliminação (Art. 18, IV)</h3>
			<p>
				Você tem o direito de solicitar a anonimização, bloqueio ou
				eliminação de dados pessoais que sejam:
			</p>
			<ul>
				<li>
					<strong>Desnecessários</strong>: dados que excedam o
					necessário para a finalidade informada;
				</li>
				<li>
					<strong>Excessivos</strong>: dados coletados além do
					estritamente necessário;
				</li>
				<li>
					<strong>Tratados em desconformidade</strong>: dados tratados
					sem base legal adequada ou em desacordo com a LGPD.
				</li>
			</ul>
			<p>
				Ressalvamos que alguns dados podem ser retidos por obrigação
				legal (como dados fiscais por 5 anos) mesmo após solicitação de
				eliminação.
			</p>

			<h3>2.5. Portabilidade (Art. 18, V)</h3>
			<p>
				Você tem o direito de solicitar a portabilidade de seus dados
				pessoais a outro fornecedor de serviço ou produto, mediante
				requisição expressa. Os dados serão fornecidos em formato
				estruturado, comumente utilizado e de leitura automatizada (JSON
				ou CSV), conforme regulamentação da ANPD.
			</p>

			<h3>
				2.6. Eliminação de Dados Tratados com Consentimento (Art. 18,
				VI)
			</h3>
			<p>
				Quando o tratamento de dados for baseado no consentimento, você
				pode solicitar a eliminação desses dados a qualquer momento.
				Isso inclui, por exemplo, dados coletados para fins de
				marketing. A eliminação será realizada em até 30 (trinta) dias,
				exceto quando houver obrigação legal de retenção.
			</p>
			<p>
				Para solicitar a exclusão completa de sua conta e dados, acesse
				nossa página de{" "}
				<a href="/exclusao-de-dados">Exclusão de Dados</a>.
			</p>

			<h3>2.7. Informação sobre Compartilhamento (Art. 18, VII)</h3>
			<p>
				Você tem o direito de ser informado sobre as entidades públicas
				e privadas com as quais compartilhamos seus dados pessoais.
				Atualmente, compartilhamos dados com:
			</p>
			<ul>
				<li>
					<strong>Stripe, Inc.</strong> (EUA) - processamento de
					pagamentos internacionais;
				</li>
				<li>
					<strong>Sicredi</strong> (Brasil) - processamento de PIX e
					boleto;
				</li>
				<li>
					Provedores de infraestrutura cloud - hospedagem e
					armazenamento;
				</li>
				<li>Provedores de e-mail - comunicações transacionais;</li>
				<li>Autoridades competentes - quando exigido por lei.</li>
			</ul>
			<p>
				Para detalhes completos, consulte a Seção 5 de nossa{" "}
				<a href="/politica-de-privacidade">Política de Privacidade</a>.
			</p>

			<h3>2.8. Informação sobre Não Consentimento (Art. 18, VIII)</h3>
			<p>
				Você tem o direito de ser informado sobre a possibilidade de não
				fornecer consentimento e sobre as consequências da negativa. No
				caso da Zé da API:
			</p>
			<ul>
				<li>
					<strong>Consentimento obrigatório</strong>: sem ele, não é
					possível criar uma conta e utilizar a Plataforma (base
					legal: execução contratual);
				</li>
				<li>
					<strong>Consentimento opcional (marketing)</strong>: a
					recusa não afeta o uso da Plataforma. Você apenas deixará de
					receber comunicações promocionais;
				</li>
				<li>
					<strong>Consentimento de cookies (não essenciais)</strong>:
					a recusa não afeta o funcionamento básico, mas pode limitar
					a personalização da experiência.
				</li>
			</ul>

			<h3>2.9. Revogação do Consentimento (Art. 18, IX)</h3>
			<p>
				Você pode revogar o consentimento previamente fornecido a
				qualquer momento, de forma gratuita e facilitada. A revogação
				não afeta a legalidade do tratamento realizado anteriormente com
				base no consentimento. Formas de revogar:
			</p>
			<ul>
				<li>
					<strong>Marketing por e-mail</strong>: clique em
					&quot;descadastrar&quot; no rodapé dos e-mails promocionais;
				</li>
				<li>
					<strong>Cookies</strong>: altere suas preferências no banner
					de cookies ou nas configurações do navegador;
				</li>
				<li>
					<strong>Outros tratamentos</strong>: envie e-mail para{" "}
					<a href="mailto:privacidade@zedaapi.com">
						privacidade@zedaapi.com
					</a>{" "}
					especificando qual consentimento deseja revogar.
				</li>
			</ul>

			{/* 3 */}
			<h2 id="como-exercer">3. Como Exercer seus Direitos</h2>
			<p>
				<strong>3.1.</strong> Para exercer qualquer dos direitos
				listados acima, você pode utilizar os seguintes canais:
			</p>

			<h3>3.1.1. E-mail do Encarregado (DPO)</h3>
			<p>
				Envie sua solicitação para{" "}
				<a href="mailto:privacidade@zedaapi.com">
					privacidade@zedaapi.com
				</a>{" "}
				com as seguintes informações:
			</p>
			<ul>
				<li>Nome completo;</li>
				<li>E-mail cadastrado na Plataforma;</li>
				<li>CPF ou CNPJ (para verificação de identidade);</li>
				<li>Descrição detalhada do direito que deseja exercer;</li>
				<li>Documentos complementares, se aplicável.</li>
			</ul>

			<h3>3.1.2. Formulário de Exclusão de Dados</h3>
			<p>
				Para solicitações específicas de exclusão de conta e dados,
				utilize nosso{" "}
				<a href="/exclusao-de-dados">formulário de exclusão de dados</a>
				, que oferece um processo simplificado e guiado.
			</p>

			<h3>3.1.3. Configurações da Conta</h3>
			<p>
				Alguns direitos podem ser exercidos diretamente pela Plataforma:
			</p>
			<ul>
				<li>
					<strong>Acesso e correção</strong>: nas configurações do
					perfil;
				</li>
				<li>
					<strong>Revogação de marketing</strong>: nas preferências de
					notificação;
				</li>
				<li>
					<strong>Cookies</strong>: no banner de preferências de
					cookies (rodapé do site).
				</li>
			</ul>

			<p>
				<strong>3.2. Verificação de Identidade.</strong> Para proteger
				seus dados, poderemos solicitar a verificação de sua identidade
				antes de atender determinadas solicitações. Essa verificação
				será proporcional ao tipo de solicitação e aos dados envolvidos.
			</p>

			{/* 4 */}
			<h2 id="prazos">4. Prazos de Resposta</h2>
			<p>
				<strong>4.1.</strong> Em conformidade com o artigo 18, parágrafo
				5o da LGPD, responderemos às solicitações dos titulares nos
				seguintes prazos:
			</p>
			<table>
				<thead>
					<tr>
						<th>Tipo de Solicitação</th>
						<th>Prazo</th>
						<th>Formato</th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td>Confirmação de tratamento</td>
						<td>Imediato ou até 15 dias úteis</td>
						<td>Declaração simplificada</td>
					</tr>
					<tr>
						<td>Acesso completo aos dados</td>
						<td>Até 15 dias úteis</td>
						<td>Declaração completa com todos os dados</td>
					</tr>
					<tr>
						<td>Correção de dados</td>
						<td>Até 15 dias úteis</td>
						<td>Confirmação da alteração</td>
					</tr>
					<tr>
						<td>Eliminação de dados</td>
						<td>Até 30 dias corridos</td>
						<td>Confirmação da exclusão</td>
					</tr>
					<tr>
						<td>Portabilidade</td>
						<td>Até 15 dias úteis</td>
						<td>Arquivo JSON ou CSV</td>
					</tr>
					<tr>
						<td>Informação sobre compartilhamento</td>
						<td>Até 15 dias úteis</td>
						<td>Lista de destinatários</td>
					</tr>
					<tr>
						<td>Revogação de consentimento</td>
						<td>Imediato</td>
						<td>Confirmação por e-mail</td>
					</tr>
				</tbody>
			</table>
			<p>
				<strong>4.2.</strong> Em casos de alta complexidade técnica ou
				volume excepcional de dados, o prazo poderá ser estendido
				mediante comunicação fundamentada ao titular, indicando o novo
				prazo estimado.
			</p>
			<p>
				<strong>4.3.</strong> As solicitações são atendidas
				gratuitamente, conforme previsto na LGPD.
			</p>

			{/* 5 */}
			<h2 id="anpd">
				5. ANPD - Autoridade Nacional de Proteção de Dados
			</h2>
			<p>
				<strong>5.1.</strong> A Autoridade Nacional de Proteção de Dados
				(ANPD) é o órgão da administração pública federal responsável
				por zelar pela proteção de dados pessoais e por fiscalizar o
				cumprimento da LGPD no Brasil.
			</p>
			<p>
				<strong>5.2.</strong> Caso o titular não esteja satisfeito com a
				resposta da Zé da API a uma solicitação de exercício de
				direitos, ou acredite que seus dados pessoais estejam sendo
				tratados em desconformidade com a LGPD, ele pode apresentar
				petição à ANPD:
			</p>
			<ul>
				<li>
					<strong>Website da ANPD</strong>:{" "}
					<a
						href="https://www.gov.br/anpd"
						target="_blank"
						rel="noopener noreferrer"
					>
						gov.br/anpd
					</a>
				</li>
				<li>
					<strong>Canal de denúncias</strong>: disponível no website
					da ANPD para registro de reclamações
				</li>
			</ul>
			<p>
				<strong>5.3.</strong> Recomendamos que, antes de acionar a ANPD,
				o titular entre em contato conosco para que possamos resolver a
				questão diretamente. Na maioria dos casos, conseguimos atender
				às solicitações de forma ágil e satisfatória.
			</p>

			<hr />

			<h2 id="contato">Contato</h2>
			<p>
				Para qualquer questão relacionada à LGPD e seus direitos como
				titular:
			</p>
			<ul>
				<li>
					<strong>Encarregado de Dados (DPO)</strong>:{" "}
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
				Atendimento de segunda a sexta-feira, das 9h às 18h (horário de
				Brasília).
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
