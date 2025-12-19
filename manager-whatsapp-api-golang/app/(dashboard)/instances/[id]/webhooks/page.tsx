"use client";

import { use } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";

import { useInstance } from "@/hooks";
import { WebhookConfigForm } from "@/components/instances";
import { PageHeader } from "@/components/shared/page-header";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";

interface WebhooksPageProps {
	params: Promise<{ id: string }>;
}

export default function WebhooksPage({ params }: WebhooksPageProps) {
	const resolvedParams = use(params);
	const router = useRouter();
	const { instance, isLoading, error } = useInstance(resolvedParams.id);

	if (error) {
		return (
			<div className="space-y-6">
				<Button
					variant="ghost"
					onClick={() => router.back()}
					className="mb-4"
				>
					<ArrowLeft className="mr-2 h-4 w-4" />
					Voltar
				</Button>
				<Alert variant="destructive">
					<AlertTitle>Erro ao carregar instancia</AlertTitle>
					<AlertDescription>
						{error.message ||
							"Nao foi possivel carregar as informacoes da instancia."}
					</AlertDescription>
				</Alert>
			</div>
		);
	}

	if (isLoading || !instance) {
		return (
			<div className="space-y-6">
				<Skeleton className="h-8 w-64" />
				<Skeleton className="h-96 w-full" />
			</div>
		);
	}

	return (
		<div className="space-y-6">
			<div className="flex items-center gap-4">
				<Button
					variant="ghost"
					onClick={() => router.back()}
					size="icon"
				>
					<ArrowLeft className="h-4 w-4" />
				</Button>
				<PageHeader
					title="Configuracao de Webhooks"
					description={`Configure os webhooks para a instancia ${instance.name}`}
				/>
			</div>

			<Card>
				<CardHeader>
					<CardTitle>URLs de Webhooks</CardTitle>
					<CardDescription>
						Configure os endpoints para receber notificacoes de
						eventos do WhatsApp. Deixe em branco para desabilitar
						webhooks especificos.
					</CardDescription>
				</CardHeader>
				<CardContent>
					<WebhookConfigForm
						instanceId={instance.id}
						instanceToken={instance.instanceToken}
						initialValues={{
							deliveryCallbackUrl:
								instance.webhooks?.deliveryCallbackUrl || "",
							receivedCallbackUrl:
								instance.webhooks?.receivedCallbackUrl || "",
							receivedAndDeliveryCallbackUrl:
								instance.webhooks
									?.receivedAndDeliveryCallbackUrl || "",
							messageStatusCallbackUrl:
								instance.webhooks?.messageStatusCallbackUrl ||
								"",
							connectedCallbackUrl:
								instance.webhooks?.connectedCallbackUrl || "",
							disconnectedCallbackUrl:
								instance.webhooks?.disconnectedCallbackUrl ||
								"",
							presenceChatCallbackUrl:
								instance.webhooks?.presenceChatCallbackUrl ||
								"",
							notifySentByMe:
								instance.webhooks?.notifySentByMe ||
								instance.notifySentByMe ||
								false,
						}}
					/>
				</CardContent>
			</Card>
		</div>
	);
}
