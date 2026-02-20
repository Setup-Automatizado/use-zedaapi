"use client";

import { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { EmptyState } from "@/components/shared/empty-state";
import { Key, Plus, Copy, Eye, EyeOff, Trash2 } from "lucide-react";

interface ApiKey {
	id: string;
	name: string;
	prefix: string;
	createdAt: string;
	lastUsed: string | null;
}

interface ApiKeysClientProps {
	keys: ApiKey[];
}

export function ApiKeysClient({ keys }: ApiKeysClientProps) {
	const [createDialogOpen, setCreateDialogOpen] = useState(false);
	const [newKeyValue, setNewKeyValue] = useState<string | null>(null);
	const [showKey, setShowKey] = useState(false);
	const [revokeDialog, setRevokeDialog] = useState<{
		open: boolean;
		keyId: string | null;
	}>({ open: false, keyId: null });

	const handleCreate = () => {
		// TODO: Call server action to create key
		setCreateDialogOpen(false);
	};

	return (
		<>
			<div className="flex justify-end">
				<Button onClick={() => setCreateDialogOpen(true)}>
					<Plus className="size-4" />
					Nova Chave
				</Button>
			</div>

			{newKeyValue && (
				<Card className="border-primary/20 bg-primary/5">
					<CardContent className="py-4">
						<div className="space-y-2">
							<p className="text-sm font-medium">
								Chave criada com sucesso! Copie agora pois nao
								sera exibida novamente.
							</p>
							<div className="flex items-center gap-2">
								<Input
									value={
										showKey
											? newKeyValue
											: newKeyValue.replace(/./g, "*")
									}
									readOnly
									className="font-mono text-xs"
								/>
								<Button
									variant="outline"
									size="icon"
									onClick={() => setShowKey(!showKey)}
								>
									{showKey ? (
										<EyeOff className="size-4" />
									) : (
										<Eye className="size-4" />
									)}
								</Button>
								<Button
									variant="outline"
									size="icon"
									onClick={() =>
										navigator.clipboard.writeText(
											newKeyValue,
										)
									}
								>
									<Copy className="size-4" />
								</Button>
							</div>
							<Button
								variant="ghost"
								size="sm"
								onClick={() => {
									setNewKeyValue(null);
									setShowKey(false);
								}}
							>
								Fechar
							</Button>
						</div>
					</CardContent>
				</Card>
			)}

			{keys.length === 0 ? (
				<EmptyState
					icon={Key}
					title="Nenhuma chave API"
					description="Crie uma chave para integrar com a API do Zé da API."
					actionLabel="Criar Chave"
					onAction={() => setCreateDialogOpen(true)}
				/>
			) : (
				<div className="space-y-3">
					{keys.map((key) => (
						<Card key={key.id}>
							<CardContent className="flex items-center justify-between py-4">
								<div className="flex items-center gap-3">
									<div className="flex size-9 items-center justify-center rounded-lg bg-muted">
										<Key className="size-4 text-muted-foreground" />
									</div>
									<div>
										<p className="text-sm font-medium">
											{key.name}
										</p>
										<p className="font-mono text-xs text-muted-foreground">
											{key.prefix}
										</p>
									</div>
								</div>
								<div className="flex items-center gap-3">
									<div className="text-right text-xs text-muted-foreground">
										<p>
											Criada em{" "}
											{new Date(
												key.createdAt,
											).toLocaleDateString("pt-BR")}
										</p>
										{key.lastUsed && (
											<p>
												Usada em{" "}
												{new Date(
													key.lastUsed,
												).toLocaleDateString("pt-BR")}
											</p>
										)}
									</div>
									<Button
										variant="ghost"
										size="icon-sm"
										className="text-destructive"
										onClick={() =>
											setRevokeDialog({
												open: true,
												keyId: key.id,
											})
										}
									>
										<Trash2 className="size-4" />
									</Button>
								</div>
							</CardContent>
						</Card>
					))}
				</div>
			)}

			<Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
				<DialogContent className="sm:max-w-md">
					<DialogHeader>
						<DialogTitle>Nova Chave API</DialogTitle>
						<DialogDescription>
							Crie uma nova chave de acesso a API. A chave sera
							exibida apenas uma vez.
						</DialogDescription>
					</DialogHeader>
					<div className="space-y-4 py-4">
						<div className="space-y-2">
							<Label htmlFor="key-name">Nome da chave</Label>
							<Input
								id="key-name"
								placeholder="Ex: Producao, Desenvolvimento"
							/>
						</div>
					</div>
					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => setCreateDialogOpen(false)}
						>
							Cancelar
						</Button>
						<Button onClick={handleCreate}>Criar Chave</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<ConfirmDialog
				open={revokeDialog.open}
				onOpenChange={(open) => setRevokeDialog({ open, keyId: null })}
				title="Revogar chave API"
				description="Tem certeza que deseja revogar esta chave? Qualquer integracao que use esta chave parara de funcionar imediatamente."
				confirmLabel="Revogar"
				destructive
				onConfirm={() => setRevokeDialog({ open: false, keyId: null })}
			/>
		</>
	);
}
