"use client";

import { useState } from "react";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { MoreHorizontal, Plus, Trash2, Users } from "lucide-react";
import { cn } from "@/lib/utils";

interface Member {
	id: string;
	name: string;
	email: string;
	role: string;
	joinedAt: string;
}

interface MembersClientProps {
	members: Member[];
}

const roleConfig: Record<string, { label: string; className: string }> = {
	owner: { label: "Proprietario", className: "bg-primary/10 text-primary" },
	admin: { label: "Admin", className: "bg-chart-2/10 text-chart-2" },
	member: { label: "Membro", className: "bg-muted text-muted-foreground" },
};

export function MembersClient({ members }: MembersClientProps) {
	const [inviteDialogOpen, setInviteDialogOpen] = useState(false);
	const [removeDialog, setRemoveDialog] = useState<{
		open: boolean;
		memberId: string | null;
	}>({ open: false, memberId: null });

	const columns: Column<Member>[] = [
		{
			key: "name",
			header: "Nome",
			cell: (row) => (
				<div>
					<p className="font-medium">{row.name}</p>
					<p className="text-xs text-muted-foreground">{row.email}</p>
				</div>
			),
		},
		{
			key: "role",
			header: "Funcao",
			cell: (row) => {
				const config = roleConfig[row.role] ?? {
					label: row.role,
					className: "bg-muted text-muted-foreground",
				};
				return (
					<Badge variant="secondary" className={cn(config.className)}>
						{config.label}
					</Badge>
				);
			},
		},
		{
			key: "joinedAt",
			header: "Desde",
			cell: (row) => new Date(row.joinedAt).toLocaleDateString("pt-BR"),
		},
		{
			key: "actions",
			header: "",
			cell: (row) =>
				row.role !== "owner" ? (
					<DropdownMenu>
						<DropdownMenuTrigger asChild>
							<Button variant="ghost" size="icon-sm">
								<MoreHorizontal className="size-4" />
							</Button>
						</DropdownMenuTrigger>
						<DropdownMenuContent align="end">
							<DropdownMenuItem
								className="text-destructive"
								onClick={() =>
									setRemoveDialog({
										open: true,
										memberId: row.id,
									})
								}
							>
								<Trash2 className="size-4" />
								Remover
							</DropdownMenuItem>
						</DropdownMenuContent>
					</DropdownMenu>
				) : null,
			className: "w-12",
		},
	];

	return (
		<>
			<div className="flex justify-end">
				<Button onClick={() => setInviteDialogOpen(true)}>
					<Plus className="size-4" />
					Convidar
				</Button>
			</div>

			<DataTable
				columns={columns}
				data={members}
				emptyIcon={Users}
				emptyTitle="Nenhum membro"
				emptyDescription="Convide membros para colaborar na organizacao."
				emptyActionLabel="Convidar Membro"
				onEmptyAction={() => setInviteDialogOpen(true)}
			/>

			<Dialog open={inviteDialogOpen} onOpenChange={setInviteDialogOpen}>
				<DialogContent className="sm:max-w-md">
					<DialogHeader>
						<DialogTitle>Convidar Membro</DialogTitle>
						<DialogDescription>
							Envie um convite por email para adicionar um novo
							membro.
						</DialogDescription>
					</DialogHeader>
					<div className="space-y-4 py-4">
						<div className="space-y-2">
							<Label htmlFor="invite-email">Email</Label>
							<Input
								id="invite-email"
								type="email"
								placeholder="email@exemplo.com"
							/>
						</div>
					</div>
					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => setInviteDialogOpen(false)}
						>
							Cancelar
						</Button>
						<Button onClick={() => setInviteDialogOpen(false)}>
							Enviar Convite
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<ConfirmDialog
				open={removeDialog.open}
				onOpenChange={(open) =>
					setRemoveDialog({ open, memberId: null })
				}
				title="Remover membro"
				description="Tem certeza que deseja remover este membro da organizacao? Ele perdera acesso imediatamente."
				confirmLabel="Remover"
				destructive
				onConfirm={() =>
					setRemoveDialog({ open: false, memberId: null })
				}
			/>
		</>
	);
}
