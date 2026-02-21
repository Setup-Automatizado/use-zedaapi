"use client";

import { useState, useCallback } from "react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { PageHeader } from "@/components/shared/page-header";
import { Plus, Pencil, Loader2 } from "lucide-react";
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
import { toast } from "sonner";
import { getAdminPlans, createPlan, updatePlan } from "@/server/actions/admin";

interface Plan {
	id: string;
	name: string;
	slug: string;
	description: string | null;
	price: number;
	currency: string;
	interval: string;
	maxInstances: number;
	features: unknown;
	active: boolean;
	sortOrder: number;
}

interface PlansContentClientProps {
	initialPlans: Plan[];
}

export function PlansContentClient({ initialPlans }: PlansContentClientProps) {
	const [plans, setPlans] = useState<Plan[]>(initialPlans);
	const [dialogOpen, setDialogOpen] = useState(false);
	const [saving, setSaving] = useState(false);
	const [editPlan, setEditPlan] = useState<Plan | null>(null);
	const [form, setForm] = useState({
		name: "",
		slug: "",
		price: "",
		maxInstances: "",
	});

	const load = useCallback(async () => {
		const res = await getAdminPlans();
		if (res.success && res.data) setPlans(res.data);
	}, []);

	const openCreate = () => {
		setEditPlan(null);
		setForm({ name: "", slug: "", price: "", maxInstances: "" });
		setDialogOpen(true);
	};

	const openEdit = (plan: Plan) => {
		setEditPlan(plan);
		setForm({
			name: plan.name,
			slug: plan.slug,
			price: String(plan.price),
			maxInstances: String(plan.maxInstances),
		});
		setDialogOpen(true);
	};

	const handleSave = async () => {
		setSaving(true);
		if (editPlan) {
			const res = await updatePlan(editPlan.id, {
				name: form.name,
				price: parseFloat(form.price),
				maxInstances: parseInt(form.maxInstances),
			});
			if (res.success) {
				toast.success("Plano atualizado");
			} else {
				toast.error(res.error ?? "Erro ao atualizar plano");
			}
		} else {
			const res = await createPlan({
				name: form.name,
				slug: form.slug,
				price: parseFloat(form.price),
				maxInstances: parseInt(form.maxInstances),
				features: {},
			});
			if (res.success) {
				toast.success("Plano criado");
			} else {
				toast.error(res.error ?? "Erro ao criar plano");
			}
		}
		setSaving(false);
		setDialogOpen(false);
		load();
	};

	const handleToggleActive = async (plan: Plan) => {
		const res = await updatePlan(plan.id, { active: !plan.active });
		if (res.success) {
			toast.success(plan.active ? "Plano desativado" : "Plano ativado");
			load();
		}
	};

	return (
		<div className="space-y-6">
			<PageHeader
				title="Planos"
				description="Gerencie os planos de assinatura."
				action={
					<Button onClick={openCreate}>
						<Plus className="size-4" />
						Novo Plano
					</Button>
				}
			/>

			<div className="grid gap-4 lg:grid-cols-3">
				{plans.map((plan) => (
					<Card key={plan.id}>
						<CardHeader>
							<div className="flex items-center justify-between">
								<CardTitle>{plan.name}</CardTitle>
								<Badge
									variant={
										plan.active ? "default" : "secondary"
									}
								>
									{plan.active ? "Ativo" : "Inativo"}
								</Badge>
							</div>
							{plan.description && (
								<CardDescription>
									{plan.description}
								</CardDescription>
							)}
						</CardHeader>
						<CardContent className="space-y-4">
							<div className="text-3xl font-bold">
								R${" "}
								{plan.price.toLocaleString("pt-BR", {
									minimumFractionDigits: 2,
								})}
								<span className="text-sm font-normal text-muted-foreground">
									/
									{plan.interval === "month"
										? "mês"
										: plan.interval}
								</span>
							</div>

							<Separator />

							<dl className="space-y-2 text-sm">
								<div className="flex justify-between">
									<dt className="text-muted-foreground">
										Instâncias
									</dt>
									<dd className="font-medium">
										{plan.maxInstances}
									</dd>
								</div>
							</dl>

							<Separator />

							<div className="flex items-center justify-between">
								<span className="text-sm">Ativo</span>
								<Switch
									checked={plan.active}
									onCheckedChange={() =>
										handleToggleActive(plan)
									}
								/>
							</div>

							<Button
								variant="outline"
								size="sm"
								className="w-full"
								onClick={() => openEdit(plan)}
							>
								<Pencil className="size-4" />
								Editar
							</Button>
						</CardContent>
					</Card>
				))}
			</div>

			<Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>
							{editPlan ? "Editar Plano" : "Novo Plano"}
						</DialogTitle>
						<DialogDescription>
							{editPlan
								? "Altere as configurações do plano."
								: "Crie um novo plano de assinatura."}
						</DialogDescription>
					</DialogHeader>
					<div className="space-y-4 py-4">
						<div className="grid gap-4 sm:grid-cols-2">
							<div className="space-y-2">
								<Label>Nome</Label>
								<Input
									value={form.name}
									onChange={(e) =>
										setForm({
											...form,
											name: e.target.value,
										})
									}
									placeholder="Nome do plano"
								/>
							</div>
							<div className="space-y-2">
								<Label>Slug</Label>
								<Input
									value={form.slug}
									onChange={(e) =>
										setForm({
											...form,
											slug: e.target.value,
										})
									}
									placeholder="slug-do-plano"
									disabled={!!editPlan}
								/>
							</div>
							<div className="space-y-2">
								<Label>Preço (R$)</Label>
								<Input
									type="number"
									value={form.price}
									onChange={(e) =>
										setForm({
											...form,
											price: e.target.value,
										})
									}
									placeholder="29.00"
								/>
							</div>
							<div className="space-y-2">
								<Label>Max. Instâncias</Label>
								<Input
									type="number"
									value={form.maxInstances}
									onChange={(e) =>
										setForm({
											...form,
											maxInstances: e.target.value,
										})
									}
									placeholder="1"
								/>
							</div>
						</div>
					</div>
					<DialogFooter>
						<Button
							variant="outline"
							onClick={() => setDialogOpen(false)}
						>
							Cancelar
						</Button>
						<Button onClick={handleSave} disabled={saving}>
							{saving && (
								<Loader2 className="size-4 animate-spin" />
							)}
							{editPlan ? "Salvar" : "Criar"}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
