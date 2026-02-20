"use client";

import { useState, useRef } from "react";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { MoreHorizontal, Users, Ban, CheckCircle, Shield } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import {
	getAdminUsers,
	banUser,
	unbanUser,
	setUserRole,
} from "@/server/actions/admin";

interface User {
	id: string;
	name: string;
	email: string;
	role: string;
	banned: boolean;
	createdAt: Date;
}

const roleConfig: Record<string, { label: string; className: string }> = {
	admin: { label: "Admin", className: "bg-primary/10 text-primary" },
	user: { label: "Usuario", className: "bg-muted text-muted-foreground" },
};

interface UsersTableClientProps {
	initialData: User[];
	initialTotal: number;
}

export function UsersTableClient({
	initialData,
	initialTotal,
}: UsersTableClientProps) {
	const [users, setUsers] = useState<User[]>(initialData);
	const [loading, setLoading] = useState(false);
	const [page, setPage] = useState(1);
	const [total, setTotal] = useState(initialTotal);
	const [search, setSearch] = useState("");
	const searchTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);
	const [banDialog, setBanDialog] = useState<{
		open: boolean;
		userId: string | null;
		action: "ban" | "unban";
	}>({ open: false, userId: null, action: "ban" });
	const [banLoading, setBanLoading] = useState(false);

	async function fetchData(pageNum: number, searchTerm?: string) {
		setLoading(true);
		const res = await getAdminUsers(pageNum, searchTerm || undefined);
		if (res.success && res.data) {
			setUsers(res.data.items);
			setTotal(res.data.total);
		}
		setLoading(false);
	}

	function handlePageChange(newPage: number) {
		setPage(newPage);
		fetchData(newPage, search);
	}

	function handleSearch(value: string) {
		setSearch(value);
		clearTimeout(searchTimeoutRef.current);
		searchTimeoutRef.current = setTimeout(() => {
			setPage(1);
			fetchData(1, value);
		}, 300);
	}

	const handleBanConfirm = async () => {
		if (!banDialog.userId) return;
		setBanLoading(true);
		const res =
			banDialog.action === "ban"
				? await banUser(banDialog.userId)
				: await unbanUser(banDialog.userId);
		setBanLoading(false);
		if (res.success) {
			toast.success(
				banDialog.action === "ban"
					? "Usuario banido"
					: "Usuario desbanido",
			);
			fetchData(page, search);
		} else {
			toast.error("Erro ao atualizar usuario");
		}
		setBanDialog({ open: false, userId: null, action: "ban" });
	};

	const handleMakeAdmin = async (userId: string) => {
		const res = await setUserRole(userId, "admin");
		if (res.success) {
			toast.success("Usuario promovido a admin");
			fetchData(page, search);
		} else {
			toast.error(res.error ?? "Erro ao atualizar funcao");
		}
	};

	const columns: Column<User>[] = [
		{
			key: "name",
			header: "Usuario",
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
			key: "status",
			header: "Status",
			cell: (row) => (
				<Badge
					variant="secondary"
					className={cn(
						row.banned
							? "bg-destructive/10 text-destructive"
							: "bg-primary/10 text-primary",
					)}
				>
					{row.banned ? "Banido" : "Ativo"}
				</Badge>
			),
		},
		{
			key: "createdAt",
			header: "Criado em",
			cell: (row) => new Date(row.createdAt).toLocaleDateString("pt-BR"),
		},
		{
			key: "actions",
			header: "",
			cell: (row) => (
				<DropdownMenu>
					<DropdownMenuTrigger asChild>
						<Button variant="ghost" size="icon-sm">
							<MoreHorizontal className="size-4" />
						</Button>
					</DropdownMenuTrigger>
					<DropdownMenuContent align="end">
						{row.role !== "admin" && (
							<DropdownMenuItem
								onClick={() => handleMakeAdmin(row.id)}
							>
								<Shield className="size-4" />
								Tornar admin
							</DropdownMenuItem>
						)}
						<DropdownMenuSeparator />
						{row.banned ? (
							<DropdownMenuItem
								onClick={() =>
									setBanDialog({
										open: true,
										userId: row.id,
										action: "unban",
									})
								}
							>
								<CheckCircle className="size-4" />
								Desbanir
							</DropdownMenuItem>
						) : (
							<DropdownMenuItem
								className="text-destructive"
								onClick={() =>
									setBanDialog({
										open: true,
										userId: row.id,
										action: "ban",
									})
								}
							>
								<Ban className="size-4" />
								Banir
							</DropdownMenuItem>
						)}
					</DropdownMenuContent>
				</DropdownMenu>
			),
			className: "w-12",
		},
	];

	return (
		<>
			<DataTable
				columns={columns}
				data={users}
				loading={loading}
				emptyIcon={Users}
				emptyTitle="Nenhum usuario"
				emptyDescription="Nenhum usuario cadastrado na plataforma."
				onSearch={handleSearch}
				searchPlaceholder="Buscar usuarios..."
				page={page}
				pageSize={20}
				totalCount={total}
				onPageChange={handlePageChange}
			/>

			<ConfirmDialog
				open={banDialog.open}
				onOpenChange={(open) =>
					setBanDialog({ open, userId: null, action: "ban" })
				}
				title={
					banDialog.action === "ban"
						? "Banir usuario"
						: "Desbanir usuario"
				}
				description={
					banDialog.action === "ban"
						? "O usuario perdera acesso a plataforma e todas as instancias serao desconectadas."
						: "O usuario voltara a ter acesso a plataforma."
				}
				confirmLabel={banDialog.action === "ban" ? "Banir" : "Desbanir"}
				destructive={banDialog.action === "ban"}
				loading={banLoading}
				onConfirm={handleBanConfirm}
			/>
		</>
	);
}
