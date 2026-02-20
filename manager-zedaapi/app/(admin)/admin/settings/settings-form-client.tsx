"use client";

import { useState, useRef } from "react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Settings, Loader2 } from "lucide-react";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";
import { updateSystemSetting } from "@/server/actions/admin";

interface SystemSetting {
	id: string;
	key: string;
	value: string;
	description: string | null;
}

interface SettingsFormClientProps {
	initialSettings: SystemSetting[];
}

export function SettingsFormClient({
	initialSettings,
}: SettingsFormClientProps) {
	const [settings] = useState<SystemSetting[]>(initialSettings);
	const [saving, setSaving] = useState(false);
	const formRef = useRef<HTMLFormElement>(null);

	const handleSave = async (e: React.FormEvent) => {
		e.preventDefault();
		if (!formRef.current) return;
		setSaving(true);

		const formData = new FormData(formRef.current);
		let hasError = false;

		for (const setting of settings) {
			const newValue = formData.get(setting.key) as string;
			if (newValue !== setting.value) {
				const res = await updateSystemSetting(setting.key, newValue);
				if (!res.success) {
					toast.error(`Erro ao salvar ${setting.key}: ${res.error}`);
					hasError = true;
				}
			}
		}

		setSaving(false);
		if (!hasError) {
			toast.success("Configuracoes salvas");
		}
	};

	return (
		<form ref={formRef} onSubmit={handleSave}>
			<Card>
				<CardHeader>
					<CardTitle className="flex items-center gap-2">
						<Settings className="size-4" />
						Parametros do Sistema
					</CardTitle>
					<CardDescription>
						Alterar estes valores afeta toda a plataforma.
					</CardDescription>
				</CardHeader>
				<CardContent className="space-y-6">
					{settings.map((setting, i) => (
						<div key={setting.key}>
							{i > 0 && <Separator className="mb-6" />}
							<div className="grid gap-2 sm:grid-cols-[1fr_300px]">
								<div>
									<Label htmlFor={setting.key}>
										{setting.key}
									</Label>
									{setting.description && (
										<p className="text-xs text-muted-foreground">
											{setting.description}
										</p>
									)}
								</div>
								<Input
									id={setting.key}
									name={setting.key}
									defaultValue={setting.value}
								/>
							</div>
						</div>
					))}

					<div className="flex justify-end pt-2">
						<Button type="submit" disabled={saving}>
							{saving && (
								<Loader2 className="size-4 animate-spin" />
							)}
							Salvar configuracoes
						</Button>
					</div>
				</CardContent>
			</Card>
		</form>
	);
}
